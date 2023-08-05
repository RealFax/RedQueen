package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RealFax/RedQueen/client"
)

func main() {
	c, err := client.New(context.Background(), []string{
		//"127.0.0.1:3230",
		//"127.0.0.1:4230",
		"127.0.0.1:5230",
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// watch case
	watcher := client.NewWatcher([]byte("Key1"))
	go func() {
		if wErr := c.Watch(context.Background(), watcher); err != nil {
			log.Fatal("client watch error:", wErr)
		}
	}()
	go func() {
		notify, nErr := watcher.Notify()
		if nErr != nil {
			log.Fatal("watcher notify error:", nErr)
		}
		for {
			val := <-notify
			if val.Value == nil {
				log.Println("[Watch] key has deleted")
				return
			}
			log.Printf("[Watch] Value: %s, TTL: %d, Timestamp: %d", val.Value, val.TTL, val.Timestamp)
		}
	}()

	// basic kv case
	if err = c.Set(context.Background(), []byte("Key1"), []byte("Value1"), 60, nil); err != nil {
		log.Fatal("client set error:", err)
	}

	val, err := c.Get(context.Background(), []byte("Key1"), nil)
	if err != nil {
		log.Fatal("client get error:", err)
	}
	log.Printf("Value: %s, TTL: %d\n", val.Data, val.TTL)

	if err = c.Set(context.Background(), []byte("Key1"), []byte("Value2"), 60, nil); err != nil {
		log.Fatal("client set error:", err)
	}

	if err = c.Delete(context.Background(), []byte("Key1"), nil); err != nil {
		log.Fatal("client delete error:", err)
	}

	const maxWorker = 100
	off := int32(0)

	log.Printf("Starting benchmark...\nWorker: %d\n", maxWorker)
	b := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < maxWorker; i++ {
		b.Add(1)
		go func() {
			defer b.Done()
			c.Set(context.Background(), []byte("Key1"), []byte("Value1"), 60, nil)
			log.Println(atomic.AddInt32(&off, 1))
		}()
	}
	b.Wait()
	log.Printf("Using: %s", time.Since(start).String())
}
