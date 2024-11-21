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

func (g *Group) load(key string) (ByteView, error) {
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
