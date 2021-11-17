package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/repl/menu"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	repl, err := menu.NewREPL()
	if err != nil {
		panic(err)
	}

	go func() {
		<-ch
		repl.Close()
	}()

	repl.Run()
}
