package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/takecy/git-sync/cli"
)

const usage = `
git-sync is sync repositories in current directory.
more info:   https://github.com/takecy/git-sync#readme

Usage:
  git-sync <command> [options]

Commands:
  fetch   Alias for <git fetch>.
  pull    Alias for <git pull>.

Options:
  same of git.

`

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usageAndExit
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	run(flag.Arg(0))
}

func run(cmd string) {
	dirs, err := cli.ListDirs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	var f func(...string) error
	switch cmd {
	case "fetch":
		f = cli.Fetch
	case "pull":
		f = cli.Pull
	default:
		flag.Usage()
	}

	for _, d := range dirs {
		os.Chdir(d)
		cd, _ := os.Getwd()
		fmt.Fprintf(os.Stdout, "exec.dir:%v\n", cd)

		err = f(flag.Args()[1:]...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			fmt.Fprint(os.Stdout, "done\n\n")
		}

		os.Chdir("../")
	}

}

func usageAndExit() {
	fmt.Fprintf(os.Stderr, "%s\n", usage)
	os.Exit(1)
}
