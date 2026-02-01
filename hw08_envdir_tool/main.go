package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: go-envdir <envdir> <command> [args...]")
		os.Exit(1)
	}

	envDir := os.Args[1]
	cmd := os.Args[2:]

	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(RunCmd(cmd, env))
}
