package document

import (
	"bytes"
	"compress/gzip"
	"os"
	"testing"
	"unicode/utf8"
)

var testFiles = []string{"../../fixtures/pg10.txt.gz", "../../fixtures/pg1581.txt.gz"}
var fileCache = make(map[string]string)

func openFile(path string) string {
	if len(fileCache[path]) > 0 {
		return fileCache[path]
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	fz, err := gzip.NewReader(f)
	buf := new(bytes.Buffer)
	if err != nil {
		buf.ReadFrom(f)
	} else {
		buf.ReadFrom(fz)
	}
	fileCache[path] = buf.String()
	return fileCache[path]
}

func benchmarkHasher(b *testing.B, hasher HasherFunc, paths []string) {
	b.ResetTimer()
	b.StopTimer()
	key := HashKey{WindowSize: 15, HashWidth: 32}
	byteCount := int64(0)
	for _, path := range paths {
		text := openFile(path)
		count := uint64(utf8.RuneCountInString(text))
		hashes := make([]uint64, count)
		f := func(i int, h uint64) {
			hashes[i] = h
		}
		b.Logf("Benchmarking file %v (%v...)", path, text[:20])
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			hasher(text, count, key, f)
		}
		b.StopTimer()
		byteCount += int64(len(text) * b.N)
	}
	b.SetBytes(byteCount)
}

func Benchmark_RabinKarp(b *testing.B) {
	benchmarkHasher(b, rollingRabinKarp, testFiles)
}

func Benchmark_RabinKarp3(b *testing.B) {
	benchmarkHasher(b, rollingRabinKarp3, testFiles)
}

func testHasher(t *testing.T, hasher HasherFunc) {
	text := "Text gobble! Text"
	count := uint64(utf8.RuneCountInString(text) - 3)
	key := HashKey{WindowSize: 4, HashWidth: 10}
	hashes := make([]uint64, count)
	f := func(i int, h uint64) {
		hashes[i] = h
	}
	hasher(text, count, key, f)
	if len(hashes) != 14 {
		t.Errorf("Wrong number of hashes: %v", hashes)
	}
	if hashes[0] != hashes[13] {
		t.Errorf("Incorrect hashes created: %v %v %v", hashes[0], hashes[13], hashes)
	} else {
		t.Logf("Correct hashes created: %v %v %v", hashes[0], hashes[13], hashes)
	}
}

func Test_Hashers(t *testing.T) {
	testHasher(t, rollingRabinKarp)
	testHasher(t, rollingRabinKarp3)
}
