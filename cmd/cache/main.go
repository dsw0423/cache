package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/dsw0423/cache"
	pb "github.com/dsw0423/cache/pb/cachepb"
	"github.com/shimingyah/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

func startCacheServerGrpc(addr string, addrs []string, g *cache.Group) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	peers := cache.NewGrpcPool(addr)
	peers.SetPeers(addrs...)
	g.RegisterPeers(peers)
	log.Println("gRPC cache server is running at", addr)
	grpcServer := grpc.NewServer(
		grpc.InitialWindowSize(pool.InitialWindowSize),
		grpc.InitialConnWindowSize(pool.InitialConnWindowSize),
		grpc.MaxSendMsgSize(pool.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(pool.MaxRecvMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    pool.KeepAliveTime,
			Timeout: pool.KeepAliveTimeout,
		}),
	)
	pb.RegisterCacheServer(grpcServer, peers)
	log.Fatal(grpcServer.Serve(lis))
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
	var grpc bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&isApiServer, "api", false, "is api server?")
	flag.BoolVar(&grpc, "grpc", true, "enable grpc communication?")
	flag.Parse()

	apiAddr := `http://localhost:9999`
	addrMap := map[int]string{
		8001: `http://localhost:8001`,
		8002: `http://localhost:8002`,
		8003: `http://localhost:8003`,
	}

	addrMapGrpc := map[int]string{
		8001: `localhost:8001`,
		8002: `localhost:8002`,
		8003: `localhost:8003`,
	}

	g := createGroup()
	if isApiServer {
		go startApiServer(apiAddr, g)
	}

	if grpc {
		addrs := make([]string, 0)
		for _, addr := range addrMapGrpc {
			addrs = append(addrs, addr)
		}
		startCacheServerGrpc(addrMapGrpc[port], addrs, g)
	} else {
		addrs := make([]string, 0)
		for _, addr := range addrMap {
			addrs = append(addrs, addr)
		}
		startCacheServer(addrMap[port], addrs, g)
	}
}
