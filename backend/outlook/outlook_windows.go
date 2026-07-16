//go:build windows

/*
Package outlook provides the classic-Outlook COM implementation of
campaign.EmailSender. All COM access is confined to a single dedicated OS thread
(a locked STA worker), so the rest of the application — including the campaign
runner — never touches go-ole and remains testable without Outlook installed.

Threading model:
  - One goroutine calls runtime.LockOSThread and initializes COM once.
  - All COM work is submitted as commands over a channel and executed serially on
    that thread; the Outlook.Application object is created lazily and reused.
  - CoUninitialize is called on shutdown only when this code successfully
    acquired a COM initialization reference (S_OK or S_FALSE), never after
    RPC_E_CHANGED_MODE.
*/
package outlook

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"MailMergeApp/backend/campaign"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// COM HRESULT codes we handle explicitly.
const (
	sOK             = 0
	sFALSE          = 1
	rpcEChangedMode = 0x80010106
)

// command is a unit of work executed on the STA worker thread. fn receives the
// live, reused Outlook.Application COM object.
type command struct {
	fn    func(app *ole.IDispatch) error
	reply chan error
}

// Sender is a classic-Outlook COM EmailSender backed by an STA worker.
type Sender struct {
	startOnce sync.Once
	stopOnce  sync.Once
	cmds      chan command
	done      chan struct{}

	// startErr is set once during startup and read after ready closes.
	ready    chan struct{}
	startErr error

	mu     sync.Mutex
	status campaign.SenderStatus
}

// New returns a classic-Outlook sender. The STA worker starts lazily on first use.
func New() *Sender {
	return &Sender{
		cmds:  make(chan command),
		done:  make(chan struct{}),
		ready: make(chan struct{}),
	}
}

// Close shuts down the STA worker and releases COM.
func (s *Sender) Close() {
	s.stopOnce.Do(func() { close(s.done) })
}

// start launches the STA worker exactly once.
func (s *Sender) start() {
	s.startOnce.Do(func() {
		go s.loop()
	})
	<-s.ready
}

// loop runs on a single locked OS thread for the lifetime of the sender.
func (s *Sender) loop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ownsCOM, err := initCOM()
	s.startErr = err
	if ownsCOM {
		defer ole.CoUninitialize()
	}
	close(s.ready)
	if err != nil {
		// Drain commands with the startup error until closed.
		for {
			select {
			case <-s.done:
				return
			case c := <-s.cmds:
				c.reply <- err
			}
		}
	}

	var app *ole.IDispatch
	defer func() {
		if app != nil {
			app.Release()
		}
	}()

	for {
		select {
		case <-s.done:
			return
		case c := <-s.cmds:
			if app == nil {
				a, aerr := createOutlookApp()
				if aerr != nil {
					c.reply <- aerr
					continue
				}
				app = a
			}
			c.reply <- c.fn(app)
		}
	}
}

// initCOM initializes COM on the current thread and reports whether this call
// acquired an initialization reference that must be released with CoUninitialize.
func initCOM() (ownsCOM bool, err error) {
	e := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if e == nil {
		return true, nil // S_OK
	}
	oleErr, ok := e.(*ole.OleError)
	if !ok {
		return false, fmt.Errorf("failed to initialize COM: %w", e)
	}
	switch oleErr.Code() {
	case sOK, sFALSE:
		return true, nil
	case rpcEChangedMode:
		// COM already initialized on this thread with a different mode. We did
		// NOT acquire a matching reference, so must not CoUninitialize.
		return false, nil
	default:
		return false, fmt.Errorf("failed to initialize COM (hresult 0x%x): %w", oleErr.Code(), e)
	}
}

func createOutlookApp() (*ole.IDispatch, error) {
	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return nil, fmt.Errorf("outlook is not installed or not accessible via COM: %w", err)
	}
	app, err := unknown.QueryInterface(ole.IID_IDispatch)
	unknown.Release()
	if err != nil {
		return nil, fmt.Errorf("failed to get Outlook interface: %w", err)
	}
	return app, nil
}

// submit runs fn on the STA worker and returns its error (or a context error).
func (s *Sender) submit(ctx context.Context, fn func(app *ole.IDispatch) error) error {
	s.start()
	if s.startErr != nil {
		return s.startErr
	}
	reply := make(chan error, 1)
	c := command{fn: fn, reply: reply}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.cmds <- c:
	case <-s.done:
		return fmt.Errorf("sender is closed")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-reply:
		return e
	}
}

// Capabilities reports classic-Outlook COM capabilities.
func (s *Sender) Capabilities(context.Context) campaign.SenderCapabilities {
	return campaign.SenderCapabilities{
		SupportsHTML:             true,
		SupportsAttachments:      true,
		SupportsMultipleAccounts: true,
		SupportsSharedMailbox:    true,
		SupportsDraftOnly:        true,
		SupportsScheduling:       false,
	}
}

// Preflight probes whether classic Outlook is reachable via COM and returns an
// actionable status. It does not falsely claim support for new Outlook.
func (s *Sender) Preflight(ctx context.Context) campaign.SenderStatus {
	err := s.submit(ctx, func(app *ole.IDispatch) error {
		if app == nil {
			return fmt.Errorf("outlook application unavailable")
		}
		return nil
	})
	if err == nil {
		return campaign.SenderStatus{Available: true, State: campaign.StateClassicOutlook,
			Message: "Classic Outlook is available via COM automation."}
	}
	return campaign.SenderStatus{
		Available: false,
		State:     campaign.StateUnavailable,
		Message: "Could not reach Microsoft Outlook via COM automation. Ensure classic " +
			"Outlook (not New Outlook, which does not expose COM) is installed and a " +
			"mail account is configured. Details: " + err.Error(),
	}
}

// Send submits one rendered message through Outlook on the STA worker.
func (s *Sender) Send(ctx context.Context, msg campaign.RenderedMessage) campaign.SendReceipt {
	err := s.submit(ctx, func(app *ole.IDispatch) error {
		return sendMessage(app, msg)
	})
	if err != nil {
		// Treat inability to reach Outlook as fatal so the run stops cleanly.
		fatal := s.startErr != nil
		return campaign.SendReceipt{Submitted: false, Fatal: fatal, Err: err}
	}
	return campaign.SendReceipt{Submitted: true, Info: "submitted to Outlook"}
}

// sendMessage builds and sends a MailItem. Runs on the STA thread.
func sendMessage(app *ole.IDispatch, msg campaign.RenderedMessage) error {
	itemVar, err := oleutil.CallMethod(app, "CreateItem", 0) // olMailItem = 0
	if err != nil {
		return fmt.Errorf("failed to create mail item: %w", err)
	}
	item := itemVar.ToIDispatch()
	defer item.Release()

	if to := join(msg.To); to != "" {
		if _, err := oleutil.PutProperty(item, "To", to); err != nil {
			return fmt.Errorf("failed to set To: %w", err)
		}
	}
	if cc := join(msg.CC); cc != "" {
		if _, err := oleutil.PutProperty(item, "CC", cc); err != nil {
			return fmt.Errorf("failed to set CC: %w", err)
		}
	}
	if bcc := join(msg.BCC); bcc != "" {
		if _, err := oleutil.PutProperty(item, "BCC", bcc); err != nil {
			return fmt.Errorf("failed to set BCC: %w", err)
		}
	}
	if _, err := oleutil.PutProperty(item, "Subject", msg.Subject); err != nil {
		return fmt.Errorf("failed to set Subject: %w", err)
	}

	if msg.IsHTML {
		if _, err := oleutil.PutProperty(item, "HTMLBody", msg.HTMLBody); err != nil {
			return fmt.Errorf("failed to set HTMLBody: %w", err)
		}
	} else {
		if _, err := oleutil.PutProperty(item, "Body", msg.TextBody); err != nil {
			return fmt.Errorf("failed to set Body: %w", err)
		}
	}

	for _, a := range msg.Attachments {
		if !a.Exists || a.IsDir {
			return fmt.Errorf("invalid attachment: %s", a.Path)
		}
		attsVar, err := oleutil.GetProperty(item, "Attachments")
		if err != nil {
			return fmt.Errorf("failed to get attachments collection: %w", err)
		}
		atts := attsVar.ToIDispatch()
		_, err = oleutil.CallMethod(atts, "Add", a.Path)
		atts.Release()
		if err != nil {
			return fmt.Errorf("failed to add attachment %s: %w", a.Path, err)
		}
	}

	if _, err := oleutil.CallMethod(item, "Send"); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func join(addrs []string) string {
	out := ""
	for i, a := range addrs {
		if i > 0 {
			out += "; "
		}
		out += a
	}
	return out
}
