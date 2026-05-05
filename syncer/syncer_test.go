package syncer

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/takecy/git-here/printer"
)

type execFn func(ctx context.Context, command, dir string, args ...string) (string, string, error)

type fakeExecutor struct{ fn execFn }

func (f *fakeExecutor) Git(ctx context.Context, command, dir string, args ...string) (string, string, error) {
	return f.fn(ctx, command, dir, args...)
}

func newSyncWithFake(fn execFn, conNum int) *Sync {
	return &Sync{
		Command: "status",
		ConNum:  conNum,
		Gitter:  &fakeExecutor{fn: fn},
		Writer:  printer.NewPrinter(io.Discard),
	}
}

func TestSync_Execute(t *testing.T) {
	t.Run("all success", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := newSyncWithFake(func(_ context.Context, _, _ string, _ ...string) (string, string, error) {
			return "ok", "", nil
		}, 4)

		stats := s.execute(context.Background(), []string{"a", "b", "c"}, time.Second)
		is.Equal(len(stats.succeeded), 3)
		is.Equal(len(stats.failed), 0)
		is.Equal(len(stats.timedOut), 0)
		is.Equal(len(stats.outcomes), 3)
	})

	t.Run("partial failure", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := newSyncWithFake(func(_ context.Context, _, dir string, _ ...string) (string, string, error) {
			if filepath.Base(dir) == "b" {
				return "", "boom", errors.New("fail")
			}
			return "ok", "", nil
		}, 4)

		stats := s.execute(context.Background(), []string{"a", "b", "c"}, time.Second)
		is.Equal(len(stats.succeeded), 2)
		is.Equal(len(stats.failed), 1)
		is.Equal(len(stats.outcomes), 3)
		is.Equal(filepath.Base(stats.failed[0]), "b")
	})

	t.Run("timeout is classified separately", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := newSyncWithFake(func(ctx context.Context, _, _ string, _ ...string) (string, string, error) {
			<-ctx.Done()
			return "", "", ctx.Err()
		}, 1)

		stats := s.execute(context.Background(), []string{"a"}, 50*time.Millisecond)
		is.Equal(len(stats.timedOut), 1)
		is.Equal(len(stats.failed), 0)
		is.Equal(len(stats.succeeded), 0)
		is.Equal(len(stats.outcomes), 1)
	})

	t.Run("respects ConNum throttle", func(t *testing.T) {
		// Time-sensitive: don't run in parallel with the rest of the suite.
		is := is.New(t)

		var inflight, peak atomic.Int64

		// Atomic max via CompareAndSwap loop. The naive
		// `if n > peak.Load() { peak.Store(n) }` form races and may miss
		// the real peak.
		recordPeak := func(n int64) {
			for {
				cur := peak.Load()
				if n <= cur {
					return
				}
				if peak.CompareAndSwap(cur, n) {
					return
				}
			}
		}

		s := newSyncWithFake(func(_ context.Context, _, _ string, _ ...string) (string, string, error) {
			n := inflight.Add(1)
			recordPeak(n)
			time.Sleep(20 * time.Millisecond)
			inflight.Add(-1)
			return "", "", nil
		}, 2)

		s.execute(context.Background(), []string{"a", "b", "c", "d", "e"}, time.Second)

		is.True(peak.Load() >= 1)
		is.True(peak.Load() <= int64(s.ConNum))
	})

	t.Run("all timeout when every repo hangs", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := newSyncWithFake(func(ctx context.Context, _, _ string, _ ...string) (string, string, error) {
			<-ctx.Done()
			return "", "", ctx.Err()
		}, 1)

		stats := s.execute(context.Background(), []string{"a", "b", "c"}, 30*time.Millisecond)
		is.Equal(len(stats.timedOut), 3)
		is.Equal(len(stats.failed), 0)
		is.Equal(len(stats.succeeded), 0)
		is.Equal(len(stats.outcomes), 3)
	})

}

func TestSync_FilterRepos(t *testing.T) {
	t.Run("no filters returns all", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &Sync{}
		got, err := s.filterRepos([]string{"a", "b", "c"})
		is.NoErr(err)
		is.Equal(got, []string{"a", "b", "c"})
	})

	t.Run("target only", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &Sync{TargetDir: "^api"}
		got, err := s.filterRepos([]string{"api-foo", "web-bar", "api-baz"})
		is.NoErr(err)
		is.Equal(got, []string{"api-foo", "api-baz"})
	})

	t.Run("ignore only", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &Sync{IgnoreDir: "^test"}
		got, err := s.filterRepos([]string{"test-a", "prod-b", "test-c"})
		is.NoErr(err)
		is.Equal(got, []string{"prod-b"})
	})

	t.Run("target and ignore combined", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &Sync{TargetDir: "service$", IgnoreDir: "deprecated"}
		got, err := s.filterRepos([]string{"user-service", "deprecated-service", "web-app"})
		is.NoErr(err)
		is.Equal(got, []string{"user-service"})
	})

	t.Run("invalid target regex returns error", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &Sync{TargetDir: "[invalid("}
		_, err := s.filterRepos([]string{"a"})
		is.True(err != nil)
	})

	t.Run("invalid ignore regex returns error", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &Sync{IgnoreDir: "[invalid("}
		_, err := s.filterRepos([]string{"a"})
		is.True(err != nil)
	})
}

func TestRunSummary_HasFailures(t *testing.T) {
	t.Run("clean run returns false", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &RunSummary{Succeeded: 5, Failed: 0, TimedOut: 0}
		is.Equal(s.HasFailures(), false)
	})

	t.Run("any failure returns true", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &RunSummary{Succeeded: 4, Failed: 1, TimedOut: 0}
		is.Equal(s.HasFailures(), true)
	})

	t.Run("any timeout returns true", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &RunSummary{Succeeded: 4, Failed: 0, TimedOut: 1}
		is.Equal(s.HasFailures(), true)
	})

	t.Run("empty summary returns false (no work done)", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		s := &RunSummary{}
		is.Equal(s.HasFailures(), false)
	})
}

func TestDisplayName(t *testing.T) {
	t.Run("two segments stripped to last two", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		is.Equal(displayName("/Users/x/go/src/github.com/dxe-ai/agent"), "dxe-ai/agent")
	})

	t.Run("single parent and leaf", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		is.Equal(displayName("/tmp/agent"), "tmp/agent")
	})

	t.Run("bare leaf returns leaf only", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		// On Unix, "/agent" cleans to "/agent"; parent becomes "/" → bare leaf.
		is.Equal(displayName("/agent"), "agent")
	})

	t.Run("trailing slash is normalized", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		is.Equal(displayName("/a/b/c/"), "b/c")
	})

	t.Run("dot relative path returns just the leaf", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		is.Equal(displayName("./agent"), "agent")
	})
}
