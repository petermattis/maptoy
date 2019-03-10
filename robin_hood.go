// Copyright 2019 Peter Mattis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.package maptoy

package maptoy

import (
	"fmt"
	"math/bits"
	"strings"
	"unsafe"
)

// Fibonacci hash: https://probablydance.com/2018/06/16/fibonacci-hashing-the-optimization-that-the-world-forgot-or-a-better-alternative-to-integer-modulo/
func hash(k uint64, shift uint32) uint32 {
	k |= 1
	return uint32((k * 11400714819323198485) >> shift)
}

type robinHoodEntry struct {
	key   uint64
	value unsafe.Pointer
	dist  uint32
}

// robinHoodMap is an implementation of Robin Hood hashing. Robin Hood hashing
// is an open-address hash table using linear probing. The twist is that the
// linear probe distance is reduced by moving existing entries when inserting
// and deleting. This is accomplished by keeping track of how far an entry is
// from its "desired" slot (hash of key modulo number of slots). During
// insertion, if the new entry being inserted is farther from its desired slot
// than the target entry, we swap the target and new entry. This effectively
// steals from the "rich" target entry and gives to the "poor" new entry (thus
// the origin of the name).
//
// An extension over the base Robin Hood hashing idea comes from
// https://probablydance.com/2017/02/26/i-wrote-the-fastest-hashtable/. A cap
// is placed on the max distance an entry can be from its desired slot. When
// this threshold is reached during insertion, the size of the table is doubled
// and insertion is restarted. Additionally, the entries slice is given "max
// dist" extra entries on the end. The very last entry in the entries slice is
// never used and acts as a sentinel which terminates loops. The previous
// maxDist-1 entries act as the extra entries. For example, if the size of the
// table is 2, maxDist is computed as 4 and the actual size of the entry slice
// is 6.
//
//   +---+---+---+---+---+---+
//   | 0 | 1 | 2 | 3 | 4 | 5 |
//   +---+---+---+---+---+---+
//           ^
//          size
//
// In this scenario, the target entry for a key will always be in the range
// [0,1]. Valid entries may reside in the range [0,4] due to the linear probing
// of up to maxDist entries. The entry at index 5 will never contain a value,
// and instead acts as a sentinel (its distance is always 0). The max distance
// threshold is set to log2(num-entries). This ensures that retrieval is O(log
// N), though note that N is the number of total entries, not the count of
// valid entries.
//
// Deletion is implemented via the backward shift delete mechanism instead of
// tombstones. This preserves the performance of the table in the presence of
// deletions. See
// http://codecapsule.com/2013/11/17/robin-hood-hashing-backward-shift-deletion
// for details.
type robinHoodMap struct {
	entries    []robinHoodEntry
	entriesPtr unsafe.Pointer
	size       uint32
	shift      uint32
	count      uint32
	maxDist    uint32
}

func maxDistForSize(size uint32) uint32 {
	desired := uint32(bits.Len32(size))
	if desired < 4 {
		desired = 4
	}
	return desired
}

func newRobinHoodMap(initialCapacity int) *robinHoodMap {
	if initialCapacity < 1 {
		initialCapacity = 1
	}
	targetSize := 1 << uint(bits.Len(uint(2*initialCapacity-1)))

	m := &robinHoodMap{}
	m.rehash(uint32(targetSize))
	return m
}

func (m *robinHoodMap) rehash(size uint32) {
	oldEntries := m.entries
	m.size = size
	m.shift = uint32(64 - bits.Len32(m.size-1))
	m.maxDist = maxDistForSize(size)
	m.entries = make([]robinHoodEntry, size+m.maxDist)
	m.entriesPtr = unsafe.Pointer(&m.entries[0])
	m.count = 0

	for i := range oldEntries {
		e := &oldEntries[i]
		if e.value != nil {
			m.Put(e.key, e.value)
		}
	}
}

func (m *robinHoodMap) entry(i uint32) *robinHoodEntry {
	// Manually index into the entries array to avoid the bounds checking.
	return (*robinHoodEntry)(unsafe.Pointer(uintptr(m.entriesPtr) + uintptr(i)*unsafe.Sizeof(robinHoodEntry{})))
}

func (m *robinHoodMap) Put(k uint64, v unsafe.Pointer) {
	n := robinHoodEntry{key: k, value: v, dist: 0}
	for i := hash(n.key, m.shift); ; i++ {
		e := m.entry(i)
		if e.value == nil {
			// Found an empty entry: insert here.
			*e = n
			m.count++
			return
		}

		if e.dist < n.dist {
			// Swap the new entry with the current entry because the current is
			// rich. We then continue to loop, looking for a new location for the
			// current entry.
			n, *e = *e, n
		}

		// The new entry gradually moves away from its ideal position.
		n.dist++

		// If we've reached the max distance threshold, grow the table and restart
		// the insertion.
		if n.dist == m.maxDist {
			m.rehash(2 * m.size)
			i = hash(e.key, m.shift) - 1
			n.dist = 0
		}
	}
}

func (m *robinHoodMap) Get(k uint64) unsafe.Pointer {
	var dist uint32
	for i := hash(k, m.shift); ; i++ {
		e := m.entry(i)
		if k == e.key {
			// Found.
			return e.value
		}
		if dist > e.dist {
			// Not found.
			return nil
		}
		dist++
	}
}

func (m *robinHoodMap) Delete(k uint64) {
	var dist uint32
	for i := hash(k, m.shift); ; i++ {
		e := m.entry(i)
		if k == e.key {
			// We found the entry to delete. Shift the following entries backwards
			// until the next empty value or entry with a zero distance. Note that
			// empty values are guaranteed to have "dist == 0".
			m.count--
			for j := i + 1; ; j++ {
				t := m.entry(j)
				if t.dist == 0 {
					*e = robinHoodEntry{}
					return
				}
				e.key = t.key
				e.value = t.value
				e.dist = t.dist - 1
				e = t
			}
		}
		if dist > e.dist {
			// Not found.
			return
		}
		dist++
	}
}

func (m *robinHoodMap) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "count: %d\n", m.count)
	for _, v := range m.entries {
		fmt.Fprintf(&buf, "[%v,%v,%d]\n", v.key, v.value, v.dist)
	}
	return buf.String()
}
