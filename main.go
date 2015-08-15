// use martini as web framework
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codeskyblue/go-websocket"
	"github.com/codeskyblue/gobuild/database"
	"github.com/codeskyblue/gobuild/utils"
	"github.com/codeskyblue/goyaml"
	"github.com/qiniu/api.v6/conf"
	"github.com/qiniu/log"
)

var (
	mu         = &sync.Mutex{}
	broadcasts = make(map[string]*utils.WriteBroadcaster)
	totalUser  = 0

	//log = klog.DevLog
	//OS   = []string{"windows", "linux", "darwin"}
	Arch = []string{"386", "amd64"}
)

type StreamOutput struct {
	BufferStr string
	Reader    io.ReadCloser
	job       *Builder
}

func (p *StreamOutput) Close() {
	p.Reader.Close()
}

func GetWriteBroadcaster(project string) (wc *utils.WriteBroadcaster, newer bool) {
	mu.Lock()
	defer mu.Unlock()
	if wc = broadcasts[project]; wc == nil {
		wc = utils.NewWriteBroadcaster()
		broadcasts[project] = wc
		newer = true
	}
	return
}

func NewStreamOutput(project, branch, goos, goarch string) *StreamOutput {
	fullname := strings.Join([]string{project, branch, goos, goarch}, "-")
	wb, newer := GetWriteBroadcaster(fullname)
	if newer {
		go func() {
			// start compiling job
			_, err := NewBuilder(project, branch, goos, goarch, wb).Auto()
			if err != nil {
				log.Error(err)
			}
			mu.Lock()
			defer mu.Unlock()
			delete(broadcasts, fullname)
			log.Info("delete broadcasts", project, broadcasts)
		}()
	}

	bufbytes, rd := wb.NewReader("")
	reader := utils.NewBufReader(rd)
	return &StreamOutput{
		BufferStr: string(bufbytes),
		Reader:    reader,
	}
}

type SendMsg struct {
	Error error  `json:"error"`
	Type  string `json:"type"` // FIXME: how to omited
	Data  string `json:"data"`
}

type RecvMsg struct {
	Project string `json:"project"`
	Branch  string `json:"branch"`
	GOOS    string `json:"goos"`
	GOARCH  string `json:"goarch"`
}

// output websocket
func WsBuildServer(ws *websocket.Conn) {
	defer ws.Close()
	var err error
	recvMsg := new(RecvMsg)
	sendMsg := new(SendMsg)
	err = websocket.JSON.Receive(ws, &recvMsg)
	if err != nil {
		sendMsg.Error = err
		websocket.JSON.Send(ws, sendMsg)
		utils.Debugf("read json error: %v", err)
		return
	}

	name := ws.RemoteAddr().String()
	log.Debug(name)

	sout := NewStreamOutput(recvMsg.Project, recvMsg.Branch, recvMsg.GOOS, recvMsg.GOARCH)
	defer sout.Close()

	sendMsg.Data = sout.BufferStr
	err = websocket.JSON.Send(ws, sendMsg)
	if err != nil {
		utils.Debugf("send first sendMsg error: %v", err)
		return
	}

	// send the rest outputs
	buf := make([]byte, 100)
	for {
		n, err := sout.Reader.Read(buf)
		if n > 0 {
			sendMsg.Data = string(buf[:n])
			deadline := time.Now().Add(time.Second * 1)
			ws.SetWriteDeadline(deadline)
			if er := websocket.JSON.Send(ws, sendMsg); er != nil {
				log.Debug("write failed timeout, user logout")
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
	secure      = flag.Bool("secure", false, "use secure connection")
	opts        *Configuration
)

func init() {
	cfg := new(config)
	in, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = goyaml.Unmarshal(in, cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	log.SetOutputLevel(log.Ldebug)

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

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	var err error
	err = database.InitDB(opts.Driver, opts.DataSource)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("gobuild service stated ...")

	http.Handle("/", m)
	http.Handle("/websocket/", websocket.Handler(WsBuildServer))
	http.HandleFunc("/hello", HelloServer)

	if *secure {
		go func() {
			er := http.ListenAndServeTLS(":443", "bin/ssl.crt", "bin/ssl.key", nil)
			if er != nil {
				log.Error(er)
			}
		}()
	}
	err = http.ListenAndServe(opts.ListenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
