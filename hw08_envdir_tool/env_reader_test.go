package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "FOO"), []byte("123\n456"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "BAR"), []byte("value \t"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "EMPTY"), []byte(""), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env["FOO"].Value != "123" {
		t.Errorf("FOO expected %q, got %q", "123", env["FOO"].Value)
	}

	if env["BAR"].Value != "value" {
		t.Errorf("BAR expected %q, got %q", "value", env["BAR"].Value)
	}

	if !env["EMPTY"].NeedRemove {
		t.Errorf("EMPTY should be marked for removal")
	}
}

func TestReadDirInvalidName(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "BAD=NAME"), []byte("123"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ReadDir(dir)
	if err == nil {
		t.Fatal("expected error for invalid env var name")
	}
}
