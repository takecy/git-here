package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/takecy/git-sync/cli"
)

const version = "0.5.0"

const usage = `
git-sync is sync repositories in current directory.
more info:   https://github.com/takecy/git-sync#readme

Usage:
  git-sync [original_options] <git_command> [git_options]

Original Options:
  --target  Specific target directory with regex.
  --ignore  Specific ignore directory with regex.

Commands:
  version     Print version.
  <command>   Same as git command. (fetch, pull, status...)

Options:
  Same as git.
`

var (
	targetDir = flag.String("target", "", "")
	ignoreDir = flag.String("ignore", "", "")
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

	if flag.Arg(0) == "version" {
		fmt.Fprintf(os.Stdout, "git-sync %s\n", version)
		return
	}

	(&cli.Cmd{
		TargetDir: *targetDir,
		IgnoreDir: *ignoreDir,
		Command:   flag.Arg(0),
		Options:   flag.Args()[1:],
	}).Run()
}
