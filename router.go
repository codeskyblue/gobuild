package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/shxsun/gobuild/models"
)

func InitRouter() {
	var scribe = make(map[string]chan string)
	var GOROOT = opts.GOROOT

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})
	m.Get("/github.com/:account/:proj/:ref/:goos/:goarch", func(p martini.Params, w http.ResponseWriter, r *http.Request) {
		var err error
		var id = uuid.New()
		ch := make(chan string, 1)
		scribe[id] = ch
		// create log
		outfd, err := os.OpenFile("log/"+id, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			lg.Error(err)
		}
		defer outfd.Close()
		// build cmd
		cmd := exec.Command("bin/build", "github.com/"+p["account"]+"/"+p["proj"])
		envs := []string{}
		for k, v := range p {
			envs = append(envs, strings.ToUpper(k)+"="+v)
		}
		envs = append(envs,
			"GOROOT="+GOROOT,
			"BUILD_HOST="+"127.0.0.1:3000",
			"BUILD_ID="+id)
		cmd.Env = envs
		cmd.Stdout = outfd
		cmd.Stderr = outfd

		err = cmd.Run()
		if err != nil {
			lg.Error(err)
			return
		}

		var message string
		select {
		case message = <-ch:
		case <-time.After(time.Second * 1):
			message = "timeout"
		}
		lg.Info("finish build:", message)
		http.Redirect(w, r, message, http.StatusTemporaryRedirect)
		return
	})

	m.Get("/info/:id/output", func(p martini.Params) string {
		return "unfinished"
	})
	m.Post("/api/:id/binary", func(w http.ResponseWriter, r *http.Request, p martini.Params) string {
		addr := "xyz-xxxx"
		/*
			addr, err := uploadFile(r.Body)
			if err != nil {
				lg.Error(err)
			}
			fmt.Println(addr)
		*/
		if ch := scribe[p["id"]]; ch != nil {
			ch <- addr
			close(ch)
		}
		return "finished:" + addr
	})

	/*m.Get("/github.com/**", func(params martini.Params, r render.Render) {
		r.Redirect("/download/github.com/"+params["_1"], 302)
	})
	*/

	m.Get("/build/**", func(params martini.Params, r render.Render) {
		addr := params["_1"]
		lg.Debug(addr, "END")
		jsDir := strings.Repeat("../", strings.Count(addr, "/")+1)
		r.HTML(200, "build", map[string]string{
			"FullName":       addr,
			"Name":           filepath.Base(addr),
			"DownloadPrefix": opts.CDN,
			"Server":         opts.Server,
			"WsServer":       opts.WsServer + "/websocket",
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
		if br.Closed() {
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

		record := new(models.Project)
		record.Name = project
		record.Project = project // FIXME: delete it
		record.Sha = sha
		err := models.SyncProject(record)
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
		sha, err := models.GetSha(project)
		if err != nil {
			return 500, err.Error()
		}
		// path like: cdn://project/sha/os_arch/filename
		r.Redirect(opts.CDN+"/"+filepath.Join(project, sha, os+"_"+arch, filename), 302)
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
			"Server":  opts.Server,
			//"CDN":     opts.CDN,
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
				outfile := fmt.Sprintf("%s/%s/%s_%s_%s", opts.CDN, addr, basename, os, arch)
				if os == "windows" {
					outfile += ".exe"
				}
				files = append(files, outfile)
			}
		}
		r.HTML(200, "download", map[string]interface{}{
			"Project": addr,
			"Server":  opts.Server,
			"Name":    filepath.Base(addr),
			"CDN":     opts.CDN,
			"Files":   files,
		})
	})
}
