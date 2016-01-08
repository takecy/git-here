package cli

import (
	"io/ioutil"
	"strings"
)

func ListDirs() (dirs []string, err error) {
	files, err := ioutil.ReadDir("./")
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

func IsRepo(dirName string) bool {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return false
	}

	for _, f := range files {
		if f.Name() == ".git" {
			return true
		}
	}

	return false
}
