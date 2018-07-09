// Package maglev implements maglev consistent hashing
/*
   http://research.google.com/pubs/pub44824.html
*/
package maglev

import (
	"sort"

	"github.com/dchest/siphash"
)

const (
	SmallM = 65537
	BigM   = 655373
)

type Table struct {
	names       []string
	assignments []int16
}

type hashed struct {
	offset uint32
	skip   uint32
}

func New(names []string, size uint) *Table {
	t := &Table{
		names:       make([]string, len(names)),
		assignments: make([]int16, size),
	}
	copy(t.names, names)
	sort.Strings(t.names)
	t.assign(t.hashNames())
	return t
}

func (t *Table) Lookup(key uint64) string {
	return t.names[t.assignments[key%uint64(len(t.assignments))]]
}

func (t *Table) Rebuild(dead []string) {
	deadSorted := make([]string, len(dead))
	copy(deadSorted, dead)
	sort.Strings(deadSorted)
	deadIndexes := make([]int, len(dead))
	N := len(t.names)
	nextIndex := 0
	for i, s := range deadSorted {
		for j := nextIndex; j < N; j++ {
			if t.names[j] == s {
				deadIndexes[i] = j
				nextIndex = j + 1
				break
			}
		}
	}
	hashes := t.hashNames()
	t.assign(hashes)
	if len(dead) > 0 {
		t.reassign(hashes, deadIndexes)
	}
}

func (t *Table) m() uint {
	if SmallM < len(t.assignments) {
		return BigM
	}
	return SmallM
}

func (t *Table) hashNames() []hashed {
	M := uint64(t.m())
	hashes := make([]hashed, len(t.names))
	for i, name := range t.names {
		hash := siphash.Hash(0xdeadbeefcafebabe, 0, []byte(name))
		hashes[i].offset, hashes[i].skip = uint32((hash>>32)%M), uint32((hash&0xffffffff)%(M-1)+1)
	}
	return hashes
}

func (t *Table) assign(hashes []hashed) {
	numPartitions := len(t.assignments)
	N := len(hashes)
	assigned := 0
	cursors := make([]uint32, len(hashes))
	for partition := range t.assignments {
		t.assignments[partition] = -1
	}
	for {
		for node := 0; node < N; node++ {
			t.assignments[t.nextAvailablePartition(hashes[node], cursors, node)] = int16(node)
			assigned++
			if assigned == numPartitions {
				return
			}
		}
	}
}

func (t *Table) reassign(hashes []hashed, dead []int) {
	numPartitions := len(t.assignments)
	N := len(hashes)
	assigned := numPartitions
	cursors := make([]uint32, len(hashes))
	deadMap := make(map[int]bool, len(dead))

	for _, node := range dead {
		deadMap[node] = true
	}
	for partition, node := range t.assignments {
		if deadMap[int(node)] {
			t.assignments[partition] = -1
			assigned--
		}
	}
	for {
		d := len(dead) - 1
		for node := N - 1; node >= 0; node-- {
			if d >= 0 && dead[d] == node {
				d--
				continue
			}
			t.assignments[t.nextAvailablePartition(hashes[node], cursors, node)] = int16(node)
			assigned++
			if assigned == numPartitions {
				return
			}
		}
	}
}

func (t *Table) nextAvailablePartition(hash hashed, cursors []uint32, node int) uint {
	numPartitions := uint64(len(t.assignments))
	offset, skip, cursor, M := uint64(hash.offset), uint64(hash.skip), uint64(cursors[node]), uint64(t.m())
	partition := (offset + skip*cursor) % M
	for partition > numPartitions-1 || t.assignments[partition] >= 0 {
		cursor++
		partition = (offset + skip*cursor) % M
	}
	cursor++
	cursors[node] = uint32(cursor)
	return uint(partition)
}
