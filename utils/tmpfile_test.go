package utils

import (
	"os"
	"testing"
)

func TestTempFile(t *testing.T) {
	name, err := TempFile("./", "tmp-", ".zip")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(name)
	err = os.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}
