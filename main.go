package main

import (
	"os"

	"github.com/spf13/afero"
)

func main() {
	fs := afero.NewOsFs()
	parser := Parser{Args: os.Args, Fs: fs}
	parser.Parse()
}
