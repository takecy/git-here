package printer

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestPrinter_PrintMsg_ConcurrentSafe(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	p := NewPrinter(&buf)

	const N = 200
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.PrintMsg("hello")
		}()
	}
	wg.Wait()

	out := strings.TrimRight(buf.String(), "\n")
	lines := strings.Split(out, "\n")
	is.Equal(len(lines), N)
	for _, l := range lines {
		is.True(strings.Contains(l, "hello"))
	}
}

func TestPrinter_PrintRepoLine_ConcurrentSafe(t *testing.T) {
	// Each PrintRepoLine call must produce exactly one line and not
	// interleave with concurrent calls.
	is := is.New(t)

	var buf bytes.Buffer
	p := NewPrinter(&buf)

	const N = 200
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st := Status(i % 3)
			p.PrintRepoLine(Outcome{
				Repo:     "/path/repo",
				Display:  "org/repo",
				Status:   st,
				Duration: 100 * time.Millisecond,
				Message:  "msg",
			})
		}(i)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	is.Equal(len(lines), N)
	for _, l := range lines {
		is.True(strings.Contains(l, "org/repo"))
		is.True(strings.Contains(l, "msg"))
	}
}

func TestPrinter_PrintSummaryTable(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	var out bytes.Buffer
	p := NewPrinter(&out)

	outcomes := []Outcome{
		{Display: "org/a", Status: StatusSuccess, Duration: 400 * time.Millisecond, Message: "Already up to date."},
		{Display: "org/b", Status: StatusFailed, Duration: 200 * time.Millisecond, Message: "fatal: ..."},
		{Display: "org/c", Status: StatusTimeout, Duration: 1 * time.Second, Message: "timeout"},
	}
	summary := Summary{Succeeded: 1, Failed: 1, TimedOut: 1}
	p.PrintSummaryTable(outcomes, summary, 1500*time.Millisecond)

	got := out.String()

	is.True(strings.Contains(got, "==> Summary"))
	is.True(strings.Contains(got, "Repository"))
	is.True(strings.Contains(got, "Status"))
	is.True(strings.Contains(got, "Duration"))
	is.True(strings.Contains(got, "Message"))
	is.True(strings.Contains(got, "org/a"))
	is.True(strings.Contains(got, "org/b"))
	is.True(strings.Contains(got, "org/c"))
	is.True(strings.Contains(got, "+--"))
	is.True(strings.Contains(got, "Total: 3"))
	is.True(strings.Contains(got, "Success: 1"))
	is.True(strings.Contains(got, "Failed: 1"))
	is.True(strings.Contains(got, "Timeout: 1"))
}

func TestPrinter_PrintHeader(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	var out bytes.Buffer
	p := NewPrinter(&out)

	p.PrintHeader("pull", []string{"origin", "main"}, 8)

	got := out.String()
	is.True(strings.Contains(got, "==> Running"))
	is.True(strings.Contains(got, "pull origin main"))
	is.True(strings.Contains(got, "8 repositories"))
}

func TestPrinter_MixedPrints_ConcurrentSafe(t *testing.T) {
	// Smoke test exercising every public Printer method concurrently. Catches
	// any lock omission via the race detector.
	var buf bytes.Buffer
	p := NewPrinter(&buf)

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 4 {
			case 0:
				p.PrintMsg("m")
			case 1:
				p.PrintHeader("status", []string{"--short"}, 3)
			case 2:
				p.PrintRepoLine(Outcome{
					Display: "org/r",
					Status:  Status(i % 3),
					Message: "ok",
				})
			case 3:
				p.PrintSummaryTable(
					[]Outcome{{Display: "org/r", Status: StatusSuccess}},
					Summary{Succeeded: 1},
					time.Second,
				)
			}
		}(i)
	}
	wg.Wait()
}

func TestPrinter_NewPrinter_DefaultsToOsStdoutWhenNil(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	p := NewPrinter(nil)
	is.True(p.writer != nil)
}
