package main

import (
	"github.com/Unknwon/com"
	"net/http"
)

var searchApi = "https://gowalker.org/api/v1/search?key={keyword}&gorepo=false&gosubrepo=false&cmd=true&cgo=false"

type SearchItem struct {
	Name       string `json:"project_name"`
	Path       string `json:"project_path"`
	HomePage   string `json:"homepage"`
	ImportPath string `json:"import_path"`
}

type SearchPackages struct {
	Packages []*SearchItem `json:"packages"`
}

func NewSearch(key string) (*SearchPackages, error) {
	url := com.Expand(searchApi, map[string]string{
		"keyword": key,
	})
	packages := new(SearchPackages)
	err := com.HttpGetJSON(&http.Client{}, url, packages)
	return packages, err
}
