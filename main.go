// use martini as web framework
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
)

var (
	mu     = &sync.Mutex{}
	stdout = ""
)

func runCommand(cmd string, args ...string) chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(time.Second * 1)

	}()
	return ch
}
func main() {
	fmt.Println("Hello World!")

	m := martini.Classic()

	// websocket
	m.Get("/websocket", func(w http.ResponseWriter, r *http.Request) {
		log.Println("http websocket request")
		ws, err := websocket.Upgrade(w, r, nil, 1000, 1000)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		for {
			type Message struct {
				Error error  `json:"error"`
				Data  string `json:"data"`
			}
			msg := new(Message)
			msg.Data = "hello"
			ws.WriteJSON(msg)
			messageType, p, err := ws.ReadMessage()
			fmt.Println(messageType, string(p))
			if err != nil {
				log.Println("bye")
				return
			}
		}
	})

	m.Get("/posts/:id", func(params martini.Params) string {
		id := params["id"]
		return id
	})

	m.Run()
}
