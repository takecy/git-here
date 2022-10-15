package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/takecy/git-here/printer"
	"golang.org/x/sync/errgroup"
)

// Sync is struct
type Sync struct {
	// TimeOut is timeout of performed command on one direcotory.
	TimeOut string

	// TargetDir is target directory regex pattern.
	TargetDir string

	// IgnoreDir is ignore sync target directory regex pattern.
	IgnoreDir string

	// Command is it command.
	Command string

	// Options is git command options.
	Options []string

	// Writer is instance
	Writer *printer.Printer

	// ConNum is concurrency level
	ConNum int

	// Gitter is instance
	Gitter *Gitter
}

// Run is execute logic
func (s *Sync) Run() (err error) {
	//
	// list target directories
	//
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

	//
	// retrieve target repos
	//
	repos := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if s.IgnoreDir != "" {
			if isMatch, _ := regexp.MatchString(s.IgnoreDir, d); isMatch {
				continue
			}
		}

		if s.TargetDir != "" {
			if isMatch, _ := regexp.MatchString(s.TargetDir, d); !isMatch {
				continue
			}
		}

		repos = append(repos, d)
	}

	if len(repos) == 0 {
		s.Writer.PrintMsg("No target repositories.")
		return
	}

	targetRepoNum := len(repos)
	s.Writer.PrintMsg(fmt.Sprintf("target repositories: (%d)", targetRepoNum))

	//
	// execute command
	//
	eg := errgroup.Group{}
	start := time.Now()
	throttle := make(chan struct{}, s.ConNum)

	// set up context
	to, err := time.ParseDuration(s.TimeOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "invalid timeout value.")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	for i := range repos {
		num := i + 1
		r := repos[i]
		throttle <- struct{}{}

		eg.Go(func() error {
			defer func() {
				<-throttle
			}()

			err := s.execCmd(ctx, r)
			if err != nil {
				s.Writer.PrintMsgErr(fmt.Sprintf("Failed: %s\n%v", r, err))
			} else {
				s.Writer.PrintMsg(fmt.Sprintf("Success: %s\n", r))
			}

			s.Writer.PrintMsg(fmt.Sprintf("Done: %d/%d", num, targetRepoNum))
			return nil
		})
	}

	go func() {
		for {
			<-ctx.Done()
			s.Writer.PrintMsgErr(fmt.Sprintf("---- Timeouted (%v) [%v]----", time.Since(start).String(), ctx.Err()))
			os.Exit(1)
		}
	}()

	if err := eg.Wait(); err != nil {
		s.Writer.PrintMsgErr(fmt.Sprintf("Error.exists: %v", err))
	}

	s.Writer.PrintMsg(fmt.Sprintf("All done. (%v)", time.Since(start).Round(time.Millisecond)))
	return
}

// execCmd is execute git command
func (s *Sync) execCmd(ctx context.Context, d string) (err error) {
	absPath, err := filepath.Abs(d)
	if err != nil {
		err = errors.Wrapf(err, "get.abs.failed: %s", d)
		s.Writer.Error(printer.Result{Err: err})
		return
	}

	msg, errMsg, err := s.Gitter.Git(s.Command, absPath, s.Options...)
	if err != nil {
		s.Writer.Error(printer.Result{Repo: absPath, Err: errors.Wrapf(err, errMsg)})
	} else {
		s.Writer.Print(printer.Result{Repo: absPath, Msg: msg})
	}

	return
}
