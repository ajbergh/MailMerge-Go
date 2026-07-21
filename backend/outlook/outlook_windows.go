//go:build windows

/*
Package outlook provides the classic-Outlook COM implementation of
campaign.EmailSender. All COM access is confined to a single dedicated OS thread
(a locked STA worker), so the rest of the application — including the campaign
runner — never touches go-ole and remains testable without Outlook installed.
*/
package outlook

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"MailMergeApp/backend/campaign"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	sOK             = 0
	sFALSE          = 1
	rpcEChangedMode = 0x80010106
)

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
	stopped   chan struct{}

	ready    chan struct{}
	startErr error

	mu     sync.Mutex
	status campaign.SenderStatus
}

// New returns a classic-Outlook sender. The STA worker starts lazily on first use.
func New() *Sender {
	return &Sender{
		cmds:    make(chan command),
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
		ready:   make(chan struct{}),
	}
}

// Close shuts down the STA worker, waits for any in-flight COM command to finish,
// and releases COM-owned resources before returning.
func (s *Sender) Close() {
	s.start()
	s.stopOnce.Do(func() { close(s.done) })
	<-s.stopped
}

func (s *Sender) start() {
	s.startOnce.Do(func() {
		go s.loop()
	})
	<-s.ready
}

func (s *Sender) loop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(s.stopped)

	ownsCOM, err := initCOM()
	s.startErr = err
	if ownsCOM {
		defer ole.CoUninitialize()
	}
	close(s.ready)
	if err != nil {
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

// initCOM initializes COM on the locked worker thread. RPC_E_CHANGED_MODE is a
// startup failure because this worker promises an STA; continuing in an
// incompatible apartment would make Outlook automation unsafe.
func initCOM() (ownsCOM bool, err error) {
	e := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if e == nil {
		return true, nil
	}
	oleErr, ok := e.(*ole.OleError)
	if !ok {
		return false, fmt.Errorf("failed to initialize COM: %w", e)
	}
	switch oleErr.Code() {
	case sOK, sFALSE:
		return true, nil
	case rpcEChangedMode:
		return false, fmt.Errorf("failed to initialize Outlook STA worker: COM apartment mode is incompatible (RPC_E_CHANGED_MODE): %w", e)
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
	case <-s.done:
		return fmt.Errorf("sender is closed")
	case s.cmds <- c:
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-reply:
		return e
	case <-s.stopped:
		return fmt.Errorf("sender closed before COM command completed")
	}
}

// Capabilities reports only behavior implemented by this sender. Account and
// shared-mailbox selection are intentionally false until explicit selection is
// implemented and tested.
func (s *Sender) Capabilities(context.Context) campaign.SenderCapabilities {
	return campaign.SenderCapabilities{
		SupportsHTML:             true,
		SupportsAttachments:      true,
		SupportsMultipleAccounts: false,
		SupportsSharedMailbox:    false,
		SupportsDraftOnly:        true,
		SupportsScheduling:       false,
	}
}

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
	state := campaign.StateUnavailable
	if s.startErr != nil {
		state = campaign.StateCOMInaccessible
	}
	return campaign.SenderStatus{
		Available: false,
		State:     state,
		Message: "Could not reach Microsoft Outlook via COM automation. Ensure classic " +
			"Outlook (not New Outlook, which does not expose COM) is installed and a " +
			"mail account is configured. Details: " + err.Error(),
	}
}

func (s *Sender) Send(ctx context.Context, msg campaign.RenderedMessage) campaign.SendReceipt {
	err := s.submit(ctx, func(app *ole.IDispatch) error {
		return sendMessage(app, msg)
	})
	if err != nil {
		kind := classifySendError(err, s.startErr)
		return campaign.SendReceipt{
			Submitted: false,
			Kind:      kind,
			Fatal:     kind == campaign.SendErrorFatal,
			Err:       err,
		}
	}
	info := "submitted to Outlook"
	if msg.SaveAsDraft {
		info = "saved to Outlook Drafts"
	}
	return campaign.SendReceipt{Submitted: true, Info: info}
}

func classifySendError(err error, startErr error) campaign.SendErrorKind {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return campaign.SendErrorCancelled
	}
	if startErr != nil {
		return campaign.SendErrorFatal
	}
	msg := strings.ToLower(err.Error())
	fatalMarkers := []string{
		"sender is closed",
		"sender closed before com command completed",
		"outlook is not installed or not accessible via com",
		"failed to create mail item",
		"failed to get outlook interface",
	}
	for _, marker := range fatalMarkers {
		if strings.Contains(msg, marker) {
			return campaign.SendErrorFatal
		}
	}
	return campaign.SendErrorTransient
}

// sendMessage builds and either sends or saves a MailItem. Runs on the STA thread.
func sendMessage(app *ole.IDispatch, msg campaign.RenderedMessage) error {
	itemVar, err := oleutil.CallMethod(app, "CreateItem", 0)
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

	method := "Send"
	if msg.SaveAsDraft {
		method = "Save"
	}
	if _, err := oleutil.CallMethod(item, method); err != nil {
		return fmt.Errorf("failed to %s email: %w", strings.ToLower(method), err)
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
