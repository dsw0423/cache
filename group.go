package cache

import (
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name   string
	ca     *cache
	getter Getter
	peers  PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxBytes uint64, getter Getter) *Group {
	if getter == nil {
		panic("getter must be non nil.")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   name,
		ca:     &cache{maxBytes: maxBytes},
		getter: getter,
	}
	groups[name] = g

	return g
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key must be specified")
	}

	if v, ok := g.ca.get(key); ok {
		log.Println("[Cache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.loadFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[Cache] Failed load from peer", err)
		}
	}

	return g.loadFromLocal(key)
}

func (g *Group) loadFromLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{data: copyBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.ca.add(key, value)
}

func (g *Group) loadFromPeer(peer PeerGetter, key string) (ByteView, error) {
	if bytes, err := peer.Get(g.name, key); err != nil {
		return ByteView{}, err
	} else {
		return ByteView{bytes}, nil
	}
}
