package xxh

import (
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

const lipsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec id magna ipsum. Integer mattis vehicula aliquam. Ut felis nunc, venenatis nec aliquet vitae, gravida et massa. Donec sollicitudin, dui non dignissim dictum, elit ligula auctor massa, non maximus metus tellus id nisl. Curabitur porttitor dignissim lacinia. In mattis dolor sed molestie porttitor. Nam ornare, leo at blandit elementum, sem metus faucibus sem, vitae pharetra ligula erat fermentum leo. Suspendisse sollicitudin, lorem et faucibus tincidunt, lorem mi pretium urna, ac faucibus enim ipsum sed est. Suspendisse congue odio turpis. Suspendisse ac volutpat lacus. Integer luctus viverra hendrerit. Cras a justo risus.\n"

var data32 = []struct {
	Want  uint32
	Value string
}{
	{Value: lipsum, Want: 0x64ef824d}, // xxh32sum 0.6.5
	{Value: "abcd", Want: 0xA3643705},
	{Value: "abc", Want: 0x32D153FF},
	{Value: "hello world", Want: 0xCEBB6622},
	{Value: "I love programming in GO!", Want: 0x4FE3561F},
	{Value: strings.Repeat("abc", 999), Want: 0x89DA9B6E},
	{Value: strings.Repeat("abcd", 1000), Want: 0xE18CBEA},
	{Value: "the quick brown fox jumps over the lazy dog", Want: 0x66716377},
}

var data64 = []struct {
	Want  uint64
	Value string
}{
	{Value: lipsum, Want: 0x59f3208ca1d7b1b4},
	{Value: "the quick brown fox jumps over the lazy dog", Want: 0xed714233c5a9a792},
}

func TestHashSum64(t *testing.T) {
	data := []struct {
		Value string
		Want  uint64
	}{
		{Value: "the quick brown fox", Want: 0x150018d41c31b193},
		{Value: " jumps over the", Want: 0xb828d547a6bd6d1e},
		{Value: " lazy dog", Want: 0xed714233c5a9a792},
	}
	digest := New64(0)
	for i, d := range data {
		digest.Write([]byte(d.Value))
		hash := binary.BigEndian.Uint64(digest.Sum(nil))
		if hash != d.Want {
			t.Errorf("%d) hash mismatched! want %x, got %x", i+1, d.Want, hash)
			break
		}
	}
}

func TestHashSum32(t *testing.T) {
	data := []struct {
		Value string
		Want  uint32
	}{
		{Value: "the quick brown fox", Want: 0x9adf0164},
		{Value: " jumps over the", Want: 0x3e6a06dd},
		{Value: " lazy dog", Want: 0x66716377},
	}
	digest := New32(0)
	for i, d := range data {
		digest.Write([]byte(d.Value))
		hash := binary.BigEndian.Uint32(digest.Sum(nil))
		if hash != d.Want {
			t.Errorf("%d) hash mismatched! want %x, got %x", i+1, d.Want, hash)
			break
		}
	}
}

func TestHash32(t *testing.T) {
	for i, d := range data32 {
		w := New32(0)
		if _, err := io.Copy(w, strings.NewReader(d.Value)); err != nil {
			t.Errorf("test %d failed: %s", i+1, err)
			continue
		}
		if got := w.Sum32(); d.Want != got {
			t.Errorf("test %d failed (len: %d): want %x, got %x", i+1, len(d.Value), d.Want, got)
		}
	}
}

func TestSum32(t *testing.T) {
	for i, d := range data32 {
		got := Sum32([]byte(d.Value), 0)
		if got != d.Want {
			t.Errorf("test %d failed (len: %d): want %x, got %x", i+1, len(d.Value), d.Want, got)
		}
	}
}

func TestSum64(t *testing.T) {
	for i, d := range data64 {
		got := Sum64([]byte(d.Value), 0)
		if got != d.Want {
			t.Errorf("test %d failed (len: %d): want %x, got %x", i+1, len(d.Value), d.Want, got)
		}
	}
}

func BenchmarkNew64(b *testing.B) {
	str := []byte(lipsum)
	for n := 0; n < b.N; n++ {
		h := New64(0)
		for i := 0; i < 1000; i++ {
			h.Write(str)
		}
		h.Sum64()
	}
}

func BenchmarkNew32(b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := New32(0)
		for i := 0; i < 1000; i++ {
			io.WriteString(h, lipsum)
		}
		h.Sum32()
	}
}
