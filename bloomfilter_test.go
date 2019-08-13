package main

import (
	"log"
	"math"
	"testing"

	"github.com/spf13/afero"
)

func TestAccuracy(t *testing.T) {
	var items int32 = 1000
	var fpRate = 0.1
	var fs = afero.NewMemMapFs()
	bloomFilter := NewBloomFilter(items, fpRate, fs)
	size := bloomFilter.Size
	hashCount := bloomFilter.HashCount
	expected := (100 - (100 * fpRate))
	accuracy := math.Round(Accuracy(size, hashCount, items))
	if accuracy != expected {
		t.Errorf(
			"BloomFilter: Accuracy: expected: %f actual: %f",
			expected,
			accuracy,
		)
	}
}

func TestExpectedSizes(t *testing.T) {
	var expectedItems int32 = 3
	var fpRate = 0.01
	var expectedHashCount int32 = 6
	var expectedSize int32 = 28
	var expectedByteSize int32 = 4
	var fs = afero.NewMemMapFs()
	expectedByteSizeHuman := "4.0bytes"
	bloomFilter := NewBloomFilter(expectedItems, fpRate, fs)
	if bloomFilter.HashCount != expectedHashCount {
		t.Errorf(
			"BloomFilter: HashCount: expected: %d actual: %d",
			expectedHashCount,
			bloomFilter.HashCount,
		)
	}
	if bloomFilter.Size != expectedSize {
		t.Errorf(
			"BloomFilter: Size: expected: %d actual: %d",
			expectedSize,
			bloomFilter.Size,
		)
	}
	if bloomFilter.ByteSize != expectedByteSize {
		t.Errorf(
			"BloomFilter: ByteSize: expected: %d actual: %d",
			expectedByteSize,
			bloomFilter.ByteSize,
		)
	}
	if bloomFilter.ByteSizeHuman != expectedByteSizeHuman {
		t.Errorf(
			"BloomFilter: ByteSizeHuman: expected: %s actual: %s",
			expectedByteSizeHuman,
			bloomFilter.ByteSizeHuman,
		)
	}
}

func TestAddAndLookup(t *testing.T) {
	var items int32 = 5
	var fpRate = 0.1
	var fs = afero.NewMemMapFs()
	bloomFilter := NewBloomFilter(items, fpRate, fs)
	filePath1 := "/var/data/xx1.txt"
	filePath2 := "/var/data/xx2.txt"
	fs.MkdirAll("/var/data", 0755)
	f1, err1 := fs.Create(filePath1)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer f1.Close()
	f1.WriteString("file1")

	f2, err2 := fs.Create(filePath2)
	if err2 != nil {
		log.Fatal(err2)
	}
	defer f2.Close()
	f2.WriteString("file2")

	md5File1 := md5File(filePath1, fs)
	md5File2 := md5File(filePath2, fs)
	if bloomFilter.Lookup(md5File1) {
		t.Errorf(
			"BloomFilter: Lookup: expected not to find: %s",
			md5File1,
		)
	}
	if bloomFilter.Lookup(md5File2) {
		t.Errorf(
			"BloomFilter: Lookup: expected not to find: %s",
			md5File2,
		)
	}

	bloomFilter.Add(md5File1)
	if !bloomFilter.Lookup(md5File1) {
		t.Errorf(
			"BloomFilter: Lookup: expected to find: %s",
			md5File1,
		)
	}
	if bloomFilter.Lookup(md5File2) {
		t.Errorf(
			"BloomFilter: Lookup: expected not to find: %s",
			md5File2,
		)
	}
}

func TestSaveAndLoad(t *testing.T) {
	var items int32 = 3
	var fpRate = 0.01
	var expectedSize int32 = 28
	var fs = afero.NewMemMapFs()
	bloomFilter := NewBloomFilter(items, fpRate, fs)
	fakeDir := "/var/data"
	fakePath := fakeDir + "test_filter"
	fs.MkdirAll(fakeDir, 0755)

	if bloomFilter.Size != expectedSize {
		t.Errorf(
			"BloomFilter: Size: expected: %d actual: %d",
			bloomFilter.Size,
			expectedSize,
		)
	}

	bloomFilter.Save(fakePath)

	var newItems int32 = 5
	var newFPRate = 0.02
	var newExpectedSize int32 = 40
	newBloomFilter := NewBloomFilter(newItems, newFPRate, fs)
	if newBloomFilter.Size != newExpectedSize {
		t.Errorf(
			"New BloomFilter: Size before load: expected: %d actual: %d",
			newBloomFilter.Size,
			newExpectedSize,
		)
	}

	// Should be same size as original filter
	newBloomFilter.Load(fakePath)
	if newBloomFilter.Size != expectedSize {
		t.Errorf(
			"New BloomFilter: Size after load: expected: %d actual: %d",
			newBloomFilter.Size,
			expectedSize,
		)
	}
}

func TestCalculateAndLookupHashes(t *testing.T) {
	var items int32 = 1
	var fpRate = 0.01
	var fs = afero.NewMemMapFs()
	bloomFilter := NewBloomFilter(items, fpRate, fs)
	fakeDir := "/var/data/"
	fs.MkdirAll(fakeDir, 0755)

	filePath1 := fakeDir + "file1.txt"
	f1, err1 := fs.Create(filePath1)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer f1.Close()
	f1.WriteString("file1")

	filePath2 := fakeDir + "file2.txt"
	f2, err2 := fs.Create(filePath2)
	if err2 != nil {
		log.Fatal(err2)
	}
	defer f2.Close()
	f2.WriteString("file2")

	filePath3 := fakeDir + "file3.txt"
	f3, err3 := fs.Create(filePath3)
	if err3 != nil {
		log.Fatal(err3)
	}
	defer f3.Close()
	f3.WriteString("file3")

	fakeSubDir := "/var/data/files/"
	fs.MkdirAll(fakeSubDir, 0755)

	filePath4 := fakeSubDir + "file4.txt"
	f4, err4 := fs.Create(filePath4)
	if err4 != nil {
		log.Fatal(err4)
	}
	defer f4.Close()
	f4.WriteString("file4")

	filePath5 := fakeSubDir + "file5.txt"
	f5, err5 := fs.Create(filePath5)
	if err5 != nil {
		log.Fatal(err5)
	}
	defer f5.Close()
	f5.WriteString("file5")

	filePath6 := fakeSubDir + "file6.txt"
	f6, err6 := fs.Create(filePath6)
	if err6 != nil {
		log.Fatal(err6)
	}
	defer f6.Close()
	f6.WriteString("file6")

	fakeEmptyDir := "/var/data/empty/"
	fs.MkdirAll(fakeEmptyDir, 0755)

	bloomFilter.CalculateHashes(fakeDir)

	// Try a specific file, too.
	bloomFilter.CalculateHashes(filePath1)
	fs.Chmod(filePath1, 0111)
	bloomFilter.CalculateHashes(filePath1)

	// Make a file that isn't in the filter.
	bloomFilter.LookupHashes(fakeDir)
	filePath7 := fakeSubDir + "file7.txt"
	f7, err7 := fs.Create(filePath7)
	if err7 != nil {
		log.Fatal(err7)
	}
	defer f7.Close()
	f7.WriteString("file7")
	bloomFilter.LookupHashes(filePath7)
	fs.Chmod(filePath7, 0111)
	bloomFilter.LookupHashes(filePath7)
}
