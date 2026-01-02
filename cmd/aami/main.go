package main

import (
	"os"

	"github.com/fregataa/aami/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
