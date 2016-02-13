package cli

import "testing"

func TestIsExist(t *testing.T) {
	err := IsExist()
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestFetch(t *testing.T) {
	args := []string{"--all", "-p"}
	err := Fetch(args...)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestPull(t *testing.T) {
	args := []string{"--verbose"}
	err := Pull(args...)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
