package main

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestSyncProject(t *testing.T) {
	l := &Latest{
		Project: "pp",
		Sha:     strconv.Itoa(rand.Int()),
	}
	err := SyncProject(l)
	if err != nil {
		t.Error(err)
	}

	sha, err := GetSha("pp")
	if err != nil {
		t.Error(err)
	}
	if l.Sha != sha {
		t.Errorf("expect search result(%s), but got %s", l.Sha, sha)
	}
}
