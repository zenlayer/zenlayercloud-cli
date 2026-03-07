// Package main is the entry point for the zencli command line tool.
package main

import (
	"os"

	"github.com/zenlayer/zenlayercloud-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
