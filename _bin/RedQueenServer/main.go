package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RealFax/RedQueen"
	"github.com/RealFax/RedQueen/config"
)

func main() {
	cfg, err := config.ReadFromArgs(os.Args...)
	if err != nil {
		fmt.Println("[-] parse config failed, ", err)
		return
	}

	server, err := RedQueen.NewServer(cfg)
	if err != nil {
		fmt.Println("[-] init server failed, ", err)
		return
	}

	if err = server.ListenClient(); err != nil {
		fmt.Println("[-] listen client failed, ", err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-c:
		if err = server.Close(); err != nil {
			fmt.Println("[-] server close failed, ", err)
		}
		return
	}
}
