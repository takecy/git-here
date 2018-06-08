package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/takecy/git-here/printer"
	"github.com/takecy/git-here/syncer"
)

const version = "0.12.3"

const usage = `Run git command to all repositories in the current directory.
more info: https://github.com/takecy/git-here#readme

Usage:
  gh [original_options] <git_command> [git_options]

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

	if flag.Arg(0) == "version" {
		chackUpdate()
		fmt.Fprintf(os.Stdout, "git-here %s\n", version)
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

func chackUpdate() {
	repo := "takecy/git-here"
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		APIToken: "fa0b867fcef62ec8614dbcf2f58104630acda374", // read only of public info
	})
	if err != nil {
		fmt.Printf("Chcke update failed: %v\n", err)
		return
	}

	latest, found, err := updater.DetectLatest(repo)
	if err != nil {
		fmt.Printf("Binary update failed: %v\n", err)
		return
	}

	fmt.Printf("the latest version is %s (%s)\n", latest.Version, latest.PublishedAt.Format("2006-01-02"))

	v := semver.MustParse(version)
	if !found || latest.Version.LTE(v) {
		fmt.Printf("Current version is the latest: %s\n", version)
		return
	}

	fmt.Printf("Do you want to update to [%s] ? (y/n): \n", latest.Version)
	input := ""
	_, err = fmt.Scanln(&input)
	if err != nil {
		fmt.Printf("Invalid input\n")
		return
	}

	switch input {
	case "y":
		fmt.Printf("updating....\n")
	// next
	case "n":
		fmt.Printf("not update.\n")
		return
	default:
		fmt.Printf("invalid input.\n")
		return
	}

	updated, err := updater.UpdateSelf(v, repo)
	if err != nil {
		fmt.Printf("Error occurred while updating binary: %v\n", err)
		return
	}
	fmt.Printf("Successfully updated to version: %s\n", updated.Version)
}
