package main

type M = map[string]interface{}

func main() {
	adapter, aCh := loadAdapters()
	scripts, sCh := loadScripts()

	sendSymbol, _ := adapter.Lookup("Send")
	sendFn := sendSymbol.(func(M))
	replySymbol, _ := adapter.Lookup("Reply")
	replyFn := replySymbol.(func(M))

	for {
		select {
		case m := <- *aCh:
			for _, s := range scripts {
				if fn, err := s.Lookup("OnMessage"); err == nil {
					fs := fn.(func() []func(M) M)()
					for _, f := range fs {
						r := f(m)
						if r != nil {
							switch r["mode"].(string) {
							case "send":
								sendFn(r)
								break
							case "reply":
								replyFn(r)
								break
							}
						}
					}
				}
			}
			break
		case m := <- *sCh:
			switch m["mode"].(string) {
			case "send":
				sendFn(m)
				break
			case "reply":
				replyFn(m)
				break
			}
		}
	}
}