package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	input := "Hello, OTUS!"
	output := reverse.String(input)
	fmt.Println(output)
}
