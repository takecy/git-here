package printer

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestPrinter_PrintMsg_ConcurrentSafe(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	p := NewPrinter(&buf, &buf)

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
	p := NewPrinter(&buf, &buf)

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

	var out, errBuf bytes.Buffer
	p := NewPrinter(&out, &errBuf)

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

	is.Equal(errBuf.Len(), 0)
}

func TestPrinter_PrintFailureDetails_HasFullStderr(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	var out, errBuf bytes.Buffer
	p := NewPrinter(&out, &errBuf)

	outcomes := []Outcome{
		{Repo: "/abs/org/ok", Display: "org/ok", Status: StatusSuccess},
		{
			Repo:    "/abs/org/bad",
			Display: "org/bad",
			Status:  StatusFailed,
			Stderr:  "fatal: not a git repository\nadditional context\n",
			Err:     errors.New("exit status 128"),
		},
	}
	p.PrintFailureDetails(outcomes)

	got := errBuf.String()
	is.True(strings.Contains(got, "/abs/org/bad"))
	is.True(strings.Contains(got, "fatal: not a git repository"))
	is.True(strings.Contains(got, "additional context"))
	is.True(strings.Contains(got, "exit status 128"))
	is.True(!strings.Contains(got, "/abs/org/ok"))

	is.Equal(out.Len(), 0)
}

func TestPrinter_PrintHeader(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	var out, errBuf bytes.Buffer
	p := NewPrinter(&out, &errBuf)

	p.PrintHeader("pull", []string{"origin", "main"}, 8)

	got := out.String()
	is.True(strings.Contains(got, "==> Running"))
	is.True(strings.Contains(got, "pull origin main"))
	is.True(strings.Contains(got, "8 repositories"))
	is.Equal(errBuf.Len(), 0)
}

// concurrencyProbe is a thin io.Writer that records the maximum number of
// concurrent in-flight Write calls. Used to verify that a single mutex covers
// BOTH writer and errWriter in Printer.
type concurrencyProbe struct {
	inflight atomic.Int32
	peak     atomic.Int32
}

func (c *concurrencyProbe) Write(p []byte) (int, error) {
	n := c.inflight.Add(1)
	for {
		cur := c.peak.Load()
		if n <= cur {
			break
		}
		if c.peak.CompareAndSwap(cur, n) {
			break
		}
	}
	// Tiny sleep widens the race window so that, if PrintRepoLine and
	// PrintFailureDetails were protected by separate mutexes, the probe
	// would observe inflight >= 2.
	time.Sleep(20 * time.Microsecond)
	c.inflight.Add(-1)
	return len(p), nil
}

func TestPrinter_SharedMutex_AcrossWriterAndErrWriter(t *testing.T) {
	// Wires both writer and errWriter to the same probe. PrintRepoLine uses
	// writer, PrintFailureDetails uses errWriter. With a single mutex
	// covering both, no two Writes can be in-flight at once → peak == 1.
	is := is.New(t)

	probe := &concurrencyProbe{}
	p := NewPrinter(probe, probe)

	const N = 200
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				p.PrintRepoLine(Outcome{
					Display: "org/r",
					Status:  StatusSuccess,
					Message: "ok",
				})
			} else {
				p.PrintFailureDetails([]Outcome{{
					Repo:    "/abs/org/r",
					Display: "org/r",
					Status:  StatusFailed,
					Stderr:  "boom",
					Err:     errors.New("boom"),
				}})
			}
		}(i)
	}
	wg.Wait()

	is.Equal(int(probe.peak.Load()), 1)
}

func TestPrinter_MixedPrints_ConcurrentSafe(t *testing.T) {
	// Smoke test exercising every public Printer method concurrently. Catches
	// any lock omission via the race detector.
	var buf bytes.Buffer
	p := NewPrinter(&buf, &buf)

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 5 {
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
			case 4:
				p.PrintFailureDetails([]Outcome{{
					Repo:    "/abs/org/r",
					Display: "org/r",
					Status:  StatusFailed,
					Stderr:  "boom",
					Err:     errors.New("boom"),
				}})
			}
		}(i)
	}
	wg.Wait()
}

func TestPrinter_NewPrinter_DefaultsToOsStdoutWhenNil(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	p := NewPrinter(nil, io.Discard)
	is.True(p.writer != nil)
	is.True(p.errWriter != nil)
}
