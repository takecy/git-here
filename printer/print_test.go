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

func TestPrinter_Print_ConcurrentSafe(t *testing.T) {
	// successTmpl: "{{.Repo}}\n  {{.Msg}}\n" — 2 lines per call. If a lock is
	// missing here, the two lines from one call can interleave with another
	// call's lines, breaking the strict pair structure verified below.
	is := is.New(t)

	var buf bytes.Buffer
	p := NewPrinter(&buf, &buf)

	const N = 200
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Print(Result{Repo: "/path/repo", Msg: "ok"})
		}()
	}
	wg.Wait()

	lines := strings.Split(buf.String(), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	is.Equal(len(lines), N*2)

	for i := 0; i < len(lines); i += 2 {
		is.True(strings.Contains(lines[i], "/path/repo"))
		is.True(strings.HasPrefix(lines[i+1], "  "))
		is.True(strings.Contains(lines[i+1], "ok"))
	}
}

func TestPrinter_Error_ConcurrentSafe(t *testing.T) {
	is := is.New(t)

	var errBuf bytes.Buffer
	p := NewPrinter(io.Discard, &errBuf)

	const N = 100
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Error(Result{Repo: "/path/r", Err: errors.New("boom")})
		}()
	}
	wg.Wait()

	lines := strings.Split(errBuf.String(), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	is.Equal(len(lines), N*2)
	for i := 0; i < len(lines); i += 2 {
		is.True(strings.Contains(lines[i], "/path/r"))
		is.True(strings.HasPrefix(lines[i+1], "  "))
		is.True(strings.Contains(lines[i+1], "boom"))
	}
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
	// Tiny sleep widens the race window so that, if Print and Error were
	// protected by separate mutexes, the probe would observe inflight >= 2.
	time.Sleep(20 * time.Microsecond)
	c.inflight.Add(-1)
	return len(p), nil
}

func TestPrinter_SharedMutex_AcrossWriterAndErrWriter(t *testing.T) {
	// Wires both writer and errWriter to the same probe. Print uses writer,
	// Error uses errWriter. With a single mutex covering both, no two Writes
	// can be in-flight at once → peak == 1. With separate mutexes, peak >= 2
	// would be observed with high probability.
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
				p.Print(Result{Repo: "/r", Msg: "ok"})
			} else {
				p.Error(Result{Repo: "/r", Err: errors.New("boom")})
			}
		}(i)
	}
	wg.Wait()

	is.Equal(int(probe.peak.Load()), 1)
}

func TestPrinter_MixedPrints_ConcurrentSafe(t *testing.T) {
	// Smoke test exercising every public Printer method concurrently. Catches
	// any lock omission via the race detector even when message-shape tests
	// don't apply (e.g. PrintCmd, PrintRepoErr).
	var buf bytes.Buffer
	p := NewPrinter(&buf, &buf)

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 7 {
			case 0:
				p.PrintMsg("m")
			case 1:
				p.PrintMsgErr("e")
			case 2:
				p.PrintRepoErr("re", []string{"a", "b"})
			case 3:
				p.PrintCmd("status", []string{"--short"})
			case 4:
				p.Print(Result{Repo: "/path/r", Msg: "ok"})
			case 5:
				p.Error(Result{Repo: "/path/r", Err: errors.New("boom")})
			case 6:
				p.PrintMsg("another")
			}
		}(i)
	}
	wg.Wait()
}
