package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/takecy/git-here/printer"
	"github.com/takecy/git-here/syncer"
)

// set by build
var (
	version = "0.16.0"
)

const usage = `Run git command to all repositories in the current directory.
more info: https://github.com/takecy/git-here#readme

Usage:
  gih [original_options] <git_command> [git_options]

Original Options:
  --target   Specific target directory with regex.
  --ignore   Specific ignore directory with regex.
  --timeout  Specific timeout of performed command during on one directory. (5s, 10m...) (default: 20s)

Commands:
  version     Print version. Whether check new version exists, and ask you to upgrade to latest version.
  <command>   Same as git command. (fetch, pull, status...)

Options:
  Same as git.
`

var (
	targetDir = flag.String("target", "", "")
	ignoreDir = flag.String("ignore", "", "")
	conNum    = flag.Int("c", runtime.NumCPU(), "concurrency level")
	timeout   = flag.String("timeout", "20s", "")
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

	switch flag.Arg(0) {
	case "version":
		printVersion()
		return
	}

	if *conNum == 0 {
		*conNum = runtime.NumCPU()
	}

	fmt.Printf("args: targetDir: %s ignoreDir: %s concurrency: %d timeout: %v\n", *targetDir, *ignoreDir, *conNum, *timeout)

	writer := os.Stdout

	summary, err := (&syncer.Sync{
		TargetDir: *targetDir,
		IgnoreDir: *ignoreDir,
		TimeOut:   *timeout,
		Command:   flag.Arg(0),
		Options:   flag.Args()[1:],
		ConNum:    *conNum,
		Gitter:    syncer.NewGitter(writer, os.Stderr),
		Writer:    printer.NewPrinter(writer),
	}).Run()

	// Map run outcome to exit codes per README spec:
	//   0 - all operations completed successfully
	//   1 - setup error (invalid args, no repos, invalid regex/timeout)
	//   2 - some repositories failed or timed out
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if summary != nil && summary.HasFailures() {
		os.Exit(2)
	}
}

func printVersion() {
	_, err := fmt.Fprintf(os.Stdout, "git-here %s\n", version)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(os.Stdout, "go version %s\n", runtime.Version())
	if err != nil {
		panic(err)
	}
}
