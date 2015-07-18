package main

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type TestData struct {
	Have string
	Want string
}

var urls = []TestData{
	{"www.ge.com", "ge*.com"},
	{"ge.com", "ge*.com"},
	{"www.apple.com", "apple*.com"},
	{"foo.bar.github.com", "github*.com"},
}

func TestBreakUpURL(t *testing.T) {
	for _, data := range urls {
		have := data.Have
		want := data.Want

		got := makeURLPattern(have)
		t.Fatalf("Have: %v Want: %v Got: %v", have, want, got)
		if got != want {
			t.Fatalf("Have: %v Want: %v Got: %v", have, want, got)
			return
		}

	}

}
