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
	bloomFilter := NewBloomFilter(items, fpRate)
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
	expectedByteSizeHuman := "4.0bytes"
	bloomFilter := NewBloomFilter(expectedItems, fpRate)
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
	bloomFilter := NewBloomFilter(items, fpRate)
	filePath1 := "/var/data/xx1.txt"
	filePath2 := "/var/data/xx2.txt"
	var appfs = afero.NewOsFs()
	appfs.MkdirAll("/var/data", 0755)
	f1, err1 := appfs.Create(filePath1)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer f1.Close()
	f1.WriteString("file1")

	f2, err2 := appfs.Create(filePath2)
	if err2 != nil {
		log.Fatal(err2)
	}
	defer f2.Close()
	f2.WriteString("file2")

	md5File1 := md5File(filePath1)
	md5File2 := md5File(filePath2)
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

/*
def test_save_and_load(fs):
    expected_items = 3
    false_positive_rate = 0.01
    expected_size = 28
    fake_dir = '/var/data/'
    fake_path = fake_dir + 'test_filter'
    fs.create_dir(fake_dir)
    bloom_filter = BloomFilter(expected_items, false_positive_rate)
    assert bloom_filter.size == expected_size
    bloom_filter.save(fake_path)

    new_expected_items = 5
    new_false_positive_rate = 0.02
    new_expected_size = 40
    new_bloom_filter = BloomFilter(new_expected_items, new_false_positive_rate)
    assert new_bloom_filter.size == new_expected_size

    # Should be size of original filter
    new_bloom_filter.load(fake_path)
    assert new_bloom_filter.size == expected_size
*/

func TestSaveAndLoad(t *testing.T) {
	var items int32 = 3
	var fpRate = 0.01
	var expectedSize int32 = 28
	bloomFilter := NewBloomFilter(items, fpRate)
	fakeDir := "/var/data"
	fakePath := fakeDir + "test_filter"
	var appfs = afero.NewOsFs()
	appfs.MkdirAll(fakeDir, 0755)

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
	newBloomFilter := NewBloomFilter(newItems, newFPRate)
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
