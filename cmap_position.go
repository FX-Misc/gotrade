package gotrade

import (
	"encoding/json"
	"hash/fnv"
	"sync"
)

var SHARD_COUNT = 32

// TODO: Add Keys function which returns an array of keys for the map.

// A "thread" safe map of type string:*StockPosition.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type PositionMap []*PositionMapShared
type PositionMapShared struct {
	items        map[string]*StockPosition
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// Creates a new concurrent map.
func NewPositionMap() PositionMap {
	m := make(PositionMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &PositionMapShared{items: make(map[string]*StockPosition)}
	}
	return m
}

// Returns shard under given key
func (m PositionMap) GetShard(key string) *PositionMapShared {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return m[int(hasher.Sum32())%SHARD_COUNT]
}

// Sets the given value under the specified key.
func (m *PositionMap) Set(key string, value *StockPosition) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

// Retrieves an element from map under given key.
func (m PositionMap) Get(key string) (*StockPosition, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	return val, ok
}

// Returns the number of elements within the map.
func (m PositionMap) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Looks up an item under specified key
func (m *PositionMap) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// See if element is within shard.
	_, ok := shard.items[key]
	return ok
}

// Removes an element from the map.
func (m *PositionMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

// Checks if map is empty.
func (m *PositionMap) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type TuplePosition struct {
	Key string
	Val *StockPosition
}

// Returns an iterator which could be used in a for range loop.
func (m PositionMap) Iter() <-chan TuplePosition {
	ch := make(chan TuplePosition)
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TuplePosition{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// Returns a buffered iterator which could be used in a for range loop.
func (m PositionMap) IterBuffered() <-chan TuplePosition {
	ch := make(chan TuplePosition, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TuplePosition{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

//Reviles PositionMap "private" variables to json marshal.
func (m PositionMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]*StockPosition)

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

func (m *PositionMap) UnmarshalJSON(b []byte) (err error) {
	// Reverse process of Marshal.

	tmp := make(map[string]*StockPosition)

	// Unmarshal into a single map.
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil
	}

	// foreach key,value pair in temporary map insert into our concurrent map.
	for key, val := range tmp {
		m.Set(key, val)
	}
	return nil
}
