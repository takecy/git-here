package cli

import (
	"fmt"
	"os"
	"os/exec"
)

// IsExist is check git command
func IsExist() error {
	s, err := exec.LookPath("git")
	fmt.Fprintf(os.Stdout, "%s", s)
	return err
}

// Fetch is git fetch
func Fetch(args ...string) error {
	cmdArgs := append([]string{"fetch"}, args...)
	c := exec.Command("git", cmdArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

// Pull is git pull
func Pull(args ...string) error {
	cmdArgs := append([]string{"pull"}, args...)
	c := exec.Command("git", cmdArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
