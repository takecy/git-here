package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/takecy/git-here/printer"
	"golang.org/x/sync/errgroup"
)

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

// failedRepo carries the repository path together with the error that caused
// the failure so that it can be reported in the final summary.
type failedRepo struct {
	Repo string
	Err  error
}

// runStats accumulates per-repository outcomes safely from concurrent goroutines.
type runStats struct {
	mu        sync.Mutex
	succeeded []string
	failed    []failedRepo
	timedOut  []string
}

func (s *runStats) addSuccess(r string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.succeeded = append(s.succeeded, r)
}

func (s *runStats) addFailed(r string, e error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed = append(s.failed, failedRepo{Repo: r, Err: e})
}

func (s *runStats) addTimedOut(r string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timedOut = append(s.timedOut, r)
}

// Run is execute logic
func (s *Sync) Run() (err error) {
	dirs, err := ListDirs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if len(dirs) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", "There is no repositories...")
		os.Exit(1)
	}

	fmt.Printf("repositories are found: (%d)\n", len(dirs))
	s.Writer.PrintCmd(s.Command, s.Options)

	repos, err := s.filterRepos(dirs)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		s.Writer.PrintMsg("No target repositories.")
		return
	}
	s.Writer.PrintMsg(fmt.Sprintf("target repositories: (%d)", len(repos)))

	perRepoTimeout, err := time.ParseDuration(s.TimeOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "invalid timeout value.")
		os.Exit(1)
	}

	stats := s.execute(context.Background(), repos, perRepoTimeout)
	s.printSummary(stats)
	return
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
				stats.addFailed(r, err)
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
		repos := make([]string, len(stats.failed))
		for i, f := range stats.failed {
			repos[i] = f.Repo
		}
		s.Writer.PrintRepoErr("Failed repositories:", repos)
	}
	if len(stats.timedOut) > 0 {
		s.Writer.PrintRepoErr("Timed out repositories:", stats.timedOut)
	}
}
