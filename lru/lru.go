package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes uint64
	curBytes uint64
	elems    *list.List
	cache    map[string]*list.Element
	/* callback function invoked on evicting the key-value entry. It can be nil.*/
	onEvicted func(key string, value Value)
}

type Value interface {
	/* Size returns number of bytes of the value. */
	Size() int
}

/* Entry is the value of the element of cache's list. */
type Entry struct {
	key   string
	value Value
}

/*
New returns an initialized cache, with maxBytes and callback onEvicted.
zero maxBytes means no limit to cache capacity.
*/
func New(maxBytes uint64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		curBytes:  0,
		elems:     list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

/*
Add adds a new key-value entry into the cache, or update the value if the key already exists.
After adding the new entry, if cache has limit on memory capacity
and current cache size is greater than the max size, a few old entrys will be evicted
in the way of LRU algorithm.
*/
func (cache *Cache) Add(key string, value Value) {
	if elem, ok := cache.cache[key]; ok {
		cache.elems.MoveToFront(elem)
		entry := elem.Value.(*Entry)
		cache.curBytes += uint64(value.Size()) - uint64(entry.value.Size())
		entry.value = value
	} else {
		elem := cache.elems.PushFront(&Entry{
			key:   key,
			value: value,
		})
		cache.cache[key] = elem
		cache.curBytes += uint64(len(key)) + uint64(value.Size())
	}

	for cache.maxBytes != 0 && cache.curBytes > cache.maxBytes {
		cache.removeOldest()
	}
}

func (cache *Cache) removeOldest() {
	elem := cache.elems.Back()
	entry := cache.elems.Remove(elem).(*Entry)
	delete(cache.cache, entry.key)
	cache.curBytes -= uint64(entry.value.Size()) + uint64(len(entry.key))
	if cache.onEvicted != nil {
		cache.onEvicted(entry.key, entry.value)
	}
}

func (cache *Cache) Get(key string) (Value, bool) {
	for k, elem := range cache.cache {
		if k == key {
			cache.elems.MoveToFront(elem)
			entry := elem.Value.(*Entry)
			return entry.value, true
		}
	}
	return nil, false
}

func (cache *Cache) Len() int {
	return cache.elems.Len()
}
