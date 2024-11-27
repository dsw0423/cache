package cache

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/dsw0423/cache/consistenthash"
	pb "github.com/dsw0423/cache/pb/cachepb"
	"github.com/shimingyah/pool"
)

type GrpcPool struct {
	pb.UnimplementedCacheServer

	/* <IP>:<PORT> e.g. localhost:8001 */
	self        string
	mu          sync.Mutex
	peers       *consistenthash.Mapping
	grpcGetters map[string]*grpcGetter
}

func NewGrpcPool(self string) *GrpcPool {
	return &GrpcPool{self: self}
}

// implements CacheServer interface
func (g *GrpcPool) Get(ctx context.Context, r *pb.CacheRequest) (*pb.CacheResponse, error) {
	g.Log("%s RPC invoked", g.self)

	groupName := r.GetGroup()
	key := r.GetKey()

	group := GetGroup(groupName)
	if group == nil {
		return nil, fmt.Errorf("no such group %s", groupName)
	}

	bv, err := group.Get(key)
	if err != nil {
		return nil, err
	}
	return &pb.CacheResponse{Value: bv.ByteSlice()}, nil
}

func (g *GrpcPool) Log(format string, v ...any) {
	log.Printf("[Server %s] %s", g.self, fmt.Sprintf(format, v...))
}

func (g *GrpcPool) SetPeers(peers ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.peers = consistenthash.NewMapping(defaultReplicas, nil)
	g.peers.Add(peers...)
	g.grpcGetters = make(map[string]*grpcGetter, len(peers))
	for _, peer := range peers {
		g.grpcGetters[peer] = &grpcGetter{addr: peer}
	}
}

func (g *GrpcPool) PickPeer(key string) (PeerGetter, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if peer := g.peers.Get(key); peer != "" && peer != g.self {
		g.Log("Pick Peer %s", peer)
		getter := g.grpcGetters[peer]
		if getter.connPool == nil {
			p, err := pool.New(peer, pool.DefaultOptions)
			if err != nil {
				return nil, false
			}
			getter.connPool = p
		}
		return getter, true
	}
	return nil, false
}

type grpcGetter struct {
	addr     string
	connPool pool.Pool
}

func (g *grpcGetter) Get(group, key string) ([]byte, error) {
	conn, err := g.connPool.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewCacheClient(conn.Value())
	resp, err := client.Get(context.Background(), &pb.CacheRequest{Group: group, Key: key})
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

var _ PeerGetter = (*grpcGetter)(nil)
var _ PeerPicker = (*GrpcPool)(nil)
var _ pb.CacheServer = (*GrpcPool)(nil)
