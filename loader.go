package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
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
		if f.Mode()&0100 != 0 {
			p := path.Join(t, n)
			r = append(r, p)
			log.Println("found:", p)
		}
	}
	return r
}

func loadAdapter() *Process {
	log.Println("search adapters")

	files := lookup("./adapters")
	if len(files) == 0 {
		log.Fatalln("adapter not found")
	}
	file := files[0]
	log.Println("load:", file)

	return mustNewProcess(file)
}

func loadStore() *Process {
	log.Println("search stores")

	files := lookup("./stores")
	if len(files) == 0 {
		// FIXME: storeは無くてもいいけどこのままだとGetStorageしたら二度と返ってこなくなる
		return newDumbProcess()
	}
	file := files[0]
	log.Println("load:", file)

	return mustNewProcess(file)
}

func loadScripts() ([]*Process, *chan *InternalMessage) {
	log.Println("search scripts")

	files := lookup("./scripts")
	scripts := make([]*Process, len(files))
	ch := make(chan *InternalMessage)

	if len(files) == 0 {
		log.Println("script not found")
	}
	for i, file := range files {
		log.Println("load:", file)
		scripts[i] = mustNewProcess(file)

		go func(index int) {
			for {
				msg := <-*scripts[index].MessageCh
				ch <- msg
			}
		}(i)
	}

	return scripts, &ch
}
