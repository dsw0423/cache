package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dsw0423/cache"
)

var (
	data = map[string]string{
		"dsw":  "23",
		"tom":  "24",
		"jack": "25",
	}
)

func createGroup() *cache.Group {
	return cache.NewGroup("ages", 1<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := data[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *cache.Group) {
	peers := cache.NewHTTPPool(addr)
	peers.SetPeers(addrs...)
	g.RegisterPeers(peers)
	log.Println("cache server is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startApiServer(addr string, g *cache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			bv, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(bv.ByteSlice())
		}))

	log.Println("API server is runningg at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

func main() {
	var port int
	var isApiServer bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&isApiServer, "api", false, "is api server?")
	flag.Parse()

	apiAddr := `http://localhost:9999`
	addrMap := map[int]string{
		8001: `http://localhost:8001`,
		8002: `http://localhost:8002`,
		8003: `http://localhost:8003`,
	}

	addrs := make([]string, 0)
	for _, addr := range addrMap {
		addrs = append(addrs, addr)
	}

	g := createGroup()
	if isApiServer {
		go startApiServer(apiAddr, g)
	}
	startCacheServer(addrMap[port], addrs, g)
}
