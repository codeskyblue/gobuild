// use martini as web framework
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"launchpad.net/goyaml"

	"code.google.com/p/go.net/websocket"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/qiniu/api/conf"
	"github.com/shxsun/gobuild/models"
	"github.com/shxsun/gobuild/utils"
	"github.com/shxsun/klog"
)

var (
	mu         = &sync.Mutex{}
	broadcasts = make(map[string]*utils.WriteBroadcaster)
	totalUser  = 0

	lg = klog.DevLog

	OS   = []string{"windows", "linux", "darwin"}
	Arch = []string{"386", "amd64"}
)

type Project struct {
	BufferStr string
	Reader    io.ReadCloser
	job       *Job
}

func (p *Project) Close() {
	p.Reader.Close()
}

func NewProject(addr, name string) *Project {
	mu.Lock()
	defer mu.Unlock()
	var wc *utils.WriteBroadcaster
	if wc = broadcasts[addr]; wc == nil {
		wc = utils.NewWriteBroadcaster()
		broadcasts[addr] = wc

		// start compiling job
		go func() {
			_, err := NewJob(addr, "-", "linux", "amd64", wc).Auto()
			if err != nil {
				lg.Error(err)
			}
		}()
	}

	bufbytes, rd := wc.NewReader(name)
	reader := utils.NewBufReader(rd)
	return &Project{
		BufferStr: string(bufbytes),
		Reader:    reader,
	}
}

type Message struct {
	Error error  `json:"error"`
	Type  string `json:"type"` // FIXME: how to omited
	Data  string `json:"data"`
}

func WsBuildServer(ws *websocket.Conn) {
	defer ws.Close()
	var err error
	clientMsg := new(Message)
	if err = websocket.JSON.Receive(ws, &clientMsg); err != nil {
		utils.Debugf("read json error: %v", err)
		return
	}
	addr := clientMsg.Data
	name := ws.RemoteAddr().String()
	lg.Debug(addr, name)

	proj := NewProject(addr, name)
	defer proj.Close()

	firstMsg := &Message{
		Data: proj.BufferStr,
	}
	err = websocket.JSON.Send(ws, firstMsg)
	if err != nil {
		utils.Debugf("send first msg error: %v", err)
		return
	}

	// send the rest outputs
	buf := make([]byte, 100)
	msg := new(Message)
	for {
		n, err := proj.Reader.Read(buf)
		if n > 0 {
			msg.Data = string(buf[:n])
			deadline := time.Now().Add(time.Second * 1)
			ws.SetWriteDeadline(deadline)
			if er := websocket.JSON.Send(ws, msg); er != nil {
				lg.Debug("write failed timeout, user logout")
				return
			}
		}
		if err != nil {
			return
		}
	}
}

type Configuration struct {
	Hostname   string `yaml:"hostname"`
	ListenAddr string `yaml:"listen"`
	GOROOT     string `yaml:"goroot"`
	Driver     string `yaml:"driver"`
	DataSource string `yaml:"data_source"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
}

type config struct {
	Development Configuration `yaml:"development"`
	Production  Configuration `yaml:"production"`
}

var (
	m           = martini.Classic()
	environment = flag.String("e", "development", "select environment <development|production>")
	opts        *Configuration
)

func init() {
	cfg := new(config)
	in, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		lg.Fatal(err)
	}
	err = goyaml.Unmarshal(in, cfg)
	if err != nil {
		lg.Fatal(err)
	}

	flag.Parse()

	if *environment == "development" {
		opts = &cfg.Development
	} else {
		opts = &cfg.Production
	}

	fmt.Println("== environment:", *environment, "==")
	utils.Dump(opts)

	conf.ACCESS_KEY = opts.AccessKey
	conf.SECRET_KEY = opts.SecretKey

	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))
	InitRouter()
}

func main() {
	var err error
	err = models.InitDB(opts.Driver, opts.DataSource)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Info("gobuild service stated ...")

	http.Handle("/", m)
	http.Handle("/websocket", websocket.Handler(WsBuildServer))

	if err = http.ListenAndServe(opts.ListenAddr, nil); err != nil {
		lg.Fatal(err)
	}
}
