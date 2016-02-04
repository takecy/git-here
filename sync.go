package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/takecy/git-sync/cli"
)

const version = "0.3.0"

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

Original Options:
  --target-dir  Specific target directory with regex.
  --ignore-dir  Specific ignore directory with regex.
`

var (
	targetDir = flag.String("target-dir", "", "")
	ignoreDir = flag.String("ignore-dir", "", "")
)

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
		TargetDir: *targetDir,
		IgnoreDir: *ignoreDir,
		Args:      flag.Args()[1:],
		Fn:        f,
	}).Run()
}
