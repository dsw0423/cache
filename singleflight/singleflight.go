package singleflight

import "sync"

type call struct {
	wg    sync.WaitGroup
	value any
	err   error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if call, ok := g.m[key]; ok {
		g.mu.Unlock()
		call.wg.Wait()
		return call.value, call.err
	}

	call := new(call)
	g.m[key] = call
	call.wg.Add(1)
	g.mu.Unlock()

	call.value, call.err = fn()
	call.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return call.value, call.err
}
