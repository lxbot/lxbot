package main

import (
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"strings"
)



func mustGetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func lookup(relativePath string) []string {
	wd := mustGetWd()
	t := path.Join(wd, relativePath)
	files, err := ioutil.ReadDir(t)
	if err != nil {
		panic(err)
	}
	r := make([]string, 0, 0)
	for _, f := range files {
		n := f.Name()
		if !f.IsDir() && strings.HasSuffix(n, ".so") {
			p := path.Join(t, n)
			r = append(r, p)
		}
	}
	return r
}

func loadPlugins() {
	// TODO:
}

func loadAdapters() *chan map[string]interface{} {
	ch := make(chan map[string]interface{})
	files := lookup("./adapters")
	for _, f := range files {
		p, err := plugin.Open(f)
		if err == nil {
			if f, err := p.Lookup("Boot"); err == nil {
				f.(func(*chan map[string]interface{}))(&ch)
			}
		}
	}
	return &ch
}

func loadScripts() {
	// TODO:
}