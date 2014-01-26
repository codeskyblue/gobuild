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
	Package("gobuildrc")
}
