package main

import (
	"os"

	"github.com/gbrindisi/littlebox/cmd/littlebox/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
