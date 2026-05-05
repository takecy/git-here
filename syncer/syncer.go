package syncer

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
// It retains both the ordered Outcome list (for the final summary table) and
// the per-status path buckets (for cheap len() access used by RunSummary and
// existing tests).
type runStats struct {
	mu        sync.Mutex
	succeeded []string
	failed    []string
	timedOut  []string
	outcomes  []printer.Outcome
}

func (s *runStats) addOutcome(o printer.Outcome) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outcomes = append(s.outcomes, o)
	switch o.Status {
	case printer.StatusSuccess:
		s.succeeded = append(s.succeeded, o.Repo)
	case printer.StatusFailed:
		s.failed = append(s.failed, o.Repo)
	case printer.StatusTimeout:
		s.timedOut = append(s.timedOut, o.Repo)
	}
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

// printerSummary mirrors RunSummary onto the printer-side type used by
// PrintSummaryTable, avoiding an import cycle.
func (s *runStats) printerSummary() printer.Summary {
	rs := s.summary()
	return printer.Summary{
		Succeeded: rs.Succeeded,
		Failed:    rs.Failed,
		TimedOut:  rs.TimedOut,
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

	repos, err := s.filterRepos(dirs)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		// Filter narrowed down to zero — no work, but not an error.
		s.Writer.PrintMsg("No target repositories.")
		return &RunSummary{}, nil
	}

	perRepoTimeout, err := time.ParseDuration(s.TimeOut)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid timeout value: %s", s.TimeOut)
	}

	s.Writer.PrintHeader(s.Command, s.Options, len(repos))

	runStart := time.Now()
	stats := s.execute(context.Background(), repos, perRepoTimeout)
	elapsed := time.Since(runStart)

	s.Writer.PrintSummaryTable(stats.outcomes, stats.printerSummary(), elapsed)
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
// remaining ones. Each goroutine builds a printer.Outcome, stores it in
// runStats, and emits a single completed line via PrintRepoLine.
func (s *Sync) execute(parent context.Context, repos []string, perRepoTimeout time.Duration) *runStats {
	stats := &runStats{}

	eg := &errgroup.Group{}
	if s.ConNum > 0 {
		eg.SetLimit(s.ConNum)
	}

	for _, r := range repos {
		r := r
		eg.Go(func() error {
			started := time.Now()
			ctx, cancel := context.WithTimeout(parent, perRepoTimeout)
			defer cancel()

			absPath, absErr := filepath.Abs(r)
			if absErr != nil {
				o := printer.Outcome{
					Repo:     r,
					Display:  displayName(r),
					Status:   printer.StatusFailed,
					Duration: time.Since(started),
					Message:  absErr.Error(),
					Err:      errors.Wrapf(absErr, "get.abs.failed: %s", r),
				}
				stats.addOutcome(o)
				s.Writer.PrintRepoLine(o)
				return nil
			}

			msg, errMsg, err := s.Gitter.Git(ctx, s.Command, absPath, s.Options...)
			o := printer.Outcome{
				Repo:     absPath,
				Display:  displayName(absPath),
				Duration: time.Since(started),
			}
			switch {
			case err == nil:
				o.Status = printer.StatusSuccess
				o.Message = firstLine(msg)
			case errors.Is(ctx.Err(), context.DeadlineExceeded):
				o.Status = printer.StatusTimeout
				o.Message = "timeout"
				o.Err = ctx.Err()
			default:
				o.Status = printer.StatusFailed
				// Prefer the first line of git stderr when present; fall back
				// to err.Error() so failures that happen before the child
				// process can write to stderr (e.g. the repo was removed
				// between discovery and execution, or cmd.Dir cannot be
				// entered) still surface actionable text in the per-repo
				// line and summary table.
				if msg := firstLine(errMsg); msg != "" {
					o.Message = msg
				} else {
					o.Message = err.Error()
				}
				o.Err = err
			}
			stats.addOutcome(o)
			s.Writer.PrintRepoLine(o)
			return nil
		})
	}
	_ = eg.Wait()

	return stats
}

// displayName shortens an absolute repository path to its trailing two
// segments for compact display (e.g. ".../dxe-ai/agent" -> "dxe-ai/agent").
// A bare leaf is returned untouched.
func displayName(p string) string {
	p = filepath.Clean(p)
	parent, leaf := filepath.Split(p)
	parent = strings.TrimRight(parent, string(filepath.Separator))
	// Treat empty / "." / Unix root / Windows drive root (e.g. "C:") as
	// "no meaningful parent" so a repo directly under such a root renders
	// as the bare leaf instead of "C:/repo".
	if parent == "" || parent == "." || parent == string(filepath.Separator) || isDriveRoot(parent) {
		return leaf
	}
	return filepath.Base(parent) + "/" + leaf
}

// isDriveRoot reports whether s looks like a Windows drive root such as
// "C:" — i.e. a single ASCII letter followed by a colon and nothing else.
// We intentionally don't gate on runtime.GOOS so the check stays a cheap
// pure helper that's easy to test on any platform.
func isDriveRoot(s string) bool {
	if len(s) != 2 || s[1] != ':' {
		return false
	}
	c := s[0]
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// firstLine returns the first non-empty trimmed line of s, or "" if none.
func firstLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			return ln
		}
	}
	return ""
}
