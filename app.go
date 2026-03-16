package main

import (
	"fmt"
	"os"

	"github.com/elC0mpa/aws-doctor/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	return cmd.Execute(version, commit, date)
}
