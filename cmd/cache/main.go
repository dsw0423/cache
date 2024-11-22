package main

import (
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

func main() {
	cache.NewGroup("ages", 1<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := data[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	pool := cache.NewHTTPPool(addr)
	log.Println("cache server is running at", addr)
	log.Fatal(http.ListenAndServe(addr, pool))
}
