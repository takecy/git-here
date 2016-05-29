package syncer

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

// Git is execute git command
func Git(command string, args ...string) error {
	cmdArgs := append([]string{command}, args...)
	c := exec.Command("git", cmdArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
