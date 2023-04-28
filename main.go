package main

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"log"
	"net/http"
)

func main() {

	db, dErr := nuts.New(nuts.Config{
		NodeNum: 1,
		Sync:    true,
		DataDir: "/tmp/nuts",
	})
	if dErr != nil {
		log.Fatal(dErr)
	}

	locks := make(map[string]*store.Lock)

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		lock, err := store.NewLock(db)
		if err != nil {
			w.Write([]byte("创建锁失败: " + err.Error()))
			return
		}
		locks[lock.UUID] = lock
		w.Write([]byte(lock.UUID))
	})

	http.HandleFunc("/lock", func(w http.ResponseWriter, r *http.Request) {
		lockID := r.URL.Query().Get("id")
		lock, ok := locks[lockID]
		if !ok {
			w.Write([]byte("锁: " + lockID + " 不存在"))
			return
		}

		if err := lock.Lock(); err != nil {
			w.Write([]byte("加锁失败: " + err.Error()))
			return
		}

		w.Write([]byte("success"))
	})

	http.HandleFunc("/unlock", func(w http.ResponseWriter, r *http.Request) {
		lockID := r.URL.Query().Get("id")
		lock, ok := locks[lockID]
		if !ok {
			w.Write([]byte("锁: " + lockID + " 不存在"))
			return
		}

		if err := lock.Unlock(); err != nil {
			w.Write([]byte("解锁失败: " + err.Error()))
			return
		}

		w.Write([]byte("success"))
	})

	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}
