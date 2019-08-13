package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestCountFiles(t *testing.T) {
	fakeDir := "/var/data/"
	var fs = afero.NewMemMapFs()
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

	var expectedDirCount int32 = 6
	dirCount := countFiles(fakeDir, fs)
	if dirCount != expectedDirCount {
		t.Errorf(
			"countFiles in directory: expected: %d actual: %d",
			expectedDirCount,
			dirCount,
		)
	}

	var expectedFileCount int32 = 1
	fileCount := countFiles(filePath1, fs)
	if fileCount != expectedFileCount {
		t.Errorf(
			"countFiles for single file: expected: %d actual: %d",
			expectedFileCount,
			fileCount,
		)
	}
}

func TestIsMD5(t *testing.T) {
	data := []byte("money money money money money")
	md5Hex := fmt.Sprintf("%x", md5.Sum(data))
	if !isMD5(md5Hex) {
		t.Errorf(
			"isMD5: %s expected: %t actual: %t",
			md5Hex,
			true,
			isMD5(md5Hex),
		)
	}

	notMD5Hex := strings.Repeat("x", 32)
	if isMD5(notMD5Hex) {
		t.Errorf(
			"isMD5: %s expected: %t actual: %t",
			notMD5Hex,
			false,
			isMD5(notMD5Hex),
		)
	}

	fakeDir := "/var/data/"
	var fs = afero.NewMemMapFs()
	fs.MkdirAll(fakeDir, 0755)

	filePath := fakeDir + "file.txt"
	f, err := fs.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.WriteString("file")

	fileHash := md5File(filePath, fs)
	if !isMD5(fileHash) {
		t.Errorf(
			"isMD5 for file: %s expected: %t actual: %t",
			filePath,
			true,
			isMD5(fileHash),
		)
	}
}

func TestNoPermission(t *testing.T) {
	fakeDir := "/var/data/"
	var fs = afero.NewMemMapFs()
	fs.MkdirAll(fakeDir, 0755)

	filePath := fakeDir + "file.txt"
	f, err := fs.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	contents := strings.Repeat("x", 8193)
	f.WriteString(contents)
	if !readableFile(filePath, fs) {
		t.Errorf(
			"readableFile: %s expected: %t actual: %t",
			filePath,
			true,
			readableFile(filePath, fs),
		)
	}
	if !writeableFile(filePath, fs) {
		t.Errorf(
			"writeableFile: %s expected: %t actual: %t",
			filePath,
			true,
			writeableFile(filePath, fs),
		)
	}

	fakeSubDir := "/var/data/files/"
	fs.MkdirAll(fakeSubDir, 0111)
	filePath = fakeSubDir + "file.txt"
	if readableFile(filePath, fs) {
		t.Errorf(
			"readableFile: %s expected: %t actual: %t",
			filePath,
			false,
			readableFile(filePath, fs),
		)
	}
	/* Seems like Afero doesn't throw error on Create when it should?
	if writeableFile(filePath, fs) {
		t.Errorf(
			"writeableFile: %s expected: %t actual: %t",
			filePath,
			false,
			writeableFile(filePath, fs),
		)
	}
	*/
}

func TestCalculateAndLookup(t *testing.T) {
	var fs = afero.NewMemMapFs()
	fakeDir := "/var/data/"
	fs.MkdirAll(fakeDir, 0755)
	filePath := fakeDir + "file.txt"
	f, err := fs.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString("file")
	f.Close()

	fakeFilterDir := "/tmp/filters/"
	fs.MkdirAll(fakeFilterDir, 0755)
	args := []string{"mdd", "calculate", fakeFilterDir + "filterfile", filePath}
	parser := Parser{Args: args, Fs: fs}
	parser.Calculate()

	lookupArgs := []string{"mdd", "lookup", fakeFilterDir + "filterfile", filePath}
	parser = Parser{Args: lookupArgs, Fs: fs}
	parser.Lookup()
}
