package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"sort"
	"sync"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	lru "github.com/hashicorp/golang-lru/v2"
)

const cache_size = 65000

type ValuesetCache[K string, V map[string]string] struct {
	ARCCache *lru.ARCCache[K, V]
	mu       sync.Mutex
}

func NewValuesetCache[K string, V map[string]string]() (ValuesetCache[K, V], error) {
	cache, err := lru.NewARC[K, V](cache_size)
	if err != nil {
		return ValuesetCache[K, V]{}, err
	}
	return ValuesetCache[K, V]{ARCCache: cache}, nil
}

func (cache *ValuesetCache[K, V]) Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	val, ok := cache.ARCCache.Get(K(generateHashFromPair(req.KeyValues)))
	if !ok {
		return ls.ValuesetLookupResponse{}, false
	}
	return ls.ValuesetLookupResponse{KeyValues: V(val)}, true
}

// ValuesetCache.Set caches a generated hash as a key, storing the request as its corresponding value
func (cache *ValuesetCache[K, V]) Set(req ls.ValuesetLookupRequest, res ls.ValuesetLookupResponse) {
	cache.mu.Lock()
	cache.ARCCache.Add(K(generateHashFromPair(req.KeyValues)), V(req.KeyValues))
	cache.mu.Unlock()
}

func generateHashFromPair(keyValues map[string]string) string {
	pairList := make(pairList, len(keyValues))
	pairListIdx := 0
	for k, v := range keyValues {
		pairList[pairListIdx] = pair{k, v}
		pairListIdx++
	}
	sort.Sort(pairList)
	// generate hash on sorted pair list
	h := sha256.New()
	h.Write(hashPair(pairList))
	return hex.EncodeToString(h.Sum(nil))
}

func hashPair(s []pair) []byte {
	var b bytes.Buffer
	gob.NewEncoder(&b).Encode(s)
	return b.Bytes()
}

type pair struct {
	Key   string
	Value string
}

type pairList []pair

type Sort interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

func (p pairList) Len() int           { return len(p) }
func (p pairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
