package main

import (
	"os"
)

func main() {
	parser := Parser{Args: os.Args}
	parser.Parse()
}
