package syncer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"golang.org/x/net/context"
)

// Cmd is struct
type Cmd struct {
	// TimeOunt is timeout of performed command on one direcotory.
	TimeOunt string

	// TargetDir is target directory regex pattern.
	TargetDir string

	// IgnoreDir is ignore sync target directory regex pattern.
	IgnoreDir string

	// Command is it command.
	Command string

	// Options is git command options.
	Options []string
}

// Run is execute logic
func (s *Cmd) Run() {
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

	//
	// set up context
	//
	to, err := time.ParseDuration(s.TimeOunt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "invalid timeout value.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	fmt.Printf("exec.git.command: %s %v\n\n", s.Command, s.Options)
	key := "target.dir.path"

	//
	// execute command for directories
	//
	for _, d := range dirs {
		dctx := context.WithValue(ctx, key, d)

		err := s.callGit(dctx, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "callGit.failed: %s: %v\n", d, err)
			continue
		}
	}
}

// callGit is call git command
func (s *Cmd) callGit(ctx context.Context, d string) (err error) {
	if s.IgnoreDir != "" {
		if isMatch, _ := regexp.MatchString(s.IgnoreDir, d); isMatch {
			return
		}
	}

	if s.TargetDir != "" {
		if isMatch, _ := regexp.MatchString(s.TargetDir, d); !isMatch {
			return
		}
	}

	absPath, err := filepath.Abs(d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get.abs.failed: %s: %v\n", d, err)
		return
	}

	err = os.Chdir(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cd.failed: %s: %s: %v\n", d, absPath, err)
		return
	}

	execDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Getwd.failed: %s: %s: %v\n", d, absPath, err)
		return
	}

	fmt.Printf("exec.dir:%v\n", execDir)

	err = Git(s.Command, s.Options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Printf("error.\n\n")
	} else {
		fmt.Printf("done.\n\n")
	}

	os.Chdir("../")

	return
}
