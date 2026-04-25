package syncer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func TestIsRepo(t *testing.T) {
	t.Run("regular repo with .git directory", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		dir := t.TempDir()
		repo := filepath.Join(dir, "myrepo")
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}

		is.True(IsRepo(repo))
	})

	t.Run("worktree with .git file", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		dir := t.TempDir()
		worktree := filepath.Join(dir, "myworktree")
		if err := os.MkdirAll(worktree, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		// git worktrees place a .git file (not directory) containing
		// a `gitdir: <path>` pointer. The exact contents don't matter for
		// IsRepo — only that the entry exists.
		gitFile := filepath.Join(worktree, ".git")
		if err := os.WriteFile(gitFile, []byte("gitdir: /tmp/main/.git/worktrees/x\n"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		is.True(IsRepo(worktree))
	})

	t.Run("non-git directory returns false", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		dir := t.TempDir()
		plain := filepath.Join(dir, "plain")
		if err := os.MkdirAll(plain, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}

		is.True(!IsRepo(plain))
	})

	t.Run("non-existent directory returns false", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		is.True(!IsRepo(filepath.Join(t.TempDir(), "does-not-exist")))
	})
}
