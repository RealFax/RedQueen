package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/RealFax/RedQueen"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type httpServer struct {
	raft *RedQueen.Raft
	db   store.Store
}

func (s *httpServer) commit(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// convert http request data
	{
		var cmd RedQueen.LogPayload
		if err = json.Unmarshal(b, &cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if b, err = msgpack.Marshal(cmd); err != nil {
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
	namespace := r.URL.Query().Get("namespace")

	db, err := s.db.Namespace(namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	value, err := db.Get([]byte(key))
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
	http.HandleFunc("/commit", s.commit)
	http.HandleFunc("/get", s.get)
	return http.ListenAndServe(addr, nil)
}

var (
	nodeID   = flag.String("node", "node-1", "node id")
	raftAddr = flag.String("cluster", "127.0.0.1:50001", "cluster listen addr")
	httpAddr = flag.String("http", "127.0.0.1:50000", "http server listen addr")
)

type Cluster struct {
	ID      string `json:"id"`
	Address string `json:"addr"`
}

func main() {
	flag.Parse()
	if *nodeID == "" || *raftAddr == "" || *httpAddr == "" {
		log.Fatal("invalid server args")
	}

	f, err := os.Open("./clusters.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var clusters []Cluster
	if err = json.NewDecoder(f).Decode(&clusters); err != nil {
		log.Fatal(err)
	}

	db, err := nuts.New(nuts.Config{
		NodeNum: 3,
		Sync:    true,
		DataDir: "./red_queen/" + *nodeID + "/nuts",
	})
	if err != nil {
		log.Fatal("init nuts db fatal, error:", err)
	}
	defer db.Close()

	raftServer, err := RedQueen.NewRaft(RedQueen.RaftConfig{
		ServerID:              *nodeID,
		Addr:                  *raftAddr,
		BoldStorePath:         "./red_queen/" + *nodeID + "/bolt",
		FileSnapshotStorePath: "./red_queen/" + *nodeID + "/",
		FSM:                   RedQueen.NewFSM(db),
		Clusters: func() []raft.Server {
			c := make([]raft.Server, len(clusters))
			for i, cluster := range clusters {
				c[i] = raft.Server{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(cluster.ID),
					Address:  raft.ServerAddress(cluster.Address),
				}
			}
			return c
		}(),
	})
	if err != nil {
		log.Fatal("init raft server fatal, error: ", err)
	}

	server := &httpServer{
		raft: raftServer,
		db:   db,
	}

	if err = server.Start(*httpAddr); err != nil {
		log.Fatal("init http server fatal, error: ", err)
	}
}
