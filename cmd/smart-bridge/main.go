package main

import (
	"os"

	"github.com/skel2007/smart-bridge/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
