// use martini as web framework
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/gorilla/websocket"
)

var (
	mu         = &sync.Mutex{}
	broadcasts = make(map[string]*WriteBroadcaster)
	totalUser  = 0

	OS   = []string{"windows", "linux", "darwin"}
	Arch = []string{"386", "amd64"}
)

func runCommand(cmd string, args ...string) chan string {
	ch := make(chan string)
	go func() {
		for {
			time.Sleep(time.Second * 1)
			mu.Lock()
			for _, bc := range broadcasts {
				bc.Write([]byte(time.Now().String() + "\n"))
			}
			mu.Unlock()
		}
	}()
	return ch
}

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

	ch := make(chan string)
	bufbytes, rd := bc.NewReader(name)
	go func() {
		br := bufio.NewReader(rd)
		defer rd.Close()
		defer close(ch)
		charBuf := make([]byte, 100)
		for {
			n, err := br.Read(charBuf)
			if n > 0 {
				ch <- string(charBuf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	return &Project{
		BufferStr: string(bufbytes),
		Channel:   ch,
	}
}

func websocketHandle(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 20, 20)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	// new request comes

	type Message struct {
		Error error  `json:"error"`
		Type  string `json:"type"` // FIXME: how to omited
		Data  string `json:"data"`
	}

	clientMsg := new(Message)
	if err = ws.ReadJSON(clientMsg); err != nil {
		log.Println("read json:", err)
		return
	}
	addr := clientMsg.Data
	name := ws.RemoteAddr().String()

	proj := NewProject(addr, name)

	log.Println("handle request:", addr, name)

	firstMsg := &Message{
		Data: proj.BufferStr,
	}
	err = ws.WriteJSON(firstMsg)
	if err != nil {
		log.Println(err)
		return
	}
	for line := range proj.Channel {
		//log.Println("send message")
		msg := new(Message)
		msg.Data = line
		err := ws.WriteJSON(msg)
		if err != nil {
			log.Println("write failed, user logout")
			return
		}
	}
	log.Println("loop ends")
}

var DownloadPrefix = "http://goplay.qiniudn.com"

func main() {
	m := martini.Classic()

	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	m.Get("/build/**", func(params martini.Params, r render.Render) {
		log.Println(params)
		addr := params["_1"]
		r.HTML(200, "build", map[string]string{
			"FullName":       addr,
			"Name":           filepath.Base(addr),
			"DownloadPrefix": DownloadPrefix,
		})
	})
	m.Get("/rebuild/**", func(params martini.Params, r render.Render) {
		log.Println(params)
		addr := params["_1"]
		delete(broadcasts, addr)
		r.Redirect("/build/"+addr, 302)
	})

	m.Get("/download/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		basename := filepath.Base(addr)

		files := []string{}
		for _, os := range OS {
			for _, arch := range Arch {
				outfile := fmt.Sprintf("%s/%s/%s_%s_%s", DownloadPrefix, addr, basename, os, arch)
				if os == "windows" {
					outfile += ".exe"
				}
				files = append(files, outfile)
			}
		}
		r.HTML(200, "download", map[string]interface{}{
			"FullName":       addr,
			"Name":           filepath.Base(addr),
			"DownloadPrefix": DownloadPrefix,
			"Files":          files,
		})
	})

	m.Get("/websocket", websocketHandle)

	m.Run()
}
