package cli

import (
	"fmt"
	"os"
	"regexp"
)

// Cmd is struct
type Cmd struct {
	// sync target directory
	TargetDir string

	// ignore sync target directory
	IgnoreDir string

	// git command args
	Args []string

	// git command
	Fn func(...string) error
}

// Run is run command
func (s *Cmd) Run() {
	dirs, err := ListDirs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if len(dirs) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", "There is no repositories...")
		os.Exit(1)
	}

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

		os.Chdir(d)
		cd, _ := os.Getwd()
		fmt.Fprintf(os.Stdout, "exec.dir:%v\n", cd)

		err = s.Fn(s.Args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			fmt.Fprint(os.Stdout, "error.\n\n")
		} else {
			fmt.Fprint(os.Stdout, "done.\n\n")
		}

		os.Chdir("../")
	}
}
