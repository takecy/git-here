package syncer

import (
	"os"
	"testing"
)

func TestIsExist(t *testing.T) {
	gi := NewGitter(os.Stdout, os.Stderr)
	err := gi.IsExist()
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestFetch(t *testing.T) {
	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--all", "-p"}
	_, _, err := gi.Git("fetch", ".", args...)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestPull(t *testing.T) {
	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--verbose"}
	_, _, err := gi.Git("pull", ".", args...)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
