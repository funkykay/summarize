package main

import (
	"os"

	"github.com/funkykay/summarize/internal/cli"
)

func main() {
	os.Exit(cli.New(os.Stdout, os.Stderr).Run(os.Args[1:]))
}
