package cli

import (
	"fmt"
	"os"
)

type Cmd struct {
	Args []string
	Fn   func(...string) error
}

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
