package main

import (
	"fmt"
	"os"

	"gitlab.com/kinnalru/syncerman/internal/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	return cmd.Execute()
}
