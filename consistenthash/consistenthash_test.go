package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	//使用简单的 hashM，即把 数字字符串 -> 数字
	hashM := New(3, func(key []byte) uint32 {
		hashed, _ := strconv.Atoi(string(key))
		return uint32(hashed)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hashM.Add("6", "4", "2")

	if len(hashM.keys) != 9 {
		t.Fatalf("nodes add error")
	}

	//t.Log(hashM.keys, hashM.keyMap)

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		name := hashM.Get(k)
		if name != v {
			t.Fatalf("Asking for %s, expected %s, got %s", k, v, name)
		}
	}

	// Adds 8, 18, 28
	hashM.Add("8")

	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		name := hashM.Get(k)
		if name != v {
			t.Fatalf("Asking for %s, expected %s, got %s", k, v, name)
		}
	}
}
