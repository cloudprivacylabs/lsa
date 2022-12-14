package valueset

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	lru "github.com/hashicorp/golang-lru/v2"
)

const cache_size = 65000

type ValuesetCache interface {
	Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, bool)
	Set(req ls.ValuesetLookupRequest, res ls.ValuesetLookupResponse)
}

type NoCache struct{}

func (NoCache) Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, bool) {
	return ls.ValuesetLookupResponse{}, false
}

func (NoCache) Set(req ls.ValuesetLookupRequest, res ls.ValuesetLookupResponse) {}

type LRUCache[K string, V map[string]string] struct {
	ARCCache *lru.ARCCache[K, V]
}

func NewValuesetCache[K string, V map[string]string]() (LRUCache[K, V], error) {
	cache, err := lru.NewARC[K, V](cache_size)
	if err != nil {
		return LRUCache[K, V]{}, err
	}
	return LRUCache[K, V]{ARCCache: cache}, nil
}

// ValuesetCache.Lookup returns the cached ValuesetLookupResponse if exists
func (cache *LRUCache[K, V]) Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, bool) {
	val, ok := cache.ARCCache.Get(K(generateHashFromRequest(req)))
	if !ok {
		return ls.ValuesetLookupResponse{}, false
	}
	return ls.ValuesetLookupResponse{KeyValues: V(val)}, true
}

// ValuesetCache.Set caches a generated hash as a key, storing the request as its corresponding value
func (cache *LRUCache[K, V]) Set(req ls.ValuesetLookupRequest, res ls.ValuesetLookupResponse) {
	cache.ARCCache.Add(K(generateHashFromRequest(req)), V(req.KeyValues))
}

func generateHashFromRequest(req ls.ValuesetLookupRequest) string {
	collection := make([]string, 0, len(req.KeyValues)*2+len(req.TableIDs))
	collection = append(collection, req.TableIDs...)
	for k, v := range req.KeyValues {
		collection = append(collection, k, v)
	}
	sort.SliceStable(collection, func(i, j int) bool {
		return collection[i] < collection[j]
	})
	// generate hash on sorted list
	h := sha256.New()
	for _, v := range collection {
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}
