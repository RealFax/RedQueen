package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	"github.com/RealFax/RedQueen/client"
	"github.com/RealFax/RedQueen/internal/hack"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func Pointer[T comparable](in T) *T {
	var empty T
	if in == empty {
		return nil
	}
	return &in
}

func RPCSet(c *cli.Context) error {
	var (
		key      = c.String("key")
		rawValue = c.String("value")
		ttl      = c.Uint("ttl")

		namespace = c.String("namespace")

		hexed    = c.Bool("hex")
		base64ed = c.Bool("base64")

		err   error
		value []byte
	)
	switch {
	case key == "":
		return errors.New("key should not be empty")
	case hexed:
		if value, err = hex.DecodeString(rawValue); err != nil {
			return err
		}
	case base64ed:
		if value, err = base64.StdEncoding.DecodeString(rawValue); err != nil {
			return err
		}
	default:
		value = hack.String2Bytes(rawValue)
	}

	return invoker.Set(c.Context, hack.String2Bytes(key), value, uint32(ttl), Pointer(namespace))
}

func RPCGet(c *cli.Context) error {
	var (
		key       = c.String("key")
		namespace = c.String("namespace")
	)
	if key == "" {
		return errors.New("key should be not be empty")
	}

	value, err := invoker.Get(c.Context, hack.String2Bytes(key), Pointer(namespace))
	if err != nil {
		return err
	}

	if value.Key != nil {
		fmt.Printf("Key: %s\n", client.BString(value.Key))
	}

	if value.TTL == 0 {
		fmt.Println("TTL: never")
		goto Output
	}
	fmt.Printf("TTL: %d second\n", value.TTL)

Output:
	fmt.Printf("Data: %s\n", client.BString(value.Data))
	return nil
}

func RPCPrefixScan(c *cli.Context) error {
	var (
		prefix    = c.String("prefix")
		reg       = c.String("reg")
		namespace = c.String("namespace")

		offset = c.Uint64("offset")
		limit  = c.Uint64("limit")
	)

	entries, err := invoker.PrefixScan(
		c.Context,
		hack.String2Bytes(prefix),
		offset,
		limit,
		Pointer(reg),
		Pointer(namespace),
	)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Key != nil {
			fmt.Printf("Key: %s, ", client.BString(entry.Key))
		}

		if entry.TTL == 0 {
			fmt.Print("TTL: never, ")
			goto Output
		}
		fmt.Printf("TTL: %d second, ", entry.TTL)
	Output:
		fmt.Printf("Data: %s\n", client.BString(entry.Data))
	}
	return nil
}

func RPCDel(c *cli.Context) error {
	var (
		key       = c.String("key")
		namespace = c.String("namespace")
	)
	if key == "" {
		return errors.New("key should be not be empty")
	}

	return invoker.Delete(c.Context, hack.String2Bytes(key), Pointer(namespace))
}

func RPCWatch(c *cli.Context) error {
	var (
		key       = c.String("key")
		namespace = c.String("namespace")

		ignoreError = c.Bool("ignoreError")
	)
	if key == "" {
		return errors.New("key should be not be empty")
	}

	opts := make([]client.WatcherOption, 0)
	if ignoreError {
		opts = append(opts, client.WatchWithIgnoreErrors())
	}
	if namespace != "" {
		opts = append(opts, client.WatchWithNamespace(Pointer(namespace)))
	}

	watcher := client.NewWatcher(hack.String2Bytes(key), opts...)
	go func() {
		if err := invoker.Watch(c.Context, watcher); err != nil {
			log.Fatal("[-] client watch error:", err)
		}
	}()

	notify, err := watcher.Notify()
	if err != nil {
		return err
	}

	for {
		select {
		case <-c.Done():
			return nil
		case val := <-notify:
			if val.Value == nil {
				log.Printf("[-] key: %s has deleted", val.Key)
				continue
			}
			out := hack.Bytes2String(val.Value)
			if !utf8.FullRune(val.Value) {
				out = hex.EncodeToString(val.Value)
			}

			log.Printf("[+] TTL: %d, Timestamp: %s, Value: %s",
				val.TTL,
				time.Unix(val.Timestamp, 0),
				out,
			)
		}
	}
}

func RPCWatchPrefix(c *cli.Context) error {
	var (
		prefix    = c.String("prefix")
		namespace = c.String("namespace")
	)

	opts := []client.WatcherOption{client.WatchWithPrefix()}
	if namespace != "" {
		opts = append(opts, client.WatchWithNamespace(Pointer(namespace)))
	}

	watcher := client.NewWatcher(hack.String2Bytes(prefix), opts...)
	go func() {
		if err := invoker.WatchPrefix(c.Context, watcher); err != nil {
			log.Fatal("[-] client watch error: ", err)
		}
	}()

	notify, err := watcher.Notify()
	if err != nil {
		return err
	}

	for {
		select {
		case <-c.Done():
			return nil
		case val := <-notify:
			if val.Value == nil {
				log.Printf("[-] key: %s has deleted", val.Key)
				continue
			}
			out := hack.Bytes2String(val.Value)
			if !utf8.FullRune(val.Value) {
				out = hex.EncodeToString(val.Value)
			}

			log.Printf("[+] TTL: %d, Timestamp: %s, Value: %s",
				val.TTL,
				time.UnixMilli(val.Timestamp),
				out,
			)
		}
	}
}
