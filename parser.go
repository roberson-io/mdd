package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func usage(progName string) {
	fmt.Printf("usage: %s <calculate|lookup|fromfile|filters> <filterfile> <file1> [file2 ...]\n", progName)
	os.Exit(1)
}

func isMD5(value string) bool {
	if len([]rune(value)) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	if err != nil {
		return false
	}
	return true
}

func md5First8192(filename string) {
	return
}

func md5File(path string) string {
	f, err := os.Open(path)
	if os.IsPermission(err) {
		return ""
	}
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func countFiles(path string) int32 {
	var count int32
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q while counting files: %v\n", path, err)
			return err
		}
		if info.Mode().IsRegular() {
			count++
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return count
}

func readableFile(path string) bool {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return false
	}
	return true
}

func writeableFile(path string) bool {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return false
	}
	return true
}

// Parser for command line.
type Parser struct {
	Args []string
}

// Calculate command parser.
func (p Parser) Calculate() {
	progName := p.Args[0]
	if len(p.Args) < 4 {
		usage(progName)
	}
	filterFile := p.Args[2]
	files := p.Args[3:]
	if !writeableFile(filterFile) {
		fmt.Printf("[-] Unable to open %s for writing\n", filterFile)
		usage(progName)
	}

	fmt.Print("[+] Counting files. This may take a while\n")
	var size int32
	for _, file := range files {
		size += countFiles(file)
	}
	fmt.Printf("Counted %d files.\n", size)

	bloomFilter := NewBloomFilter(size, 0.01)

	fmt.Print("[+] Calculating hashes.\n")

	for _, file := range files {
		bloomFilter.CalculateHashes(file)
	}

	fmt.Printf(
		"[+] Saving %s filter to outfile: %s\n",
		bloomFilter.ByteSizeHuman,
		filterFile,
	)

	bloomFilter.Save(filterFile)
	fmt.Print("[+] Done.\n")
}

// Filters command parser.
func (p Parser) Filters() {
	command := p.Args[2]
	switch command {
	case "fetch":
		if len(p.Args) > 3 {
			target := p.Args[3]
			fetchFilter(target)
		} else {
			usage(p.Args[0])
		}
	case "list":
		config := getConfig()
		if len(p.Args) > 3 {
			target := p.Args[3]
			if target == "remote" {
				listRemote(config.Repo)
			} else {
				listRemote(target)
			}
		} else {
			listLocal(config.HashAlg)
		}
	case "update":
		updateFilters()
	default:
		fmt.Printf("Invalid command: %s\n", command)
		usage(p.Args[0])
	}
}

// FromFile command parser.
func (p Parser) FromFile() {
	progName := p.Args[0]
	if len(p.Args) < 4 {
		usage(progName)
	}
	filterFile := p.Args[2]
	files := p.Args[3:]
	if !writeableFile(filterFile) {
		fmt.Printf("[-] Unable to open %s for writing\n", filterFile)
		usage(progName)
	}

	fmt.Printf("[+] Counting hashes in %s\n", files)
	var count int32
	for _, hashFile := range files {
		fmt.Printf("%s\n", hashFile)
		f, err := os.Open(hashFile)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !(strings.HasPrefix(line, "#")) && isMD5(line) {
				count++
			}
		}
	}

	fmt.Printf("    Counted %d files.\n", count)

	bloomFilter := NewBloomFilter(count, 0.01)

	fmt.Printf("[+] Adding hashes from %s\n", files)

	for _, hashFile := range files {
		fmt.Printf("%s\n", hashFile)
		f, err := os.Open(hashFile)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !(strings.HasPrefix(line, "#")) && isMD5(line) {
				bloomFilter.Add(strings.ToLower(line))
			}
		}
	}

	fmt.Printf(
		"[+] Saving %s filter to outfile: %s\n",
		bloomFilter.ByteSizeHuman,
		filterFile,
	)
	bloomFilter.Save(filterFile)
	fmt.Print("[+] Done.\n")
}

// Lookup command parser.
func (p Parser) Lookup() {
	progName := p.Args[0]
	if len(p.Args) < 4 {
		usage(progName)
	}
	filterFile := p.Args[2]
	files := p.Args[3:]
	if !readableFile(filterFile) {
		fmt.Printf("[-] Unable to open %s for reading\n", filterFile)
		usage(progName)
	}
	bloomFilter := NewBloomFilter(1, 0.01)
	bloomFilter.Load(filterFile)
	for _, file := range files {
		bloomFilter.LookupHashes(file)
	}
}

// Parse command line args.
func (p Parser) Parse() {
	progName := p.Args[0]
	if len(p.Args) < 3 {
		usage(progName)
	}
	command := os.Args[1]
	switch command {
	case "calculate":
		p.Calculate()
	case "filters":
		p.Filters()
	case "fromfile":
		p.FromFile()
	case "lookup":
		p.Lookup()
	default:
		fmt.Printf("Invalid command: %s\n", command)
		usage(progName)
	}
}
