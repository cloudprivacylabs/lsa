package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	lru "github.com/hashicorp/golang-lru/v2"
)

const cache_size = 65000

// ValuesetCache is a thread-safe fixed size cache which uses ARC; stores the hash of a ValuesetRequest
// and its matching ValuesetResponse
type ValuesetCache[K string, V map[string]string] struct {
	ARCCache *lru.ARCCache[K, V]
}

func NewValuesetCache[K string, V map[string]string]() (ValuesetCache[K, V], error) {
	cache, err := lru.NewARC[K, V](cache_size)
	if err != nil {
		return ValuesetCache[K, V]{}, err
	}
	return ValuesetCache[K, V]{ARCCache: cache}, nil
}

// ValuesetCache.Lookup returns the cached ValuesetLookupResponse if exists
func (cache *ValuesetCache[K, V]) Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, bool) {
	val, ok := cache.ARCCache.Get(K(generateHashFromRequest(req)))
	if !ok {
		return ls.ValuesetLookupResponse{}, false
	}
	return ls.ValuesetLookupResponse{KeyValues: V(val)}, true
}

// ValuesetCache.Set caches a generated hash as a key, storing the request as its corresponding value
func (cache *ValuesetCache[K, V]) Set(req ls.ValuesetLookupRequest, res ls.ValuesetLookupResponse) {
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
