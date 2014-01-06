// use martini as web framework
package main

import (
	"bufio"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/gorilla/websocket"
)

var (
	mu         = &sync.Mutex{}
	broadcasts = make(map[string]*WriteBroadcaster)
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

// FIXME: no close method
func NewCompileProject(addr string, name string) (bufstr string, ch chan string) {
	mu.Lock()
	defer mu.Unlock()
	if broadcasts[addr] == nil {
		writer := NewWriteBroadcaster()
		broadcasts[addr] = writer

		// start to run build command
		cmd := exec.Command("./autobuild", addr)
		cmd.Stdout = writer
		cmd.Stderr = writer
		go func() {
			err := cmd.Run()
			if err != nil {
				log.Println(err)
			}
			// FIXME: need to log to broadcast
			//...
			writer.CloseWriters()
		}()
	}
	bc := broadcasts[addr]

	ch = make(chan string)
	bufbytes, rd := bc.NewReader(name)
	go func() {
		br := bufio.NewReader(rd)
		charBuf := make([]byte, 100)
		for {
			n, err := br.Read(charBuf)
			if n > 0 {
				ch <- string(charBuf[:n])
			}
			if err != nil {
				close(ch)
				return
			}
		}
	}()
	return string(bufbytes), ch
}

func main() {
	m := martini.Classic()

	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	m.Get("/build/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		r.HTML(200, "build", addr)
	})

	// websocket
	m.Get("/websocket", func(w http.ResponseWriter, r *http.Request) {
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
			Data  string `json:"data"`
		}

		clientMsg := new(Message)
		if err = ws.ReadJSON(clientMsg); err != nil {
			log.Println("read json:", err)
			return
		}
		addr := clientMsg.Data
		name := ws.RemoteAddr().String()

		bufstr, ch := NewCompileProject(addr, name)

		log.Println("handle request:", addr, name)
		//log.Println("client msg:", clientMsg)

		//bufstr = "345678\n"
		firstMsg := &Message{
			Data: bufstr,
		}
		err = ws.WriteJSON(firstMsg)
		if err != nil {
			log.Println(err)
			return
		}
		for line := range ch {
			//log.Println("send message")
			msg := new(Message)
			msg.Data = line
			err := ws.WriteJSON(msg)
			if err != nil {
				log.Println("write failed, user logout")
				return
			}
		}
	})

	m.Run()
}
