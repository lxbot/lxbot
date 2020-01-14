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

func loadAdapters() (*plugin.Plugin, *chan map[string]interface{}) {
	ch := make(chan map[string]interface{})
	files := lookup("./adapters")
	if len(files) == 0 {
		panic("adapter not found")
	}
	file := files[0]
	p, err := plugin.Open(file)
	if err == nil {
		if _, err := p.Lookup("Send"); err != nil {
			panic(err)
		}
		if _, err := p.Lookup("Reply"); err != nil {
			panic(err)
		}
		if fn, err := p.Lookup("Boot"); err == nil {
			fn.(func(*chan map[string]interface{}))(&ch)
		}
	}
	return p, &ch
}

func loadScripts() ([]*plugin.Plugin, *chan map[string]interface{}){
	ch := make(chan map[string]interface{})
	files := lookup("./scripts")
	if len(files) == 0 {
		panic("adapter not found")
	}
	plugins := make([]*plugin.Plugin, 0)
	for _, file := range files {
		p, err := plugin.Open(file)
		if err == nil {
			if _, err := p.Lookup("OnMessage"); err == nil {
				plugins = append(plugins, p)
				if fn, err := p.Lookup("Boot"); err == nil {
					fn.(func(*chan map[string]interface{}))(&ch)
				}

			}
		}
	}
	return plugins, &ch
}