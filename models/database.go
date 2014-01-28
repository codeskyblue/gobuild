package models

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/astaxie/beego/utils"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shxsun/klog"
)

var (
	dbName = filepath.Join(utils.SelfDir(), "./sqlite.db")
	x      *xorm.Engine
	lg     = klog.DevLog
)

type Project struct {
	Name      string //`xorm:"unique(p)"`
	Uuid      string //`xorm:"unique(p)"`
	Ref       string
	Mtime     string
	Address   string
	ViewCount int

	Project string `xorm:"pk"`
	Sha     string
}

func SyncProject(l *Project) error {
	if l == nil {
		return errors.New("sync nil")
	}
	affec, err := x.Where("project=?", l.Project).Update(l)
	if err != nil {
		return err
	}
	if affec == 0 {
		if _, err := x.InsertOne(l); err != nil {
			return err
		}
	}
	return nil
}

func GetSha(project string) (sha string, err error) {
	l := new(Project)
	ok, err := x.Where("project=?", project).Get(l)
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
	x, err = xorm.NewEngine("sqlite3", dbName)
	if err != nil {
		lg.Fatal(err)
	}
	if err = x.Sync(new(Project)); err != nil {
		lg.Fatal(err)
	}
}
