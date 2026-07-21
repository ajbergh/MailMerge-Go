package campaign

import (
	"os"
	"path/filepath"

	"MailMergeApp/backend/email"
	"MailMergeApp/backend/htmlutil"
	"MailMergeApp/backend/mergefield"
	"MailMergeApp/backend/models"
)

// Renderer turns a Campaign + Contact into a RenderedMessage. It is the single
// rendering path shared by preview, test send, and bulk send.
type Renderer struct {
	mf              *mergefield.Renderer
	stat            func(string) (os.FileInfo, error)
	attachmentCache map[string]ResolvedAttachment
}

// NewRenderer builds a renderer for a campaign's schema.
func NewRenderer(c Campaign) *Renderer {
	return &Renderer{
		mf:              mergefield.NewRenderer(c.Schema()),
		stat:            os.Stat,
		attachmentCache: make(map[string]ResolvedAttachment),
	}
}

// withStat overrides the filesystem stat function (used in tests). Replacing the
// stat source also resets cached metadata so tests and callers never observe
// entries produced by a previous stat implementation.
func (r *Renderer) withStat(fn func(string) (os.FileInfo, error)) *Renderer {
	r.stat = fn
	r.attachmentCache = make(map[string]ResolvedAttachment)
	return r
}

// Render produces the message for a single contact.
func (r *Renderer) Render(c Campaign, contact models.Contact) RenderedMessage {
	if c.OverrideEmail && c.ToOverride != "" {
		contact.Email = c.ToOverride
	}

	msg := RenderedMessage{IsHTML: c.IsHTML, SaveAsDraft: c.DraftOnly}
	var diags []mergefield.Diagnostic

	subject, sd := r.mf.Render(c.SubjectTemplate, contact, false)
	msg.Subject = subject
	diags = append(diags, sd...)

	htmlBody, hd := r.mf.Render(c.BodyTemplate, contact, true)
	textBody, _ := r.mf.Render(c.BodyTemplate, contact, false)
	msg.HTMLBody = htmlutil.Sanitize(htmlutil.NormalizeForEmail(htmlBody))
	msg.TextBody = textBody
	if c.IsHTML {
		diags = append(diags, hd...)
	}

	toAddr := c.ToOverride
	if toAddr == "" {
		toAddr = contact.Email
	}
	msg.To = renderAddressList(r.mf, toAddr, contact)

	if c.CCTemplate != "" {
		cc, _ := r.mf.Render(c.CCTemplate, contact, false)
		msg.CC = renderAddressList(r.mf, cc, contact)
	}
	if c.BCCTemplate != "" {
		bcc, _ := r.mf.Render(c.BCCTemplate, contact, false)
		msg.BCC = renderAddressList(r.mf, bcc, contact)
	}

	for _, a := range c.Attachments {
		path, _ := r.mf.Render(a, contact, false)
		msg.Attachments = append(msg.Attachments, r.resolveAttachment(path))
	}

	msg.Diagnostics = diags
	return msg
}

func renderAddressList(mf *mergefield.Renderer, rendered string, _ models.Contact) []string {
	valid, invalid := email.ParseList(rendered)
	out := make([]string, 0, len(valid)+len(invalid))
	for _, a := range valid {
		out = append(out, a.String())
	}
	out = append(out, invalid...)
	return out
}

func (r *Renderer) resolveAttachment(path string) ResolvedAttachment {
	if cached, ok := r.attachmentCache[path]; ok {
		return cached
	}

	ra := ResolvedAttachment{Path: path, Name: filepath.Base(path)}
	if path == "" {
		ra.Error = "empty attachment path"
		r.attachmentCache[path] = ra
		return ra
	}
	info, err := r.stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			ra.Error = "file not found"
		} else {
			ra.Error = err.Error()
		}
		r.attachmentCache[path] = ra
		return ra
	}
	ra.Exists = true
	ra.IsDir = info.IsDir()
	ra.Size = info.Size()
	if ra.IsDir {
		ra.Error = "attachment path is a directory"
	}
	r.attachmentCache[path] = ra
	return ra
}
