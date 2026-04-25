package syncer

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestExistGit_NoStdoutSideEffect(t *testing.T) {
	// Cannot use t.Parallel(): mutates global os.Stdout.
	is := is.New(t)

	r, w, err := os.Pipe()
	is.NoErr(err)
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	err = ExistGit()
	is.NoErr(err)

	if cerr := w.Close(); cerr != nil {
		t.Fatalf("close pipe writer: %v", cerr)
	}
	out, rerr := io.ReadAll(r)
	is.NoErr(rerr)
	is.Equal(string(out), "") // ExistGit must not write anything to stdout
}

func TestFetch(t *testing.T) {
	is := is.New(t)

	// bare repo serves as the "remote". Fetching from it never reaches the
	// network, so this test is hermetic and CI-stable.
	remote := t.TempDir()
	runGit(t, remote, "init", "--bare", "-q")

	// working repo: register the bare repo as origin and push one commit so
	// there is something for the fetch to discover.
	work := t.TempDir()
	runGit(t, work, "init", "-q")
	runGit(t, work, "config", "user.email", "test@example.invalid")
	runGit(t, work, "config", "user.name", "tester")
	runGit(t, work, "config", "commit.gpgsign", "false")
	runGit(t, work, "commit", "--allow-empty", "-m", "init", "-q")
	runGit(t, work, "remote", "add", "origin", remote)
	runGit(t, work, "push", "-q", "origin", "HEAD:refs/heads/main")

	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--all", "-p"}
	_, _, err := gi.Git(context.Background(), "fetch", work, args...)
	is.NoErr(err)

	// err==nil alone would let a no-op fetch pass. Verify the fetch actually
	// populated the remote-tracking ref by resolving refs/remotes/origin/main.
	cmd := exec.Command("git", "-C", work, "rev-parse", "refs/remotes/origin/main")
	out, rpErr := cmd.CombinedOutput()
	if rpErr != nil {
		t.Fatalf("rev-parse refs/remotes/origin/main failed: %v\n%s", rpErr, out)
	}
	sha := strings.TrimSpace(string(out))
	is.True(sha != "")
}

func TestPull(t *testing.T) {
	t.Skip()
	is := is.New(t)

	gi := NewGitter(os.Stdout, os.Stderr)
	args := []string{"--verbose"}
	_, _, err := gi.Git(context.Background(), "pull", ".", args...)
	is.NoErr(err)
}

// runGit is a fixture helper. Failures here mean the test setup is broken,
// not the system under test, so abort early with t.Fatalf rather than is.NoErr.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, out)
	}
}
