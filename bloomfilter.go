package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/roberson-io/mmh3"
)

// Accuracy calculates a filter's accuracy given
// size, hash count, and expected number
// of elements.
func Accuracy(size, hashCount, elements int) float64 {
	s := float64(size)
	hc := float64(hashCount)
	e := float64(elements)
	fp := math.Pow(1-math.Pow((1-1/s), (hc*e)), hc)
	fmt.Printf("FALSE: %.4f\n", fp)
	return 100 - fp*100
}

func idealSize(expected int, fpRate float64) int {
	return int(-(float64(expected) * math.Log(fpRate)) / math.Pow(math.Log(2.0), 2))
}

func idealHashCount(size, expected int) int {
	return int((float64(size) / float64(expected)) * math.Log(2))
}

func byteSize(size int) int {
	return int(math.Ceil(float64(size) / 8.0))
}

func byteSizeHuman(size int) string {
	suffix := [...]string{
		"bytes", "Kb", "Mb", "Gb", "Tb", "Pb", "Eb", "Zb", "Yb",
	}
	var order int
	if size != 0 {
		order = int(math.Log2(math.Ceil(float64(size))/8.0) / 10)
	}
	human := math.Ceil(float64(size)/8.0) / float64(uint(1)<<uint(order*10))
	return fmt.Sprintf("%.4f%s", human, suffix[order])
}

// NewBloomFilter constructs a Bloom filter given the expected number
// of elements in the Bloom filter and the acceptable rate of false
// positives. For example, 0.01 will tolerate 0.01% chance of false
// positives.
func NewBloomFilter(expectedItems int, fpRate float64) *BloomFilter {
	bf := new(BloomFilter)
	bf.Size = idealSize(expectedItems, fpRate)
	bf.HashCount = idealHashCount(bf.Size, expectedItems)
	var filter BitField
	filter.Bitfield = make([]byte, bf.ByteSize)
	bf.Filter = filter
	bf.ByteSize = byteSize(bf.Size)
	bf.ByteSizeHuman = byteSizeHuman(bf.Size)
	return bf
}

// BloomFilter implements Bloom filter. You should probably use
// NewBloomFilter unless you know what you're doing.
type BloomFilter struct {
	Size          int
	HashCount     int
	Filter        BitField
	ByteSize      int
	ByteSizeHuman string
}

// Add adds an element to the filter.
func (bf *BloomFilter) Add(element string) {
	for seed := 0; seed < bf.HashCount; seed++ {
		key := []byte(element)
		result := int(binary.LittleEndian.Uint32(
			mmh3.Hashx86_32(key, uint32(seed)),
		)) % bf.Size
		bf.Filter.SetBit(result)
	}
}

// Lookup checks if an element exists in the filter.
func (bf *BloomFilter) Lookup(element string) bool {
	for seed := 0; seed < bf.HashCount; seed++ {
		key := []byte(element)
		result := int(binary.LittleEndian.Uint32(
			mmh3.Hashx86_32(key, uint32(seed)),
		)) % bf.Size
		if bf.Filter.GetBit(result) == false {
			return false
		}
	}
	return true
}

// Save saves the filter's current state to a file.
func (bf *BloomFilter) Save(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	size := make([]byte, 16)
	binary.LittleEndian.PutUint64(size, uint64(bf.Size))
	f.Write(size)

	hashCount := make([]byte, 16)
	binary.LittleEndian.PutUint64(hashCount, uint64(bf.HashCount))
	f.Write(hashCount)

	f.Write(bf.Filter.Bitfield)
}

// Load loads a saved filter.
func (bf *BloomFilter) Load(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	sizeBytes := make([]byte, 16)
	_, err = f.Read(sizeBytes)
	if err != nil {
		log.Fatal(err)
	}
	bf.Size = int(binary.LittleEndian.Uint64(sizeBytes))
	bf.ByteSize = byteSize(bf.Size)
	bf.ByteSizeHuman = byteSizeHuman(bf.Size)

	hcBytes := make([]byte, 16)
	_, err = f.Read(hcBytes)
	if err != nil {
		log.Fatal(err)
	}
	bf.HashCount = int(binary.LittleEndian.Uint64(hcBytes))

	bitfield := make([]byte, bf.ByteSize)
	_, err = f.Read(bitfield)
	if err != nil {
		log.Fatal(err)
	}
	bf.Filter.Size = bf.Size
	bf.Filter.Bitfield = bitfield
}

// CalculateHashes calculates MD5 hashes of all files within a
// directory, adding them to a Bloom filter.
func (bf *BloomFilter) CalculateHashes(path string) {
	info, err := os.Stat(path)
	if err != nil {
		if !(os.IsPermission(err)) {
			log.Fatal(err)
		}
	}
	if info.Mode().IsRegular() {
		digest := md5File(path)
		if digest != "" {
			fmt.Printf("  %s    %s\n", path, digest)
		}
		return
	}
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q calculating hashes: %v\n", path, err)
			return err
		}
		// We only care about files.
		if info.Mode().IsRegular() {
			digest := md5File(path)
			if digest != "" {
				fmt.Printf("  %s    %s\n", path, digest)
				bf.Add(digest)
			} else {
				fmt.Print("Permission Denied\n")
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// LookupHashes determines if files within a directory have
// hashes within the Bloom filter.
func (bf *BloomFilter) LookupHashes(path string) {
	info, err := os.Stat(path)
	if err != nil {
		if !(os.IsPermission(err)) {
			log.Fatal(err)
		}
	}
	if info.Mode().IsRegular() {
		digest := md5File(path)
		if digest != "" && !(bf.Lookup(digest)) {
			fmt.Printf("%s is not in filter\n", path)
		}
		return
	}
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q looking up hashes: %v\n", path, err)
			return err
		}
		// We only care about files.
		if info.Mode().IsRegular() {
			digest := md5File(path)
			if !(bf.Lookup(digest)) {
				fmt.Printf("%s is not in filter\n", path)
			} else {
				fmt.Printf("%s is in filter\n", path)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
