package main

import (
	"io/ioutil"
	"log"
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
		if f.IsDir() {
			n := path.Join(relativePath, f.Name())
			s := lookup(n)
			r = append(r, s...)
		}
		if strings.HasSuffix(n, ".so") {
			p := path.Join(t, n)
			r = append(r, p)
			log.Println("found:", p)
		}
	}
	return r
}

func loadAdapters() (*plugin.Plugin, *chan map[string]interface{}) {
	log.Println("search adapters")

	ch := make(chan map[string]interface{})
	files := lookup("./adapters")
	if len(files) == 0 {
		log.Fatalln("adapter not found")
	}
	file := files[0]
	log.Println("load:", file)

	p, err := plugin.Open(file)
	if err == nil {
		if _, err := p.Lookup("Send"); err != nil {
			panic(err)
		}
		if _, err := p.Lookup("Reply"); err != nil {
			panic(err)
		}
		if fn, err := p.Lookup("Boot"); err == nil {
			log.Println("boot:", file)
			fn.(func(*chan map[string]interface{}))(&ch)
		}
	}

	return p, &ch
}

func loadStores() (*plugin.Plugin, *chan map[string]interface{}){
	log.Println("search stores")

	ch := make(chan map[string]interface{})
	files := lookup("./stores")
	if len(files) == 0 {
		log.Fatalln("store not found")
	}
	file := files[0]
	log.Println("load:", file)

	p, err := plugin.Open(file)
	if err == nil {
		if _, err := p.Lookup("Set"); err != nil {
			panic(err)
		}
		if _, err := p.Lookup("Get"); err != nil {
			panic(err)
		}
		if fn, err := p.Lookup("Boot"); err == nil {
			log.Println("boot:", file)
			fn.(func(*chan map[string]interface{}))(&ch)
		}
	}
	return p, &ch
}

func loadPlugins(store *plugin.Plugin, scripts []*plugin.Plugin) ([]*plugin.Plugin, []*plugin.Plugin, *chan map[string]interface{}){
	log.Println("search plugins")

	ch := make(chan map[string]interface{})
	files := lookup("./plugins")
	beforeScriptsPlugins := make([]*plugin.Plugin, 0)
	afterScriptPlugins := make([]*plugin.Plugin, 0)

	if len(files) == 0 {
		log.Println("plugins not found")
	}
	for _, file := range files {
		log.Println("load:", file)

		p, err := plugin.Open(file)
		if err == nil {
			ok := false
			if _, err := p.Lookup("BeforeScripts"); err == nil {
				ok = true
				beforeScriptsPlugins = append(beforeScriptsPlugins, p)
			}
			if _, err := p.Lookup("AfterScript"); err == nil {
				ok = true
				afterScriptPlugins = append(afterScriptPlugins, p)
			}
			if fn, err := p.Lookup("Boot"); err == nil && ok {
				log.Println("boot:", file)
				fn.(func(*plugin.Plugin, []*plugin.Plugin, *chan map[string]interface{}))(store, scripts, &ch)
			}
		}
	}
	return beforeScriptsPlugins, afterScriptPlugins, &ch
}

func loadScripts(store *plugin.Plugin) ([]*plugin.Plugin, *chan map[string]interface{}){
	log.Println("search scripts")

	ch := make(chan map[string]interface{})
	files := lookup("./scripts")
	plugins := make([]*plugin.Plugin, 0)

	if len(files) == 0 {
		log.Println("script not found")
	}
	for _, file := range files {
		log.Println("load:", file)

		p, err := plugin.Open(file)
		if err == nil {
			if _, err := p.Lookup("OnMessage"); err == nil {
				plugins = append(plugins, p)
				if fn, err := p.Lookup("Boot"); err == nil {
					log.Println("boot:", file)
					fn.(func(*plugin.Plugin, *chan map[string]interface{}))(store, &ch)
				}
			}
		}
	}
	return plugins, &ch
}