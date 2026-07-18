from pathlib import Path

app_path = Path("app.go")
text = app_path.read_text()

history_line = "\thistory         storage.CampaignRepository   // Campaign run history (JSON-backed)"
emitter_line = "\temitProgress    func(models.ProgressUpdate) // Set only from the Wails lifecycle context"
if emitter_line not in text:
    if history_line not in text:
        raise SystemExit("campaign history field not found")
    text = text.replace(history_line, history_line + "\n" + emitter_line, 1)

startup_start = text.index("func (a *App) startup(ctx context.Context) {")
startup_end = text.index("\n}\n", startup_start) + len("\n}\n")
startup = """func (a *App) startup(ctx context.Context) {
\ta.ctx = ctx
\ta.emitProgress = func(update models.ProgressUpdate) {
\t\truntime.EventsEmit(ctx, \"email:progress\", update)
\t}
}
"""
text = text[:startup_start] + startup + text[startup_end:]

begin_start = text.index("func (a *App) beginRun() (context.Context, context.CancelFunc) {")
begin_line = "\tctx, cancel := context.WithCancel(a.ctx)"
if begin_line in text[begin_start : begin_start + 300]:
    text = text[:begin_start] + text[begin_start:].replace(
        begin_line,
        "\tbase := a.ctx\n\tif base == nil {\n\t\tbase = context.Background()\n\t}\n\tctx, cancel := context.WithCancel(base)",
        1,
    )
elif "base = context.Background()" not in text[begin_start : begin_start + 400]:
    raise SystemExit("beginRun context creation not found")

progress_comment = text.index("// progressSink")
helpers_marker = text.index("// ---- campaign helpers ----", progress_comment)
progress = """// progressSink forwards campaign progress through an emitter created only
// from the Wails lifecycle context. Headless execution and tests intentionally
// operate without an emitter.
func (a *App) progressSink() campaign.ProgressSink {
\treturn campaign.ProgressFunc(func(p campaign.Progress) {
\t\temit := a.emitProgress
\t\tif emit == nil {
\t\t\treturn
\t\t}
\t\temit(models.ProgressUpdate{
\t\t\tCurrent:   p.Current,
\t\t\tTotal:     p.Total,
\t\t\tStatus:    p.Status,
\t\t\tEmail:     p.Email,
\t\t\tMessage:   p.Message,
\t\t\tTimestamp: time.Now().Format(time.RFC3339),
\t\t})
\t})
}

"""
text = text[:progress_comment] + progress + text[helpers_marker:]
app_path.write_text(text)

ci_path = Path(".github/workflows/ci.yml")
ci = ci_path.read_text()
marker = "\n  apply-progress-emitter-remediation:\n"
index = ci.find(marker)
if index < 0:
    raise SystemExit("self-removal marker not found")
ci_path.write_text(ci[:index] + "\n")
