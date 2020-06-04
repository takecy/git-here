package syncer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Gitter is strcut
type Gitter struct {
	writer    io.Writer
	errWriter io.Writer
}

// NewGitter is constructor
func NewGitter(writer, errWriter io.Writer) *Gitter {
	return &Gitter{
		writer:    writer,
		errWriter: errWriter,
	}
}

// IsExist is check git command
func (*Gitter) IsExist() error {
	s, err := exec.LookPath("git")
	fmt.Fprintf(os.Stdout, "%s", s)
	return err
}

// Git is execute git command
func (g *Gitter) Git(command, dir string, args ...string) (msg, errMsg string, err error) {
	wr := new(bytes.Buffer)
	errWr := new(bytes.Buffer)

	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = wr
	cmd.Stderr = errWr

	err = cmd.Run()
	msg = wr.String()
	errMsg = errWr.String()
	return
}
