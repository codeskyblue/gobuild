// use martini as web framework
package main

import (
	"fmt"

	"github.com/codegangsta/martini"
)

func main() {
	fmt.Println("Hello World!")
	m := martini.Classic()

	m.Get("/:name", func(params martini.Params) string {
		return "Hello " + params["name"]
	})

	//m.Get("/", func() string {
	//	return "this is index.html"
	//})

	m.Get("/posts/:id", func(params martini.Params) string {
		id := params["id"]
		return id
	})

	m.Run()
}
