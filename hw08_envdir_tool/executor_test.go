package main

import (
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	env := Environment{
		"FOO": {Value: "bar"},
	}

	code := RunCmd([]string{"sh", "-c", "test \"$FOO\" = bar"}, env)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunCmdRemoveEnv(t *testing.T) {
	os.Setenv("REMOVE_ME", "123")
	defer os.Unsetenv("REMOVE_ME")

	env := Environment{
		"REMOVE_ME": {NeedRemove: true},
	}

	code := RunCmd([]string{"sh", "-c", "test -z \"$REMOVE_ME\""}, env)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
