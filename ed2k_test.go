package ed2k

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
)

// A "fake" reader that never writes anything to the []byte.
// Effectively always reads len(p) of NULs.
type FakeReader struct{}

func (_ *FakeReader) Read(p []byte) (n int, err error) {
	return len(p), nil
}

type testVector struct {
	Mode bool
	Data io.Reader
	Size int64
	Hash string
}

const chunkSize = BlockSize

var fakeReader = &FakeReader{}

func Test(t *testing.T) {
	for name, vec := range map[string]testVector{
		"empty string hash":                              {Mode: false, Data: nil, Size: 0, Hash: "31d6cfe0d16ae931b73c59d7e0c089c0"},
		"\"small example\" hash":                         {Mode: false, Data: strings.NewReader("small example"), Size: -1, Hash: "3e01197bc54364cb86a41738b06ae679"},
		"new blockchain finishing algorihtm, one block":  {Mode: true, Data: fakeReader, Size: chunkSize, Hash: "fc21d9af828f92a8df64beac3357425d"},
		"new blockchain finishing algorihtm, two blocks": {Mode: true, Data: fakeReader, Size: 2 * chunkSize, Hash: "114b21c63a74b6ca922291a11177dd5c"},
		"old blockchain finishing algorihtm, one block":  {Mode: false, Data: fakeReader, Size: chunkSize, Hash: "d7def262a127cd79096a108e7a9fc138"},
		"old blockchain finishing algorihtm, two blocks": {Mode: false, Data: fakeReader, Size: 2 * chunkSize, Hash: "194ee9e4fa79b2ee9f8829284c466051"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ed2k := New(vec.Mode)
			if vec.Size == 0 { // do nothing
			} else if vec.Size < 0 {
				_, err := io.Copy(ed2k, vec.Data)
				if err != nil {
					t.Error(err)
				}
			} else {
				_, err := io.CopyN(ed2k, vec.Data, vec.Size)
				if err != nil {
					t.Error(err)
				}
			}
			if got := ed2k.(fmt.Stringer).String(); got != vec.Hash {
				t.Errorf("sum() = %q, expected %q", got, vec.Hash)
			}
		})
	}
}

func Example_hexString() {
	e := New(false)
	_, _ = io.Copy(e, strings.NewReader("small example"))
	fmt.Println(hex.EncodeToString(e.Sum(nil)))

	// for convenience, ed2k implements Stringer by doing just that
	fmt.Println(e)
	// Output:
	// 3e01197bc54364cb86a41738b06ae679
	// 3e01197bc54364cb86a41738b06ae679
}

func Example_noNullChunk() {
	e := New(false)
	_, _ = io.Copy(e, strings.NewReader("small example"))
	h := e.Sum(nil)
	fmt.Println(h)
	// Output: [62 1 25 123 197 67 100 203 134 164 23 56 176 106 230 121]
}

func bench(b *testing.B, mode bool, size int64) {
	b.Helper()
	b.SetBytes(size)

	ed2k := New(mode)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed2k.Reset()
		_, err := io.CopyN(ed2k, fakeReader, size)
		if err != nil {
			b.Error(err)
		}
		ed2k.Sum(nil)
	}
}

func Benchmark_nullChunk(b *testing.B) {
	bench(b, true, chunkSize)
}

func Benchmark_noNullChunk(b *testing.B) {
	bench(b, false, chunkSize)
}

func Benchmark_1MB(b *testing.B) {
	bench(b, false, 1*1024*1024)
}

func Benchmark_10MB(b *testing.B) {
	bench(b, false, 10*1024*1024)
}

func Benchmark_100MB(b *testing.B) {
	bench(b, false, 100*1024*1024)
}

func Benchmark_1GB(b *testing.B) {
	bench(b, false, 1*1024*1024*1024)
}

func Benchmark_10GB(b *testing.B) {
	bench(b, false, 10*1024*1024*1024)
}
