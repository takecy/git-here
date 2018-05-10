package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/takecy/git-sync/printer"
	"github.com/takecy/git-sync/syncer"
)

const version = "0.11.5"

const usage = `
git-sync is sync repositories in current directory.
more info:   https://github.com/takecy/git-sync#readme

Usage:
  gs [original_options] <git_command> [git_options]

Original Options:
  --target   Specific target directory with regex.
  --ignore   Specific ignore directory with regex.
	--timeout  Specific timeout of performed commnad during on one directory. (5s, 10m...) (default: 30s)

Commands:
  version     Print version.
  <command>   Same as git command. (fetch, pull, status...)

Options:
  Same as git.
`

var (
	targetDir = flag.String("target", "", "")
	ignoreDir = flag.String("ignore", "", "")
	timeout   = flag.String("timeout", "30s", "")
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

	writer := os.Stdout
	errWriter := os.Stderr

	err := (&syncer.Cmd{
		TargetDir: *targetDir,
		IgnoreDir: *ignoreDir,
		TimeOut:   *timeout,
		Command:   flag.Arg(0),
		Options:   flag.Args()[1:],
		Giter:     syncer.NewGiter(writer, errWriter),
		Writer:    printer.NewPrinter(writer, errWriter),
	}).Run()

	if err != nil {
		panic(err)
	}
}
