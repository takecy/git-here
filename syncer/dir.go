package syncer

import (
	"os"
	"path/filepath"
	"strings"
)

// ListDirs returns target directories
func ListDirs() (dirs []string, err error) {
	files, err := os.ReadDir("./")
	if err != nil {
		return
	}

	dirs = make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() && IsRepo(f.Name()) && strings.Index(f.Name(), ".") != 0 {
			dirs = append(dirs, f.Name())
		}
	}

	return
}

// IsRepo reports whether dirName contains a `.git` entry. It accepts both a
// regular repository (where `.git` is a directory) and a git worktree
// (where `.git` is a file pointing at the main repo's git directory),
// because exec'd git commands work in either layout.
func IsRepo(dirName string) bool {
	_, err := os.Stat(filepath.Join(dirName, ".git"))
	return err == nil
}
