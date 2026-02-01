package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyWholeFile(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	data := []byte("hello world")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Copy(src, dst, 0, 0); err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != string(data) {
		t.Fatalf("expected %q, got %q", data, result)
	}
}

func TestCopyWithOffsetAndLimit(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	data := []byte("abcdef")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Copy(src, dst, 2, 3); err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != "cde" {
		t.Fatalf("expected cde, got %q", result)
	}
}

func TestOffsetExceedsFile(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	if err := os.WriteFile(src, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Copy(src, dst, 10, 0)
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Fatalf("expected ErrOffsetExceedsFileSize, got %v", err)
	}
}

func TestLimitExceedsFile(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	data := []byte("abc")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Copy(src, dst, 0, 100); err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != string(data) {
		t.Fatalf("expected %q, got %q", data, result)
	}
}
