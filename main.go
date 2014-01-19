// use martini as web framework
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/jessevdk/go-flags"
	"github.com/shxsun/klog"
)

var (
	mu         = &sync.Mutex{}
	broadcasts = make(map[string]*WriteBroadcaster)
	totalUser  = 0

	lg = klog.DevLog

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
			Debugf("start cmd error: %v", err)
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
		writer := NewWriteBroadcaster()
		broadcasts[addr] = writer
		startCommand(writer, "./autobuild", addr)
	}
	bc := broadcasts[addr]

	Debugf("%s: lock 2.2 new reader done", name)
	bufbytes, rd := bc.NewReader(name)
	reader := NewBufReader(rd)
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
		Debugf("read json error: %v", err)
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
		Debugf("send first msg error: %v", err)
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
	options struct {
		Server   string `short:"s" long:"serverAddr"`
		WsServer string `short:"w" long:"wsAddr"`
		CDN      string `short:"c" long:"cdn"`
	}
	args []string

	listenAddr = ""
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
		options.CDN = "http://" + options.Server + "/files"
	}
	if options.WsServer == "" {
		options.WsServer = "ws://" + options.Server
	}
	sepIndex := strings.Index(options.Server, ":")
	listenAddr = options.Server[sepIndex:]
	return err
}

var m = martini.Classic()

func init() {
	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))
	initRouter()
}

func initRouter() {
	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})
	m.Get("/build/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		lg.Debug(addr, "END")
		jsDir := strings.Repeat("../", strings.Count(addr, "/")+1)
		r.HTML(200, "build", map[string]string{
			"FullName":       addr,
			"Name":           filepath.Base(addr),
			"DownloadPrefix": options.CDN,
			"Server":         options.Server,
			"WsServer":       options.WsServer + "/websocket",
			"JsDir":          jsDir,
		})
	})
	m.Get("/rebuild/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		mu.Lock()
		defer func() {
			mu.Unlock()
			r.Redirect("/build/"+addr, 302) // FIXME: this not good with nginx proxy
		}()
		br := broadcasts[addr]
		lg.Debug("rebuild", addr, "END")
		if br == nil {
			return
		}
		if br.closed {
			lg.Debug("rebuild:", addr)
			delete(broadcasts, addr)
		}
		lg.Debug("end rebuild")
	})

	// for autobuild script upload result
	m.Post("/api/update", func(req *http.Request) (int, string) {
		// for secure reason, only accept 127.0.0.1 address
		lg.Warnf("Unexpected request: %s", req.RemoteAddr)
		if !strings.HasPrefix(req.RemoteAddr, "127.0.0.1:") {
			lg.Warnf("Unexpected request: %s", req.RemoteAddr)
			return 200, ""
		}
		project, sha := req.FormValue("p"), req.FormValue("sha")
		lg.Debug(project, sha)

		record := new(Latest)
		record.Project = project
		record.Sha = sha
		err := SyncProject(record)
		if err != nil {
			lg.Error(err)
			return 500, err.Error()
		}
		return 200, "OK"
	})

	m.Get("/dl", func(req *http.Request, r render.Render) (code int, body string) {
		os, arch := req.FormValue("os"), req.FormValue("arch") //"windows", "amd64"
		project := req.FormValue("p")                          //"github.com/shxsun/fswatch"
		filename := filepath.Base(project)
		if os == "windows" {
			filename += ".exe"
		}

		// sha should get from db
		//sha := "d1077e2e106489b81c6a404e6951f1fca8967172"
		sha, err := GetSha(project)
		if err != nil {
			return 500, err.Error()
		}
		// path like: cdn://project/sha/os_arch/filename
		r.Redirect(options.CDN+"/"+filepath.Join(project, sha, os+"_"+arch, filename), 302)
		return
	})

	m.Get("/dlscript/**", func(params martini.Params) (s string, err error) {
		project := params["_1"]
		t, err := template.ParseFiles("templates/dlscript.sh.tmpl")
		if err != nil {
			lg.Error(err)
			return
		}
		buf := bytes.NewBuffer(nil)
		err = t.Execute(buf, map[string]interface{}{
			"Project": project,
			"Server":  options.Server,
			//"CDN":     options.CDN,
		})
		if err != nil {
			lg.Error(err)
			return
		}
		return string(buf.Bytes()), nil
	})

	m.Get("/download/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		basename := filepath.Base(addr)

		files := []string{}
		for _, os := range OS {
			for _, arch := range Arch {
				outfile := fmt.Sprintf("%s/%s/%s_%s_%s", options.CDN, addr, basename, os, arch)
				if os == "windows" {
					outfile += ".exe"
				}
				files = append(files, outfile)
			}
		}
		r.HTML(200, "download", map[string]interface{}{
			"Project": addr,
			"Server":  options.Server,
			"Name":    filepath.Base(addr),
			"CDN":     options.CDN,
			"Files":   files,
		})
	})
}

func main() {
	var err error
	if err = parseConfig(); err != nil {
		return
	}
	lg.Info("gobuild service stated ...")
	fmt.Println("\tlisten address:", listenAddr)
	fmt.Println("\twebsocket addr:", options.WsServer)
	fmt.Println("\tCDN:", options.CDN)

	http.Handle("/", m)
	http.Handle("/websocket", websocket.Handler(WsBuildServer))

	if err = http.ListenAndServe(listenAddr, nil); err != nil {
		lg.Fatal(err)
	}
}
