package xxh

import (
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
}

var data64 = []struct {
	Want  uint64
	Value string
}{
	{Value: lipsum, Want: 0x59f3208ca1d7b1b4},
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
