package main

import "testing"

var sample_yaml = `---
filesets:
    includes:
        - public
        - README.*
    excludes:
        - .svn
`

func TestPackage(t *testing.T) {
	addr, err := Package([]string{"README.md"}, "gobuildrc")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr)
}
