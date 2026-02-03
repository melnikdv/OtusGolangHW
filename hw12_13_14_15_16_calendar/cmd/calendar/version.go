package main

import "fmt"

var (
	release   = "dev"
	buildDate = "unknown"
	gitHash   = "unknown"
)

func printVersion() {
	fmt.Printf("calendar version %s\n", release)
	fmt.Printf("build time: %s\n", buildDate)
	fmt.Printf("git commit: %s\n", gitHash)
}
