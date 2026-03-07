// Package main is the entry point for the zeno command line tool.
package main

import (
	"fmt"
	"os"

	"github.com/zenlayer/zenlayercloud-cli/cmd"
)

func main() {
	if err := cmd.Execute(ApisFS); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'zeno --help' for usage.\n")
		os.Exit(1)
	}
}
