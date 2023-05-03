package RedQueen_test

import (
	"encoding/json"
	"fmt"
	"github.com/RealFax/RedQueen"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"github.com/hashicorp/raft"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

const (
	raftAddr string = "127.0.0.1:50001"
	httpAddr string = "127.0.0.1:50000"
)

type httpServer struct {
	raft *RedQueen.Raft
	db   store.Store
}

func (s *httpServer) set(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// convert http request data
	{
		var v map[string]string
		if err = json.Unmarshal(b, &v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if b, err = json.Marshal(map[string][]byte{
			"key":   []byte(v["key"]),
			"value": []byte(v["value"]),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	future := s.raft.Raft.Apply(b, time.Millisecond*500)
	if err = future.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if e := future.Response(); e != nil {
		http.Error(w, fmt.Sprintf("error: %v", e), http.StatusBadRequest)
		return
	}

	w.Write([]byte("OK"))
}

func (s *httpServer) get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	value, err := s.db.Get([]byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"key":   key,
		"value": string(value.Data),
	})

}

func (s *httpServer) Start(addr string) error {
	http.HandleFunc("/set", s.set)
	http.HandleFunc("/get", s.get)
	return http.ListenAndServe(addr, nil)
}

func TestNewRaft(t *testing.T) {
	db, err := nuts.New(nuts.Config{
		NodeNum: 3,
		Sync:    true,
		DataDir: "/tmp/red_queen/nuts",
	})
	if err != nil {
		log.Fatal("init nuts db fatal, error:", err)
	}
	defer db.Close()

	raftServer, err := RedQueen.NewRaft(RedQueen.RaftConfig{
		ServerID:              "node-1",
		Addr:                  raftAddr,
		BoldStorePath:         "/tmp/red_queen/bolt",
		FileSnapshotStorePath: "/tmp/red_queen/snapshot",
		FSM:                   RedQueen.NewFSM(db),
		Clusters: []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:       "node-1",
				Address:  raft.ServerAddress(raftAddr),
			},
		},
	})
	if err != nil {
		t.Fatal("init raft server fatal, error: ", err)
	}

	server := &httpServer{
		raft: raftServer,
		db:   db,
	}

	if err = server.Start(httpAddr); err != nil {
		t.Fatal("init http server fatal, error: ", err)
	}
}
