// Swiss map is an efficient hash map implementation based on the SwissTable
// algorithm. This design improves upon traditional hash tables by optimizing
// for CPU cache usage and reducing the number of memory accesses during
// lookups. Unlike the original SwissTable, this Go implementation does not
// leverage SIMD instructions but still uses a similar strategy for efficient
// probing and key matching through control bytes.

// In this implementation, the hash map is divided into groups of slots, with
// each group containing 8 slots. The control bytes represent the state of
// each slot, indicating whether it is empty, full, or deleted. These control
// bytes help speed up the probing process by reducing the number of key
// comparisons required.

// For each key, the hash is split into two parts: h1 and h2. The h1 value is
// used to determine the index of the group, while h2 is compared with the
// control bytes to find matching or available slots. The control bytes store
// a part of the h2 value, allowing the map to quickly identify non-matching
// slots and avoid unnecessary key comparisons.

// When inserting a key-value pair, the map first probes the group identified
// by h1. If no empty or deleted slot is found in the group, it moves to the
// next group, continuing the probing process until a suitable slot is found.
// Deletions mark slots as "tombstones" using a special deleted value, and
// rehashing is triggered when tombstones accumulate beyond a certain threshold.

// The map’s design reduces cache misses and optimizes memory usage by keeping
// related slots close together and minimizing the number of memory accesses
// required for common operations like insertions, lookups, and deletions.
// Though this implementation does not use SIMD, it still benefits from the
// SwissTable's overall strategy for fast and cache-friendly hash table
// operations.

package swiss

import (
	"iter"
	"math/bits"
	"math/rand"
	"unsafe"

	"github.com/crn4/swiss/hash"
)

const (
	kEmpty    = 0b10000000 // -127
	kDeleted  = 0b11111110 // -2
	kSentinel = 0b11111111 // -1
	// kFull = 0b0xxxxxxx // hash bytes

	kMsbsBytes = 0x8080808080808080
	kLsbsBytes = 0x0101010101010101

	emptyContol = kMsbsBytes

	grpssz   = 8
	grpload  = 7
	maxloadf = grpload / grpssz
)

type Map[K comparable, V any] struct {
	grps       []group[K, V]
	hashfn     hash.HFunc
	seed       uintptr
	len        int
	cap        int
	tombstones int
	ngroups    uint32
}

type group[K comparable, V any] struct {
	cntrl control
	slts  [grpssz]slot[K, V]
}

type slot[K comparable, V any] struct {
	key   K
	value V
}

type control uint64

// New creates a new Swiss map with the specified initial size. It preallocates
// the necessary number of groups and sets up the hash function. The control
// bytes of each group are initialized to an empty state (kEmpty). The hash
// function and seed are also initialized. The capacity is calculated based
// on the number of groups and the load factor.
func New[K comparable, V any](size int) *Map[K, V] {
	ngroups := groupsnum(size)
	m := &Map[K, V]{
		grps:    make([]group[K, V], ngroups),
		ngroups: uint32(ngroups),
		// hashfn:  getHashFunc[K](),
		hashfn: hash.GetHashFunc[K](),
		seed:   uintptr(rand.Uint64()),
		cap:    grpload * ngroups,
	}
	m.groups(func(g *group[K, V]) bool {
		g.cntrl = emptyContol
		return true
	})
	return m
}

// Put inserts or updates a key-value pair in the map. It calculates the hash
// of the key and uses h1 to locate the appropriate group. The function probes
// the group for a matching key or an empty/deleted slot. If the key is found,
// its value is updated. If an empty or deleted slot is found, the key-value
// pair is inserted. Rehashing occurs if the map's load exceeds the capacity.
func (m *Map[K, V]) Put(key K, value V) {
	hash := m.hashfn(noescape(unsafe.Pointer(&key)), m.seed)
	ngrp := uint32(h1(hash)) % m.ngroups
	for {
		group := &m.grps[ngrp]
		equal := group.match(h2(hash))
		for equal != 0 {
			i := equal.first()
			if key == group.slts[i].key {
				group.slts[i].value = value
				return
			}
			equal = equal.rmfirst()
		}
		if empty := group.maskEmptyOrDeleted(); empty != 0 {
			i := empty.first()
			group.slts[i] = slot[K, V]{key: key, value: value}
			group.cntrl.set(i, uint8(h2(hash)))
			m.len++
			if m.len > m.cap {
				m.rehash()
			}
			return
		}
		ngrp++
		if ngrp >= m.ngroups {
			ngrp = 0
		}
	}
}

// Get retrieves the value associated with a given key. It calculates the hash
// of the key and uses h1 to find the corresponding group. The function checks
// the control bytes of the group for a matching h2. If a match is found, it
// compares the key and returns the value. If the key is not found or an empty
// slot is encountered, the function returns false.
func (m *Map[K, V]) Get(key K) (V, bool) {
	hash := m.hashfn(noescape(unsafe.Pointer(&key)), m.seed)
	ngrp := uint32(h1(hash)) % m.ngroups
	for {
		group := &m.grps[ngrp]
		equal := group.match(h2(hash))
		for equal != 0 {
			i := equal.first()
			if key == group.slts[i].key {
				return group.slts[i].value, true
			}
			equal = equal.rmfirst()
		}
		if group.maskEmpty() != 0 {
			var res V
			return res, false
		}
		ngrp++
		if ngrp >= m.ngroups {
			ngrp = 0
		}
	}
}

// Delete removes a key-value pair from the map. If the key is found, the
// slot is cleared, and the control byte is marked as either empty or deleted
// (tombstone). This optimization helps avoid wasting slots if there are
// empty slots available in the group. Tombstones are tracked and used to
// trigger rehashing when necessary.
func (m *Map[K, V]) Delete(key K) {
	hash := m.hashfn(noescape(unsafe.Pointer(&key)), m.seed)
	ngrp := uint32(h1(hash)) % m.ngroups
	for {
		group := &m.grps[ngrp]
		equal := group.match(h2(hash))
		for equal != 0 {
			i := equal.first()
			if key == group.slts[i].key {
				group.slts[i] = slot[K, V]{}
				if group.maskEmpty() != 0 {
					group.cntrl.set(i, kEmpty)
					m.len--
				} else {
					group.cntrl.set(i, kDeleted)
					m.tombstones++
				}
				return
			}
			equal = equal.rmfirst()
		}
		if group.maskEmpty() != 0 {
			return
		}
		ngrp++
		if ngrp >= m.ngroups {
			ngrp = 0
		}
	}
}

// Clear removes all key-value pairs from the map, resetting all groups to an
// empty state. The capacity remains unchanged, but the length and tombstones
// are reset to zero.
func (m *Map[K, V]) Clear() {
	m.len, m.tombstones = 0, 0
	for i := range m.grps {
		m.grps[i].cntrl = emptyContol
		for j := range m.grps[i].slts {
			m.grps[i].slts[j] = slot[K, V]{}
		}
	}
}

// Len returns the number of key-value pairs currently stored in the map,
// excluding deleted (tombstone) entries.
func (m *Map[K, V]) Len() int {
	return m.len - m.tombstones
}

// Cap returns the map’s capacity, which is based on the number of groups and
// the load factor.
func (m *Map[K, V]) Cap() int {
	return m.cap
}

func (m *Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		groups := m.grps
		for i := range groups {
			mask := groups[i].maskFull()
			for mask != 0 {
				j := mask.first()
				if !yield(groups[i].slts[j].key, groups[i].slts[j].value) {
					return
				}
				mask = mask.rmfirst()
			}
		}
	}
}

// rehash reorganizes the map by creating new groups and reinserting all
// non-deleted entries. It calculates the new capacity and resets tombstones.
// The function is triggered when the map reaches a certain load factor or
// when tombstones accumulate excessively.
func (m *Map[K, V]) rehash() {
	newsize := newsize(m.cap, m.tombstones)
	groups := m.grps
	ngroups := groupsnum(newsize)
	m.grps = make([]group[K, V], ngroups)
	m.ngroups = uint32(ngroups)
	m.cap = ngroups * grpload
	m.len, m.tombstones = 0, 0
	m.groups(func(g *group[K, V]) bool {
		g.cntrl = emptyContol
		return true
	})
	for i := range groups {
		mask := groups[i].maskFull()
		for mask != 0 {
			j := mask.first()
			m.Put(groups[i].slts[j].key, groups[i].slts[j].value)
			mask = mask.rmfirst()
		}
	}
}

func newsize(oldsize, tombstones int) int {
	if tombstones >= oldsize/2 {
		return oldsize
	}
	return oldsize * 2
}

func (m *Map[K, V]) groups(yield func(g *group[K, V]) bool) {
	for i := range m.grps {
		if !yield(&m.grps[i]) {
			break
		}
	}
}

func (c *control) set(i uint32, value uint8) {
	*(*uint8)(unsafe.Add(unsafe.Pointer(c), i)) = value
}

type bitmask uint64

func (g *group[K, V]) match(h2 uintptr) bitmask {
	// https://github.com/abseil/abseil-cpp/blob/master/absl/container/internal/raw_hash_set.h#L842
	x := uint64(g.cntrl) ^ (kLsbsBytes * uint64(h2))
	return bitmask(((x - kLsbsBytes) &^ x) & kMsbsBytes)
}

// maskEmpty returns a bitmask representing the positions of empty slots
func (g *group[K, V]) maskEmpty() bitmask {
	return bitmask((g.cntrl &^ (g.cntrl << 6)) & kMsbsBytes)
}

// maskFull returns a bitmask representing the positions of full slots
func (g *group[K, V]) maskFull() bitmask {
	return bitmask((g.cntrl ^ kMsbsBytes) & kMsbsBytes)
}

// maskNonFull returns a bitmask representing the positions of non full slots
func (g *group[K, V]) maskNonFull() bitmask {
	return bitmask(g.cntrl & kMsbsBytes)
}

func (g *group[K, V]) maskEmptyOrDeleted() bitmask {
	return bitmask((g.cntrl &^ (g.cntrl << 7)) & kMsbsBytes)
}

func (b bitmask) first() uint32 {
	return uint32(bits.TrailingZeros64(uint64(b))) >> 3
}

func (b bitmask) rmfirst() bitmask {
	return b & (b - 1)
}

// groupsnum calculates the required number of groups based on the requested
// size, accounting for the load factor.
func groupsnum(n int) int {
	if n == 0 {
		n = 10
	}
	return (n + grpload + 1) / grpload
}

// h1 and h2 split the hash value into two parts. h1 determines the group,
// while h2 is used for matching the control bytes within that group.
func h1(hash uintptr) uintptr {
	return hash >> 7
}

func h2(hash uintptr) uintptr {
	return hash & 0x7F
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input. noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime; see issues 23382 and 7921.
//
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

// find isn't used in the code, as it's inlined, but kept here for informational purposes only
func (m *Map[K, V]) find(key K, hash uintptr) (uint32, uint32, bool) {
	ngrp := uint32(h1(hash)) % m.ngroups
	for {
		group := &m.grps[ngrp]
		equal := group.match(h2(hash))
		for equal != 0 {
			i := equal.first()
			if key == group.slts[i].key {
				return ngrp, i, true
			}
			equal = equal.rmfirst()
		}
		if empty := group.maskEmpty(); empty != 0 {
			return ngrp, empty.first(), false
		}
		ngrp++
		if ngrp >= m.ngroups {
			ngrp = 0
		}
	}
}
