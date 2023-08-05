package main

import (
	"context"
	"github.com/RealFax/RedQueen/client"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := client.New(ctx, []string{
		"127.0.0.1:3230",
		"127.0.0.1:4230",
		"127.0.0.1:5230",
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	mu := client.NewMutexLock(ctx, c, 120, "lock_object")

	if err = mu.Lock(); err != nil {
		log.Fatal("client lock error:", err)
	}
	if err = mu.Unlock(); err != nil {
		log.Fatal("client unlock error:", err)
	}

	ch := make(chan struct{})
	// try lock testing
	mu.Lock()

	go func() {
		if err = mu.TryLock(time.Now().Add(time.Second * 120)); err != nil {
			log.Fatal("client tryLock error:", err)
		}
		log.Println("tryLock success")
		close(ch)
	}()

	time.Sleep(time.Second * 3)
	mu.Unlock()
	<-ch
}
