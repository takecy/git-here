package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/takecy/git-sync/cli"
)

const version = "0.2.0"

const usage = `
git-sync is sync repositories in current directory.
more info:   https://github.com/takecy/git-sync#readme

Usage:
  git-sync <command> [options]

Commands:
  version  Print version.
  fetch    Alias for <git fetch>.
  pull     Alias for <git pull>.

Options:
  Same as git.
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	var f func(...string) error
	switch flag.Arg(0) {
	case "version":
		fmt.Fprintf(os.Stdout, "git-sync %s\n", version)
		return
	case "fetch":
		f = cli.Fetch
	case "pull":
		f = cli.Pull
	default:
		flag.Usage()
	}

	(&cli.Cmd{
		Args: flag.Args()[1:],
		Fn:   f,
	}).Run()
}
