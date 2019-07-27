package main

import (
	"fmt"
	"os"
)

func isMD5(value string) {
	return
}

func md5First8192(filename string) {
	return
}

func md5File(filename string) {
	return
}

func countFiles(path string) {
	return
}

func calculateHashes() {
	return
}

func calculate() {
	return
}

func lookupHashes() {
	return
}

func lookup() {
	return
}

func fromfile() {
	return
}

func usage(progName string) {
	fmt.Printf("usage: %s <calculate|lookup|fromfile|filters> <filterfile> <file1> [file2 ...]\n", progName)
	os.Exit(1)
}

func readableFile(path string) {
	return
}

func writeableFile(path string) {
	return
}

func main() {
	progName := os.Args[0]
	if len(os.Args) < 3 {
		usage(progName)
	}
	command := os.Args[1]
	switch command {
	case "calculate":
		calculate()
	case "lookup":
		lookup()
	case "filters":
		parseFilters(os.Args)
	case "fromfile":
		fromfile()
	default:
		fmt.Printf("Invalid command: %s\n", command)
		usage(progName)
	}
}
