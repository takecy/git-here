package syncer

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestIsExist(t *testing.T) {
	is := is.New(t)

	gi := NewGitter(os.Stdout, os.Stderr)
	err := gi.IsExist()
	is.NoErr(err)
}

func TestFetch(t *testing.T) {
	is := is.New(t)

	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--all", "-p"}
	_, _, err := gi.Git("fetch", ".", args...)
	is.NoErr(err)
}

func TestPull(t *testing.T) {
	t.Skip()
	is := is.New(t)

	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--verbose"}
	_, _, err := gi.Git("pull", ".", args...)
	is.NoErr(err)
}
