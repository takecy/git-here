package syncer

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

// Executor abstracts git command execution. Defined as an interface so that
// callers can substitute fakes in tests without spawning real processes.
type Executor interface {
	Git(ctx context.Context, command, dir string, args ...string) (msg, errMsg string, err error)
}

// ExistGit reports whether the `git` executable is available on PATH.
// It returns the error from exec.LookPath when not found, and nil otherwise.
// Unlike the previous Gitter.IsExist, this function has no side effects.
func ExistGit() error {
	_, err := exec.LookPath("git")
	return err
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
