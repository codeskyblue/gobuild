package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var (
	x *xorm.Engine
)

// should be call before pacakge called
func InitDB(driver, dataSource string) (err error) {
	x, err = xorm.NewEngine(driver, dataSource)
	if err != nil {
		return
	}
	// create tables
	return x.Sync(new(Project), new(File))
}

type Project struct {
	Id        int64  `xorm:"pk autoincr"`
	Name      string `xorm:"unique(nr)"`
	Sha       string `xorm:"unique(nr)"`
	Ref       string
	Time      time.Time `xorm:"created"`
	ViewCount int
}

type File struct {
	Id        int64 `xorm:"pk autoincr"`
	ProjId    int64
	Tag       string
	Addr      string
	Log       string    `xorm:"text"`
	Time      time.Time `xorm:"created"`
	ViewCount int
	Version   int `xorm:"version"`
}

func SearchProject(name, sha string) (p *Project, err error) {
	p = new(Project)
	p.Name = name
	p.Sha = sha
	ok, err := x.Get(p)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("project:%s not found", name)
		return
	}
	return
}

func AddProject(name, ref, sha string) (id int64, err error) {
	p := &Project{
		Name: name,
		Ref:  ref,
		Sha:  sha,
	}
	affec, err := x.InsertOne(p)
	if err != nil {
		return
	}
	if affec != 1 {
		err = fmt.Errorf("insert record(%s,%s,%s) failed", name, ref, sha)
		return
	}
	sp, err := SearchProject(name, sha)
	if err != nil {
		return
	}
	return sp.Id, nil
}

func SearchFile(projID int64, tag string) (f *File, err error) {
	f = new(File)
	f.ProjId = projID
	f.Tag = tag
	ok, err := x.Get(f)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("not found: (%d, %s)", projID, tag)
	}
	return
}

func AddFile(projId int64, tag, addr, log string) (err error) {
	f := File{
		ProjId: projId,
		Tag:    tag,
		Addr:   addr,
		Log:    log,
	}
	affec, err := x.InsertOne(f)
	if err != nil {
		return
	}
	if affec != 1 {
		err = fmt.Errorf("AddFile(%d, %s) failed", projId, tag)
		return
	}
	return
}

/*
func SyncProject(l *Project) error {
	if l == nil {
		return errors.New("sync nil")
	}
	affec, err := x.Where("project=?", l.Name).Update(l)
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
*/
