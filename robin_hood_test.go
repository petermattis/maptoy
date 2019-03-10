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
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

const benchSize = 1 << 20

func TestRobinHood(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, 4)
	m := newRobinHoodMap(0)
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
		v := new(int)
		*v = i
		m.Put(keys[i], unsafe.Pointer(v))
	}

	fmt.Printf("%s\n", m)
	for i := range keys {
		fmt.Println(m.Get(keys[i]))
	}

	for i := range keys {
		m.Delete(keys[i])
		fmt.Printf("%s\n", m)
	}
}

func BenchmarkHash(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
	}
	b.ResetTimer()

	var h uint32
	for i := 0; i < b.N; {
		n := b.N - i
		if n > len(keys) {
			n = len(keys)
		}
		for j := 0; j < n; j++ {
			h = hash(keys[j], 54)
		}
		i += n
	}

	if testing.Verbose() {
		fmt.Println(h)
	}
}

func BenchmarkGoMapInsert(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
	}
	b.ResetTimer()

	var m map[uint64]unsafe.Pointer
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if m == nil || j == len(keys) {
			b.StopTimer()
			m = make(map[uint64]unsafe.Pointer, len(keys))
			j = 0
			b.StartTimer()
		}
		m[keys[j]] = nil
	}
}

func BenchmarkRobinHoodInsert(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
	}
	v := unsafe.Pointer(new(int))
	b.ResetTimer()

	var m *robinHoodMap
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if m == nil || j == len(keys) {
			b.StopTimer()
			m = newRobinHoodMap(len(keys))
			j = 0
			b.StartTimer()
		}
		m.Put(keys[j], v)
	}
}

func BenchmarkGoMapLookupHit(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	m := make(map[uint64]unsafe.Pointer, len(keys))
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
		m[keys[i]] = nil
	}
	b.ResetTimer()

	var p unsafe.Pointer
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if j == len(keys) {
			j = 0
		}
		p = m[keys[j]]
	}

	if testing.Verbose() {
		fmt.Println(p)
	}
}

func BenchmarkRobinHoodLookupHit(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	m := newRobinHoodMap(len(keys))
	v := unsafe.Pointer(new(int))
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
		m.Put(keys[i], v)
	}
	// fmt.Printf("max: %d avg: %.1f\n", m.MaxDist(), m.AvgDist())
	b.ResetTimer()

	var p unsafe.Pointer
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if j == len(keys) {
			j = 0
		}
		p = m.Get(keys[j])
	}

	if testing.Verbose() {
		fmt.Println(p)
	}
}

func BenchmarkGoMapLookupMiss(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	m := make(map[uint64]unsafe.Pointer, len(keys))
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
		m[keys[i]] = nil
		keys[i] += 1 << 20
	}
	b.ResetTimer()

	var p unsafe.Pointer
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if j == len(keys) {
			j = 0
		}
		p = m[keys[j]]
	}

	if testing.Verbose() {
		fmt.Println(p)
	}
}

func BenchmarkRobinHoodLookupMiss(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]uint64, benchSize)
	m := newRobinHoodMap(len(keys))
	v := unsafe.Pointer(new(int))
	for i := range keys {
		keys[i] = uint64(rng.Intn(1 << 20))
		m.Put(keys[i], v)
		keys[i] += 1 << 20
	}
	b.ResetTimer()

	var p unsafe.Pointer
	for i, j := 0, 0; i < b.N; i, j = i+1, j+1 {
		if j == len(keys) {
			j = 0
		}
		p = m.Get(keys[j])
	}

	if testing.Verbose() {
		fmt.Println(p)
	}
}
