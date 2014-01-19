package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/astaxie/beego/utils"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbName = filepath.Join(utils.SelfDir(), "./sqlite.db")
	//dbName = "./sqlite.db"
	Engine *xorm.Engine
)

type Latest struct {
	Project string `xorm:"pk"`
	Sha     string
}

func SyncProject(l *Latest) error {
	if l == nil {
		return errors.New("sync nil")
	}
	affec, err := Engine.Where("project=?", l.Project).Update(l)
	if err != nil {
		return err
	}
	if affec == 0 {
		if _, err := Engine.InsertOne(l); err != nil {
			return err
		}
	}
	return nil
}

func GetSha(project string) (sha string, err error) {
	l := new(Latest)
	ok, err := Engine.Where("project=?", project).Get(l)
	if err != nil {
		return
	}
	if !ok {
		return "", fmt.Errorf("query %s, not found", project)
	}
	return l.Sha, nil
}

func init() {
	var err error
	Engine, err = xorm.NewEngine("sqlite3", dbName)
	if err != nil {
		lg.Fatal(err)
	}
	if err = Engine.Sync(new(Latest)); err != nil {
		lg.Fatal(err)
	}
}
