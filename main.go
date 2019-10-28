package main

import (
	"os"
	"os/signal"

	"github.com/libp2p/repl/menu"
)

func main() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill)

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
