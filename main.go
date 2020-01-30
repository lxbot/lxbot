package main

import (
	"log"
)

type M = map[string]interface{}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)

	adapter, aCh := loadAdapters()
	store, _ := loadStores()
	scripts, sCh := loadScripts(store)
	beforeScriptsPlugins, afterScriptPlugins, pCh := loadPlugins(store, scripts)

	sendSymbol, _ := adapter.Lookup("Send")
	sendFn := sendSymbol.(func(M))
	replySymbol, _ := adapter.Lookup("Reply")
	replyFn := replySymbol.(func(M))

	send := func(m M) {
		switch m["mode"].(string) {
		case "send":
			log.Println("send:", m["message"].(M)["text"])
			sendFn(m)
			break
		case "reply":
			log.Println("send:", m["message"].(M)["text"])
			replyFn(m)
			break
		}
	}

	log.Println("lxbot start")

	for {
		select {
		case m := <- *aCh:
			for _, p := range beforeScriptsPlugins {
				if fn, err := p.Lookup("BeforeScripts"); err == nil {
					fs := fn.(func() []func(M) M)()
					for _, f := range fs {
						m = f(m)
					}
				}
			}
			for _, s := range scripts {
				if fn, err := s.Lookup("OnMessage"); err == nil {
					fs := fn.(func() []func(M) M)()
					for _, f := range fs {
						go func(gf func(M) M) {
							cm, err := deepCopy(m)
							if err != nil {
								log.Fatalln(err)
							}
							r := gf(cm)
							if r != nil {
								for _, p := range afterScriptPlugins {
									if pfn, err := p.Lookup("AfterScript"); err == nil {
										pfs := pfn.(func() []func(M) M)()
										for _, pf := range pfs {
											r = pf(r)
										}
									}
								}
								go send(r)
							}
						}(f)
					}
				}
			}
			break
		case m := <- *sCh:
			go send(m)
			break
		case m := <- *pCh:
			go send(m)
		}
	}
}
