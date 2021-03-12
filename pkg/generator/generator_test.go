package generator

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// func TestParseType(t *testing.T) {
// 	require.Equal(t, &goType{Name: "string"}, parse("string"))
// 	require.Equal(t, &goType{Name: "Time", ImportPath: "time", ImportName: "time"}, parse("time.Time"))
// 	require.Equal(t, &goType{
// 		Name:       "Foo",
// 		ImportPath: "github.com/vektah/dataloaden/pkg/generator/testdata/mismatch",
// 		ImportName: "mismatched",
// 	}, parse("github.com/vektah/dataloaden/pkg/generator/testdata/mismatch.Foo"))
// }

func parse(s string) *goType {
	t, err := parseType(s)
	if err != nil {
		panic(err)
	}

	return t
}

// func ExamplePkgLoad() {
// 	p, _ := packages.Load(&packages.Config{
// 		Dir: "/home/liuwei/dataloader",
// 	}, ".")

// 	fmt.Println(p[0].Name)
// 	// Output: sss
// 	// aaaa
// }

func BenchmarkF1B(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int64ToBytes1(1111)
	}
}

func BenchmarkF2B(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int64ToBytes2(1111)
	}
}

func BenchmarkF1A(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToInt641([]byte{0x7e})
	}
}

func BenchmarkF2A(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToInt642([]byte{0x7e})
	}
}

func Int64ToBytes1(n int64) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, n)
	return bytebuf.Bytes()
}

func BytesToInt641(bys []byte) int64 {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

func BytesToInt642(bys []byte) int64 {
	data, _ := binary.Varint(bys)
	return data
}

func Int64ToBytes2(data int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, data)
	b := buf[:n]
	return b
}
