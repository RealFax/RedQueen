package basic_key_value

import (
	"context"
	"github.com/RealFax/RedQueen/client"
	"log"
)

func main() {
	c, err := client.New(context.Background(), []string{
		"127.0.0.1:2540",
		"127.0.0.1:3540",
		"127.0.0.1:4540",
	}, false)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// watch case
	watcher := client.NewWatcher([]byte("Key1"), true)
	defer watcher.Close()
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
			log.Printf("[Watch] Value: %s, Timestamp: %d", val.Data, val.Timestamp)
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

}
