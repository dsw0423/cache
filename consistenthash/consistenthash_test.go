package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	mapping := NewMapping(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	mapping.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if got := mapping.Get(k); got != v {
			t.Errorf("Asking for %s, expected %s, but got %s", k, v, got)
		}
	}

	// Adds 8, 18, 28
	mapping.Add("8")

	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		if got := mapping.Get(k); got != v {
			t.Errorf("Asking for %s, expected %s, but got %s", k, v, got)
		}
	}

}
