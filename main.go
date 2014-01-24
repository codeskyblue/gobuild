// use martini as web framework
package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/jessevdk/go-flags"
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

func startCommand(wr *utils.WriteBroadcaster, arg0 string, args ...string) {
	// start to run build command
	cmd := exec.Command(arg0, args...)
	cmd.Stdout = wr
	cmd.Stderr = wr
	go func() {
		err := cmd.Run()
		if err != nil {
			utils.Debugf("start cmd error: %v", err)
			io.WriteString(wr, "\nERROR: "+err.Error())
		}
		wr.CloseWriters()
	}()
}

type Project struct {
	Channel   chan string
	BufferStr string
	Reader    io.ReadCloser
}

func (p *Project) Close() {
	p.Reader.Close()
}

func NewProject(addr, name string) *Project {
	mu.Lock()
	defer mu.Unlock()
	if broadcasts[addr] == nil {
		writer := utils.NewWriteBroadcaster()
		broadcasts[addr] = writer
		startCommand(writer, "./autobuild", addr)
	}
	bc := broadcasts[addr]

	utils.Debugf("%s: lock 2.2 new reader done", name)
	bufbytes, rd := bc.NewReader(name)
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
	lg.Debug("handle request project:", addr, name)

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
	lg.Debug(addr, "loop ends")
}

var (
	opts struct {
		ConfigFile string `short:"f" long:"file" description:"configuration file" default:"app.ini"`

		ListenAddr string `short:"l" long:"listen" description:"server listen address" default:":3000"`
		Hostname   string `short:"H" long:"host" description:"hostname like gobuild.io" default:"localhost"`

		Server   string `short:"s" long:"serverAddr"`
		WsServer string `short:"w" long:"wsAddr"`
		CDN      string `short:"c" long:"cdn"`
	}
	args []string

	listenAddr = ""
)

func parseConfig() (err error) {
	parser := flags.NewParser(&opts, flags.Default)
	args, err = flags.Parse(&opts)
	if err != nil {
		return
	}
	err = flags.NewIniParser(parser).ParseFile(opts.ConfigFile)
	if err != nil {
		return
	}

	// change to localhost:port
	if opts.Hostname == "localhost" {
		opts.Hostname += opts.ListenAddr[strings.Index(opts.ListenAddr, ":"):]
	}

	// FIXME: below code need to be deleted
	if opts.CDN == "" {
		opts.CDN = "http://" + opts.Server + "/files"
	}
	if opts.WsServer == "" {
		opts.WsServer = "ws://" + opts.Server
	}
	sepIndex := strings.Index(opts.Server, ":")
	listenAddr = opts.Server[sepIndex:]
	return err
}

var m = martini.Classic()

func init() {
	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))
	InitRouter()
}

func main() {
	var err error
	if err = parseConfig(); err != nil {
		return
	}
	lg.Info("gobuild service stated ...")
	fmt.Println("\tlisten address:", listenAddr)
	fmt.Println("\twebsocket addr:", opts.WsServer)
	fmt.Println("\tCDN:", opts.CDN)

	http.Handle("/", m)
	http.Handle("/websocket", websocket.Handler(WsBuildServer))

	if err = http.ListenAndServe(listenAddr, nil); err != nil {
		lg.Fatal(err)
	}
}
