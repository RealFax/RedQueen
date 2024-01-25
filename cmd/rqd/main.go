package main

import (
	"fmt"
	"github.com/RealFax/RedQueen/internal/rqd"
	"github.com/RealFax/RedQueen/internal/version"
	"os"
	"os/signal"
	"syscall"

	"github.com/RealFax/RedQueen/config"
)

func main() {
	fmt.Println("Version:", version.String())

	cfg, err := config.New(os.Args...)
	if err != nil {
		fmt.Println("[-] parse config failed, ", err)
		return
	}

	server, err := rqd.NewServer(cfg)
	if err != nil {
		fmt.Println("[-] init server failed, ", err)
		return
	}
	defer server.Close()

	if err = server.ListenAndServe(); err != nil {
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
