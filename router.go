package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/qiniu/log"
)

func InitRouter() {
	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", map[string]interface{}{
			"Hostname": opts.Hostname,
		})
	})

	m.Get("/github.com/:account/:proj/:ref/:os/:arch", func(p martini.Params, w http.ResponseWriter, r *http.Request) {
		project := "github.com/" + p["account"] + "/" + p["proj"]
		ref := p["ref"]
		os, arch := p["os"], p["arch"]
		fullname := strings.Join([]string{project, ref, os, arch}, "-")
		wb, _ := GetWriteBroadcaster(fullname)
		job := NewBuilder(project, ref, os, arch, wb)
		addr, err := job.Auto()
		if err != nil {
			log.Error("auto build error:", err)
			http.Error(w, "project build error: "+err.Error(), 500)
		}
		http.Redirect(w, r, addr, http.StatusTemporaryRedirect)
	})

	m.Get("/github.com/**", func(p martini.Params, w http.ResponseWriter, r *http.Request) {
		newAddr := "/download/github.com/" + p["_1"]
		http.Redirect(w, r, newAddr, http.StatusTemporaryRedirect)
	})

	m.Get("/download/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		r.HTML(200, "download", map[string]interface{}{
			"Project":  addr,
			"Hostname": opts.Hostname,
			"Name":     filepath.Base(addr),
		})
	})
	m.Get("/build/**", func(params martini.Params, r render.Render) {
		project := params["_1"]
		log.Debug(project, "END")
		r.HTML(200, "build", map[string]string{
			"FullName": project,
			"WsServer": opts.Hostname + "/websocket",
		})
	})

	// all out links call redirect
	m.Get("/redirect", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		log.Info(url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
	// about && document
	m.Get("/about", func(r render.Render) {
		r.HTML(200, "about", nil)
	})
	m.Get("/document", func(r render.Render) {
		path := filepath.Join(filepath.Dir(os.Args[0]), "README.md")
		readme, err := ioutil.ReadFile(path)
		if err != nil {
			log.Error(err)
		}
		data := make(map[string]interface{}, 0)
		data["Readme"] = string(readme)
		r.HTML(200, "document", data)
	})

	m.Get("/search/**", func(params martini.Params, r render.Render) {
		packages, err := NewSearch(params["_1"])
		r.HTML(200, "search", map[string]interface{}{
			"Keyword":        params["keyword"],
			"Packages":       packages.Packages,
			"PackagesLength": len(packages.Packages),
			"Error":          err,
		})
	})
	initBadge()
}
