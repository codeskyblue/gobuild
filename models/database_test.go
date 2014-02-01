package models

import "testing"

/*
func TestSyncProject(t *testing.T) {
	l := &Project{
		Name: "pp",
		Sha:  strconv.Itoa(rand.Int()),
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
*/

func TestAddProject(t *testing.T) {
	id, err := AddProject("t1", "v1.1", "asdfqwer")
	if err != nil {
		t.Error(err)
	}
	if id <= 0 {
		t.Errorf("id:%d should not be <= 0", id)
	}
	t.Log("id =", id)
	// clean test env
	defer func() {
		affec, err := x.Delete(&Project{Name: "t1"})
		if err != nil {
			t.Error(err)
		}
		_ = affec
		t.Log(affec)
	}()

	_, err = SearchProject("t1", "xxx")
	if err == nil {
		t.Error("search sha:xxx should return nil, but got something")
	}

	p, err := SearchProject("t1", "asdfqwer")
	if err != nil {
		t.Error(err)
	}
	if p == nil ||
		p.ViewCount != 0 ||
		p.Name != "t1" ||
		p.Sha != "asdfqwer" ||
		p.Ref != "v1.1" {
		t.Errorf("project should not like that: %#v", p)
	}
}

func TestAddFile(t *testing.T) {
	var err error
	id, _ := AddProject("t2", "v1.2", "asdfqwer")
	// clean test env
	defer func() {
		x.Delete(&Project{Name: "t2"})
	}()
	err = AddFile(id, "linux-amd64", "http://", "...")
	if err != nil {
		t.Error(err)
	}

	_, err = SearchFile(id, "xksjf")
	if err == nil {
		t.Error("search this file should be empty")
	}
	f, err := SearchFile(id, "linux-amd64")
	if err != nil {
		t.Error(err)
	}
	if f == nil || f.ProjId != id || f.Tag != "linux-amd64" {
		t.Errorf("should not like that: %#v", f)
	}
}
