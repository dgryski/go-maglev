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
	hashes      []hashed
	assignments []int
}

type hashed struct {
	offset uint32
	skip   uint32
}

func New(names []string, M uint) *Table {
	t := &Table{
		names:       make([]string, len(names)),
		hashes:      make([]hashed, len(names)),
		assignments: make([]int, M),
	}
	copy(t.names, names)
	sort.Strings(t.names)
	for i, name := range t.names {
		hash := siphash.Hash(0xdeadbeefcafebabe, 0, []byte(name))
		t.hashes[i].offset, t.hashes[i].skip = uint32((hash>>32)%uint64(M)), uint32((hash&0xffffffff)%(uint64(M)-1)+1)
	}
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

func permute(h hashed, M uint32, cursor uint32) uint32 {
	return (h.offset + h.skip*cursor) % M
}

func (t *Table) populate(dead []int) {
	M := uint32(len(t.assignments))
	N := len(t.names)
	cursors := make([]uint32, N)
	for partition := range t.assignments {
		t.assignments[partition] = -1
	}

	var assigned uint32
	for {
		d := dead
		for node := 0; node < N; node++ {
			if len(d) > 0 && d[0] == node {
				d = d[1:]
				continue
			}
			partition := permute(t.hashes[node], M, cursors[node])
			for t.assignments[partition] >= 0 {
				cursors[node]++
				partition = permute(t.hashes[node], M, cursors[node])
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
