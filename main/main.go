package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var (
		input string
	)

	flag.StringVar(&input, "input", input, "the input")

	flag.Parse()

	return nil
}
