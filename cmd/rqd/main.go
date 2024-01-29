package main

import (
	"fmt"
	"github.com/RealFax/RedQueen/internal/rqd"
	"github.com/RealFax/RedQueen/internal/rqd/config"
	"github.com/RealFax/RedQueen/internal/version"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Version:", version.String())

	cfg, err := config.New(os.Args...)
	if err != nil {
		fmt.Println("[-] Failed parse config, : ", err)
		return
	}

	server, err := rqd.NewServer(cfg)
	if err != nil {
		fmt.Println("[-] Failed to initialize server, ", err)
		return
	}
	defer server.Shutdown() //

	if err = server.ListenAndServe(); err != nil {
		fmt.Println("[-] Failed run server, ", err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-c:
		server.Shutdown()
		return
	}
}
