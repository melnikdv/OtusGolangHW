package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer fromFile.Close()

	info, err := fromFile.Stat()
	if err != nil {
		return ErrUnsupportedFile
	}

	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	size := info.Size()

	if offset > size {
		return ErrOffsetExceedsFileSize
	}

	var toCopy int64
	if limit == 0 || offset+limit > size {
		toCopy = size - offset
	} else {
		toCopy = limit
	}

	if _, err := fromFile.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	toFile, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("create target: %w", err)
	}
	defer toFile.Close()

	const bufSize = 32 * 1024
	buf := make([]byte, bufSize)

	var copied int64

	for copied < toCopy {
		toRead := bufSize
		if remaining := toCopy - copied; remaining < int64(bufSize) {
			toRead = int(remaining)
		}

		n, readErr := fromFile.Read(buf[:toRead])
		if n > 0 {
			if _, err := toFile.Write(buf[:n]); err != nil {
				return fmt.Errorf("write: %w", err)
			}
			copied += int64(n)
			printProgress(copied, toCopy)
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return fmt.Errorf("read: %w", readErr)
		}
	}

	fmt.Println()
	return nil
}

func printProgress(done, total int64) {
	percent := int(float64(done) / float64(total) * 100)
	fmt.Printf("\rProgress: %d%%", percent)
}
