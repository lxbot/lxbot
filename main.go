package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lxbot/lxlib/v2/common"
	"github.com/lxbot/lxlib/v2/lxtypes"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)

	adapter := loadAdapter()
	store := loadStore()
	scripts, unifiedScriptCh := loadScripts()

	log.Println("lxbot start")

	getStorageMap := map[string]string{}

	dispose := func() {
		adapter.Dispose()
		store.Dispose()
		for _, script := range scripts {
			script.Dispose()
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(c)
		dispose()
	}()
	go func() {
		for _ = range c {
			close(c)
			dispose()
			os.Exit(125)
		}
	}()

	for {
		select {
		case msg := <-*adapter.MessageCh:
			switch msg.EventType {
			case lxtypes.IncomingMessageEvent:
				for _, s := range scripts {
					s.Write(msg.Body)
				}
			case ExitEvent:
				return
			}
		case msg := <-*unifiedScriptCh:
			switch msg.EventType {
			case lxtypes.OutgoingMessageEvent:
				adapter.Write(msg.Body)
			case lxtypes.GetStorageEvent:
				getStorageMap[msg.ID] = msg.Origin
				store.Write(msg.Body)
			case lxtypes.SetStorageEvent:
				store.Write(msg.Body)
			case ExitEvent:
				return
			}
		case msg := <-*store.MessageCh:
			switch msg.EventType {
			case lxtypes.GetStorageEvent:
				origin := getStorageMap[msg.ID]
				for _, script := range scripts {
					if script.Origin() == origin {
						delete(getStorageMap, msg.ID)
						script.Write(msg.Body)
						break
					}
				}
				common.WarnLog("missing storage map key:", msg.ID)
			case ExitEvent:
				return
			}
		}
	}
}
