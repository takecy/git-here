package syncer

import (
	"context"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestIsExist(t *testing.T) {
	is := is.New(t)
	is.NoErr(ExistGit())
}

func TestFetch(t *testing.T) {
	is := is.New(t)

	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--all", "-p"}
	_, _, err := gi.Git(context.Background(), "fetch", ".", args...)
	is.NoErr(err)
}

func TestPull(t *testing.T) {
	t.Skip()
	is := is.New(t)

	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--verbose"}
	_, _, err := gi.Git(context.Background(), "pull", ".", args...)
	is.NoErr(err)
}
