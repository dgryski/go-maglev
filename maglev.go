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
	assignments []int
}

type hashed struct {
	offset uint32
	skip   uint32
}

func New(names []string, M uint) *Table {
	t := &Table{
		names:       make([]string, len(names)),
		assignments: make([]int, M),
	}
	copy(t.names, names)
	sort.Strings(t.names)
	t.populate(nil)
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
	t.populate(deadIndexes)
}

func (t *Table) hashNames() []hashed {
	M := uint64(len(t.assignments))
	hashes := make([]hashed, len(t.names))
	for i, name := range t.names {
		hash := siphash.Hash(0xdeadbeefcafebabe, 0, []byte(name))
		hashes[i].offset, hashes[i].skip = uint32((hash>>32)%uint64(M)), uint32((hash&0xffffffff)%(uint64(M)-1)+1)
	}
	return hashes
}

func (t *Table) populate(dead []int) {
	M := uint64(len(t.assignments))
	for partition := range t.assignments {
		t.assignments[partition] = -1
	}
	hashes := t.hashNames()
	t.populateOnce(hashes, nil, 0)
	if len(dead) == 0 {
		return
	}
	t.populateOnce(hashes, dead, M-t.unassign(dead))
}

func (t *Table) populateOnce(hashes []hashed, dead []int, assigned uint64) {
	M := uint64(len(t.assignments))
	N := len(hashes)
	cursors := make([]uint32, len(hashes))
	for {
		d := 0
		for node := 0; node < N; node++ {
			if d < len(dead) && dead[d] == node {
				d++
				continue
			}
			offset, skip := uint64(hashes[node].offset), uint64(hashes[node].skip)
			partition := (offset + skip*uint64(cursors[node])) % M
			for t.assignments[partition] >= 0 {
				cursors[node]++
				partition = (offset + skip*uint64(cursors[node])) % M
			}
			t.assignments[partition] = node
			cursors[node]++
			assigned++
			if assigned == M {
				return
			}
		}
	}
}

func (t *Table) unassign(dead []int) uint64 {
	deadMap := make(map[int]bool, len(dead))
	for _, node := range dead {
		deadMap[node] = true
	}
	var unassigned uint64
	for assignmentPartition, assignedNode := range t.assignments {
		if deadMap[assignedNode] {
			t.assignments[assignmentPartition] = -1
			unassigned++
		}
	}
	return unassigned
}
