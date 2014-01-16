package main

import (
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbName = "./sqlite.db"
	Engine *xorm.Engine
)

type Latest struct {
	Project string `xorm:"pk"`
	Sha     string
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
