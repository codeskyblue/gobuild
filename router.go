package main

import (
	"net/http"
	"path/filepath"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
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
		job := NewBuilder(project, ref, os, arch, nil)
		addr, err := job.Auto()
		if err != nil {
			lg.Error(err)
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

	initBadge()

	/*
		m.Get("/github.com/:account/:proj/:ref/:goos/:goarch", func(p martini.Params, w http.ResponseWriter, r *http.Request) {
			var err error
			var id = uuid.New()
			ch := make(chan string, 1)
			//scribe[id] = ch
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
				"GOROOT="+opts.GOROOT,
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

		m.Get("/github.com/**", func(params martini.Params, r render.Render) {
			r.Redirect("/download/github.com/"+params["_1"], 302)
		})
	*/

	/*
		m.Get("/build/**", func(params martini.Params, r render.Render) {
			addr := params["_1"]
			lg.Debug(addr, "END")
			jsDir := strings.Repeat("../", strings.Count(addr, "/")+1)
			r.HTML(200, "build", map[string]string{
				"FullName":       addr,
				"Name":           filepath.Base(addr),
				"DownloadPrefix": opts.Hostname,
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
	*/

}
