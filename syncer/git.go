package syncer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Executor abstracts git command execution. Defined as an interface so that
// callers can substitute fakes in tests without spawning real processes.
type Executor interface {
	Git(ctx context.Context, command, dir string, args ...string) (msg, errMsg string, err error)
}

// Gitter is struct
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
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "%s", s)
	if err != nil {
		return err
	}
	return nil
}

// Git is execute git command. The given context is forwarded to exec.CommandContext
// so that cancellation (e.g. per-repo timeout) terminates the underlying process.
func (g *Gitter) Git(ctx context.Context, command, dir string, args ...string) (msg, errMsg string, err error) {
	wr := new(bytes.Buffer)
	errWr := new(bytes.Buffer)

	cmdArgs := append([]string{command}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = wr
	cmd.Stderr = errWr

	err = cmd.Run()
	msg = wr.String()
	errMsg = errWr.String()
	return
}
