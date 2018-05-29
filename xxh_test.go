package xxh

import (
	"strings"
	"testing"
)

const lipsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec id magna ipsum. Integer mattis vehicula aliquam. Ut felis nunc, venenatis nec aliquet vitae, gravida et massa. Donec sollicitudin, dui non dignissim dictum, elit ligula auctor massa, non maximus metus tellus id nisl. Curabitur porttitor dignissim lacinia. In mattis dolor sed molestie porttitor. Nam ornare, leo at blandit elementum, sem metus faucibus sem, vitae pharetra ligula erat fermentum leo. Suspendisse sollicitudin, lorem et faucibus tincidunt, lorem mi pretium urna, ac faucibus enim ipsum sed est. Suspendisse congue odio turpis. Suspendisse ac volutpat lacus. Integer luctus viverra hendrerit. Cras a justo risus.\n"

func TestXXH32(t *testing.T) {
	data := []struct {
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
	for i, d := range data {
		v := XXH32([]byte(d.Value), 0)
		if v != d.Want {
			t.Errorf("test %d failed (len: %d): want %x, got %x", i+1, len(d.Value), d.Want, v)
		}
	}
}
