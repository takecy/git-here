package syncer

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/takecy/git-here/printer"
	"golang.org/x/sync/errgroup"
)

// RunSummary reports per-category counts after a Sync.Run completes
// successfully (i.e. the run itself was not aborted by a setup error).
type RunSummary struct {
	Succeeded int
	Failed    int
	TimedOut  int
}

// HasFailures reports whether any repository failed or timed out.
// Used by callers to decide between exit code 0 (clean) and 2 (partial).
func (s *RunSummary) HasFailures() bool {
	return s.Failed > 0 || s.TimedOut > 0
}

// Sync is struct
type Sync struct {
	// TimeOut is timeout of performed command on one directory.
	TimeOut string

	// TargetDir is target directory regex pattern.
	TargetDir string

	// IgnoreDir is ignore sync target directory regex pattern.
	IgnoreDir string

	// Command is the git subcommand to execute.
	Command string

	// Options is git command options.
	Options []string

	// Writer is the printer instance for output formatting.
	Writer *printer.Printer

	// ConNum is concurrency level
	ConNum int

	// Gitter is the git command executor instance.
	Gitter Executor
}

// runStats accumulates per-repository outcomes safely from concurrent goroutines.
// Each bucket is just a list of repository paths — the per-repo error message
// is already streamed to stderr at execution time via PrintMsgErr, so there is
// no need to retain it here.
type runStats struct {
	mu        sync.Mutex
	succeeded []string
	failed    []string
	timedOut  []string
}

func (s *runStats) addSuccess(r string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.succeeded = append(s.succeeded, r)
}

func (s *runStats) addFailed(r string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed = append(s.failed, r)
}

func (s *runStats) addTimedOut(r string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timedOut = append(s.timedOut, r)
}

// summary builds a public RunSummary snapshot of the current stats.
func (s *runStats) summary() *RunSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &RunSummary{
		Succeeded: len(s.succeeded),
		Failed:    len(s.failed),
		TimedOut:  len(s.timedOut),
	}
}

// Run discovers repositories, applies filters, and executes the configured
// git command across the matching set. It returns a summary of per-repo
// outcomes, or an error for setup failures (no repositories, invalid regex,
// invalid timeout, etc.). The returned summary is non-nil exactly when err
// is nil; callers can inspect summary.HasFailures() to decide between a
// clean run (exit 0) and a partial-failure run (exit 2).
func (s *Sync) Run() (*RunSummary, error) {
	dirs, err := ListDirs()
	if err != nil {
		return nil, errors.Wrap(err, "list directories")
	}
	if len(dirs) == 0 {
		return nil, errors.New("no git repositories found in current directory")
	}

	fmt.Printf("repositories are found: (%d)\n", len(dirs))
	s.Writer.PrintCmd(s.Command, s.Options)

	repos, err := s.filterRepos(dirs)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		// Filter narrowed down to zero — no work, but not an error.
		s.Writer.PrintMsg("No target repositories.")
		return &RunSummary{}, nil
	}
	s.Writer.PrintMsg(fmt.Sprintf("target repositories: (%d)", len(repos)))

	perRepoTimeout, err := time.ParseDuration(s.TimeOut)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid timeout value: %s", s.TimeOut)
	}

	stats := s.execute(context.Background(), repos, perRepoTimeout)
	s.printSummary(stats)
	return stats.summary(), nil
}

// filterRepos applies the target/ignore regex patterns to the discovered
// directories. It is split out from Run() so tests can exercise the matching
// logic without scanning the filesystem.
func (s *Sync) filterRepos(dirs []string) ([]string, error) {
	var ignoreRegex, targetRegex *regexp.Regexp

	if s.IgnoreDir != "" {
		re, err := regexp.Compile(s.IgnoreDir)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid ignore directory regex pattern: %s", s.IgnoreDir)
		}
		ignoreRegex = re
	}

	if s.TargetDir != "" {
		re, err := regexp.Compile(s.TargetDir)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid target directory regex pattern: %s", s.TargetDir)
		}
		targetRegex = re
	}

	repos := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if ignoreRegex != nil && ignoreRegex.MatchString(d) {
			continue
		}
		if targetRegex != nil && !targetRegex.MatchString(d) {
			continue
		}
		repos = append(repos, d)
	}
	return repos, nil
}

// execute runs the git command across all repos in parallel, throttled by
// ConNum. Each invocation is bounded by perRepoTimeout via a derived context;
// the parent context is *not* timed out so a slow repo never starves the
// remaining ones.
func (s *Sync) execute(parent context.Context, repos []string, perRepoTimeout time.Duration) *runStats {
	stats := &runStats{}
	total := len(repos)
	var done atomic.Int64
	start := time.Now()

	eg := &errgroup.Group{}
	if s.ConNum > 0 {
		eg.SetLimit(s.ConNum)
	}

	for _, r := range repos {
		r := r
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(parent, perRepoTimeout)
			defer cancel()

			err := s.execCmd(ctx, r)
			switch {
			case err == nil:
				stats.addSuccess(r)
				s.Writer.PrintMsg(fmt.Sprintf("Success: %s\n", r))
			case errors.Is(ctx.Err(), context.DeadlineExceeded):
				stats.addTimedOut(r)
				s.Writer.PrintMsgErr(fmt.Sprintf("Timeout: %s", r))
			default:
				stats.addFailed(r)
				s.Writer.PrintMsgErr(fmt.Sprintf("Failed: %s\n%v", r, err))
			}

			n := done.Add(1)
			s.Writer.PrintMsg(fmt.Sprintf("Done: %d/%d", n, total))
			return nil
		})
	}
	_ = eg.Wait()

	s.Writer.PrintMsg(fmt.Sprintf("All done. (%v)", time.Since(start).Round(time.Millisecond)))
	return stats
}

// execCmd is execute git command
func (s *Sync) execCmd(ctx context.Context, d string) error {
	absPath, err := filepath.Abs(d)
	if err != nil {
		return errors.Wrapf(err, "get.abs.failed: %s", d)
	}

	msg, errMsg, err := s.Gitter.Git(ctx, s.Command, absPath, s.Options...)
	if err != nil {
		return errors.Wrapf(err, "%s", errMsg)
	}
	s.Writer.Print(printer.Result{Repo: absPath, Msg: msg})
	return nil
}

// printSummary emits the post-run summary: counts, plus the list of failed
// and timed-out repositories. Reuses printer.PrintRepoErr (previously unused).
func (s *Sync) printSummary(stats *runStats) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	s.Writer.PrintMsg(fmt.Sprintf(
		"Summary: success=%d failed=%d timeout=%d",
		len(stats.succeeded), len(stats.failed), len(stats.timedOut),
	))

	if len(stats.failed) > 0 {
		s.Writer.PrintRepoErr("Failed repositories:", stats.failed)
	}
	if len(stats.timedOut) > 0 {
		s.Writer.PrintRepoErr("Timed out repositories:", stats.timedOut)
	}
}
