// use martini as web framework
package main

import (
	//"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/jessevdk/go-flags"
)

var (
	mu         = &sync.Mutex{}
	broadcasts = make(map[string]*WriteBroadcaster)
	totalUser  = 0

	OS   = []string{"windows", "linux", "darwin"}
	Arch = []string{"386", "amd64"}
)

func startCommand(wr *WriteBroadcaster, arg0 string, args ...string) {
	// start to run build command
	cmd := exec.Command(arg0, args...)
	cmd.Stdout = wr
	cmd.Stderr = wr
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			io.WriteString(wr, "\nERROR: "+err.Error())
		}
		log.Println("done")
		wr.CloseWriters()
	}()
}

type Project struct {
	Channel   chan string
	BufferStr string
	reader    io.ReadCloser
}

func (p *Project) Close() {
	p.reader.Close()
	//close(p.Channel)
}

func NewProject(addr, name string) *Project {
	Debugf("project get lock, %s", name)
	mu.Lock()
	defer func() {
		Debugf("release lock")
		mu.Unlock()
	}()
	Debugf("%s: lock 1 new reader done", name)
	if broadcasts[addr] == nil {
		Debugf("%s: lock 1.1 new reader done", name)
		writer := NewWriteBroadcaster()
		broadcasts[addr] = writer
		Debugf("%s: lock 1.2 new reader done", name)

		startCommand(writer, "./autobuild", addr)
	}
	Debugf("%s: lock 2 new reader done", name)
	bc := broadcasts[addr]
	Debugf("%s: lock 2.1 new reader done", name)

	ch := make(chan string)
	Debugf("%s: new reader req", name)
	Debugf("%s: lock 2.2 new reader done", name)
	bufbytes, rd := bc.NewReader(name)
	Debugf("%s: lock 2.3 new reader done", name)
	Debugf("%s: reader get", name)
	Debugf("%s: lock new reader done", name)
	go func() {
		charBuf := make([]byte, 100)
		defer close(ch)
		for {
			n, err := rd.Read(charBuf)
			if n > 0 {
				ch <- string(charBuf[:n]) // FIXME: if no one read channel, that is a really a problem(but test result it is not a problem), I donot know what `for line := ch does`
			}
			if err != nil {
				return
			}
		}
	}()
	return &Project{
		BufferStr: string(bufbytes),
		Channel:   ch,
		reader:    rd,
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
	log.Println("handle request from:", ws.RemoteAddr())
	clientMsg := new(Message)
	if err = websocket.JSON.Receive(ws, &clientMsg); err != nil {
		log.Println("read json error:", err)
		return
	}
	addr := clientMsg.Data
	name := ws.RemoteAddr().String()
	log.Println("handle request project:", addr, name)

	proj := NewProject(addr, name)
	defer proj.Close()

	log.Println("handle request:", addr, name)

	firstMsg := &Message{
		Data: proj.BufferStr,
	}
	log.Println("send message:", firstMsg)
	err = websocket.JSON.Send(ws, firstMsg)
	if err != nil {
		log.Println("send first msg error:", err)
		return
	}
	log.Println("message sended")
	for line := range proj.Channel {
		//log.Println("send message")
		msg := new(Message)
		msg.Data = line
		deadline := time.Now().Add(time.Second * 1)
		ws.SetWriteDeadline(deadline)
		err := websocket.JSON.Send(ws, msg)
		if err != nil {
			log.Println("write failed timeout, user logout")
			break
		}
	}
	log.Println(addr, "loop ends")
}

var (
	options struct {
		Server   string `short:"s" long:"serverAddr"`
		WsServer string `short:"w" long:"wsAddr"`
		CDN      string `short:"c" long:"cdn"`
	}
	args []string
)

func parseConfig() (err error) {
	parser := flags.NewParser(&options, flags.Default)
	err = flags.NewIniParser(parser).ParseFile("app.ini")
	if err != nil {
		return
	}
	args, err = flags.Parse(&options)
	if err != nil {
		return
	}
	if options.CDN == "" {
		options.CDN = "http://" + options.Server
	}
	if options.WsServer == "" {
		options.WsServer = "ws://" + options.Server + "/websocket"
	}
	return err
}

func main() {
	err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(options, args)
	//return
	m := martini.Classic()

	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})

	m.Get("/build/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		jsDir := strings.Repeat("../", strings.Count(addr, "/")+1)
		r.HTML(200, "build", map[string]string{
			"FullName":       addr,
			"Name":           filepath.Base(addr),
			"DownloadPrefix": options.CDN,
			"Server":         options.Server,
			"WsServer":       options.WsServer,
			"JsDir":          jsDir,
		})
	})
	m.Get("/rebuild/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		mu.Lock()
		defer mu.Unlock()
		br := broadcasts[addr]
		if br == nil {
			return
		}
		if br.closed {
			log.Println("rebuild:", addr)
			delete(broadcasts, addr)
		}
		r.Redirect("/build/"+addr, 302)
	})

	m.Get("/download/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		basename := filepath.Base(addr)

		files := []string{}
		for _, os := range OS {
			for _, arch := range Arch {
				outfile := fmt.Sprintf("http://%s/%s/%s_%s_%s", options.CDN, addr, basename, os, arch)
				if os == "windows" {
					outfile += ".exe"
				}
				files = append(files, outfile)
			}
		}
		r.HTML(200, "download", map[string]interface{}{
			"FullName":       addr,
			"Name":           filepath.Base(addr),
			"DownloadPrefix": options.CDN,
			"Files":          files,
		})
	})

	http.Handle("/", m)
	http.Handle("/websocket", websocket.Handler(WsBuildServer))
	if err = http.ListenAndServe(options.Server, nil); err != nil {
		log.Fatal(err)
	}
}
