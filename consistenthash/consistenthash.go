package consistenthash

import (
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
)

type HashFunc func([]byte) uint32

type Mapping struct {
	hash     HashFunc
	replicas int
	keys     []uint32
	hashMap  map[uint32]string
}

func NewMapping(replicas int, hash HashFunc) *Mapping {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}

	return &Mapping{
		hash:     hash,
		replicas: replicas,
		keys:     make([]uint32, 0),
		hashMap:  make(map[uint32]string),
	}
}

func (m *Mapping) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hashValue := m.hash([]byte(strconv.Itoa(i) + key))
			m.keys = append(m.keys, hashValue)
			m.hashMap[hashValue] = key
		}
	}
	slices.Sort(m.keys)
}

func (m *Mapping) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hashValue := m.hash([]byte(key))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hashValue
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
