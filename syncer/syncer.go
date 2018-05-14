package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/takecy/git-sync/printer"
	"golang.org/x/sync/errgroup"
)

// Cmd is struct
type Cmd struct {
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

	// Giter is instance
	Giter *Giter
}

// Run is execute logic
func (s *Cmd) Run() (err error) {
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

	fmt.Printf("repositorie are found: (%d)\n", len(dirs))

	//
	// set up context
	//
	to, err := time.ParseDuration(s.TimeOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "invalid timeout value.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

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
		s.Writer.PrintMsg(fmt.Sprintf("No target repositories."))
		return
	}

	s.Writer.PrintMsg(fmt.Sprintf("target repositorie: (%d)", len(repos)))

	//
	// execute command
	//
	eg := errgroup.Group{}
	start := time.Now()

	throttle := make(chan struct{}, runtime.NumCPU())

	yets := sync.Map{}
	done := make(chan struct{}, len(repos))

	for i := range repos {
		r := repos[i]
		yets.Store(r, struct{}{})

		throttle <- struct{}{}

		eg.Go(func() error {
			err := s.callGit(ctx, r)
			<-throttle
			yets.Delete(r)
			done <- struct{}{}
			return err
		})
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.Writer.PrintMsgErr(fmt.Sprintf("---- Timeouted (done:%d) ----", len(done)))
				fails := make([]string, 0, len(repos))
				yets.Range(func(key, val interface{}) bool {
					fails = append(fails, key.(string))
					return true
				})
				s.Writer.PrintRepoErr(fmt.Sprintf("Failrues (%d)", len(fails)), fails)
				os.Exit(1)
			}
		}
	}()

	if err := eg.Wait(); err != nil {
		s.Writer.PrintMsgErr(fmt.Sprintf("Error.exists: %v", err))
	}

	s.Writer.PrintMsg(fmt.Sprintf("All done. (%v)", time.Now().Sub(start)))
	return
}

// callGit is call git command
func (s *Cmd) callGit(ctx context.Context, d string) (err error) {
	absPath, err := filepath.Abs(d)
	if err != nil {
		err = errors.Wrapf(err, "get.abs.failed: %s", d)
		s.Writer.Error(printer.Result{Err: err})
		return
	}

	msg, errMsg, err := s.Giter.Git(s.Command, absPath, s.Options...)
	if err != nil {
		s.Writer.Error(printer.Result{Repo: absPath, Err: errors.Wrapf(err, errMsg)})
	} else {
		s.Writer.Print(printer.Result{Repo: absPath, Msg: msg})
	}

	return
}
