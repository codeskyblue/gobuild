package main

import "testing"

func TestPackage(t *testing.T) {
	addr, err := Package([]string{"README.md"}, "public/gobuildrc")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr)
}
