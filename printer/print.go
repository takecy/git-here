package printer

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
)

// Status classifies a per-repo result for table rendering and icon selection.
type Status int

const (
	StatusSuccess Status = iota
	StatusFailed
	StatusTimeout
)

// Outcome carries every field needed to render both the streaming
// "completed" line and the final summary table row for a single repository.
type Outcome struct {
	// Repo is the absolute path. Used for failure detail dump.
	Repo string

	// Display is the shortened display name (e.g. "dxe-ai/agent").
	Display string

	// Status classifies the outcome (success / failed / timeout).
	Status Status

	// Duration is the per-repo elapsed time.
	Duration time.Duration

	// Message is the first non-empty line of git stdout (success) or
	// stderr (failure) — used in the streaming line and table cell.
	Message string

	// Err is populated only when Status != StatusSuccess.
	Err error
}

// Summary is a compact view of the per-status counts. Mirrors syncer.RunSummary
// but is duplicated here to avoid an import cycle (syncer already imports
// printer).
type Summary struct {
	Succeeded int
	Failed    int
	TimedOut  int
}

// Total reports the aggregate count.
func (s Summary) Total() int { return s.Succeeded + s.Failed + s.TimedOut }

// Printer is struct
type Printer struct {
	// mu serialises every public method's writes so that the multiple Write
	// calls per render don't interleave across goroutines.
	mu     sync.Mutex
	writer io.Writer
}

// NewPrinter is constructor
func NewPrinter(writer io.Writer) *Printer {
	if writer == nil {
		writer = os.Stdout
	}

	return &Printer{
		writer: writer,
	}
}

// PrintMsg prints message
func (p *Printer) PrintMsg(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := fmt.Fprintln(p.writer, color.GreenString(msg)); err != nil {
		log.Println(err)
	}
}

// PrintHeader emits the run-start banner: "==> Running <cmd> <opts> on N
// repositories".
func (p *Printer) PrintHeader(cmd string, options []string, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	parts := []string{cmd}
	parts = append(parts, options...)
	cmdLine := strings.Join(parts, " ")

	header := fmt.Sprintf("==> Running %s on %d repositories", cmdLine, total)
	if _, err := fmt.Fprintln(p.writer, color.GreenString(header)); err != nil {
		log.Println(err)
	}
}

// PrintRepoLine emits a single line summarising one repository's outcome.
// Format: "<icon> <Display>  <Message>  <duration>".
func (p *Printer) PrintRepoLine(o Outcome) {
	p.mu.Lock()
	defer p.mu.Unlock()

	line := fmt.Sprintf("%s %-30s %-40s %s",
		statusIconColored(o.Status),
		o.Display,
		truncate(o.Message, 40),
		formatDuration(o.Duration),
	)
	if _, err := fmt.Fprintln(p.writer, line); err != nil {
		log.Println(err)
	}
}

// PrintSummaryTable renders a "==> Summary" heading, the bordered 4-column
// table (Repository / Status / Duration / Message), and a totals line
// including elapsed wall-clock time.
func (p *Printer) PrintSummaryTable(outcomes []Outcome, summary Summary, elapsed time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := fmt.Fprintln(p.writer, color.GreenString("==> Summary")); err != nil {
		log.Println(err)
	}

	header := []string{"Repository", "Status", "Duration", "Message"}
	rows := make([][]string, 0, len(outcomes))
	for _, o := range outcomes {
		rows = append(rows, []string{
			o.Display,
			statusIcon(o.Status),
			formatDuration(o.Duration),
			truncate(o.Message, 60),
		})
	}
	renderTable(p.writer, header, rows)

	totals := fmt.Sprintf("Total: %d  Success: %d  Failed: %d  Timeout: %d  Elapsed: %s",
		summary.Total(),
		summary.Succeeded,
		summary.Failed,
		summary.TimedOut,
		formatDuration(elapsed),
	)
	if _, err := fmt.Fprintln(p.writer, totals); err != nil {
		log.Println(err)
	}
}

// writeLine is a small helper that funnels Fprintln errors through log so the
// errcheck linter is satisfied while preserving the same write semantics
// every other Printer method uses.
func writeLine(w io.Writer, s string) {
	if _, err := fmt.Fprintln(w, s); err != nil {
		log.Println(err)
	}
}

// statusIcon returns a plain (uncolored) status glyph used inside table cells
// where ANSI codes would interfere with width calculation.
func statusIcon(s Status) string {
	switch s {
	case StatusSuccess:
		return "✓"
	case StatusFailed:
		return "✗"
	case StatusTimeout:
		return "⏱"
	}
	return "?"
}

// statusIconColored returns the colored variant for use in streaming repo
// lines (outside the table).
func statusIconColored(s Status) string {
	switch s {
	case StatusSuccess:
		return color.GreenString("✓")
	case StatusFailed:
		return color.RedString("✗")
	case StatusTimeout:
		return color.YellowString("⏱")
	}
	return "?"
}

// formatDuration renders a duration in a compact, fixed style: "0.4s", "12s",
// "1m23s". Avoids the noisy "1.234567ms" shape from Duration.String.
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	if d < time.Second {
		ms := float64(d) / float64(time.Millisecond)
		return fmt.Sprintf("%.0fms", ms)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second))
	}
	return d.Round(time.Second).String()
}

// truncate caps s to maxRunes runes (visual cells, approximately) by replacing
// the tail with an ellipsis so the table stays aligned even for long messages.
func truncate(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	if maxRunes <= 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:maxRunes-1]) + "…"
}

// renderTable draws a +----+ bordered table using rune-count widths for
// alignment. Treats every rune as a single visual cell; that is correct for
// the ASCII-heavy data we render plus the three single-width status glyphs.
func renderTable(w io.Writer, header []string, rows [][]string) {
	if len(header) == 0 {
		return
	}
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if cw := utf8.RuneCountInString(cell); cw > widths[i] {
				widths[i] = cw
			}
		}
	}

	border := buildBorder(widths)
	writeLine(w, border)
	writeLine(w, buildRow(header, widths))
	writeLine(w, border)
	for _, row := range rows {
		writeLine(w, buildRow(row, widths))
	}
	writeLine(w, border)
}

func buildBorder(widths []int) string {
	var b strings.Builder
	b.WriteByte('+')
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		b.WriteByte('+')
	}
	return b.String()
}

func buildRow(cells []string, widths []int) string {
	var b strings.Builder
	b.WriteByte('|')
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteByte(' ')
		b.WriteString(cell)
		pad := w - utf8.RuneCountInString(cell)
		if pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteByte(' ')
		b.WriteByte('|')
	}
	return b.String()
}
