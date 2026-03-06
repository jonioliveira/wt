// Package main is the entry point for the wt CLI.
package main

import (
	"fmt"
	"os"

	"github.com/jonioliveira/wt/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
