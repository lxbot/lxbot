package main

import (
	"log"
)

type M = map[string]interface{}

func main() {
	adapter, aCh := loadAdapters()
	store, _ := loadStores()
	scripts, sCh := loadScripts(store)

	sendSymbol, _ := adapter.Lookup("Send")
	sendFn := sendSymbol.(func(M))
	replySymbol, _ := adapter.Lookup("Reply")
	replyFn := replySymbol.(func(M))

	send := func(m M) {
		switch m["mode"].(string) {
		case "send":
			sendFn(m)
			break
		case "reply":
			replyFn(m)
			break
		}
	}

	log.Println("lxbot start")

	for {
		select {
		case m := <- *aCh:
			for _, s := range scripts {
				if fn, err := s.Lookup("OnMessage"); err == nil {
					fs := fn.(func() []func(M) M)()
					for _, f := range fs {
						go func() {
							r := f(m)
							if r != nil {
								send(r)
							}
						}()
					}
				}
			}
			break
		case m := <- *sCh:
			send(m)
		}
	}
}
