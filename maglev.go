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
	hashes      []uint64
	assignments []int
}

func New(names []string, M uint64) *Table {
	sortedNames := make([]string, len(names))
	copy(sortedNames, names)
	sort.Strings(sortedNames)
	hashes := make([]uint64, len(names))
	for i, name := range sortedNames {
		hashes[i] = siphash.Hash(0xdeadbeefcafebabe, 0, []byte(name))
	}
	return &Table{
		names:       sortedNames,
		hashes:      hashes,
		assignments: populate(hashes, M, nil),
	}
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
	t.assignments = populate(t.hashes, uint64(len(t.assignments)), deadIndexes)
}

func permute(hash uint64, M uint64, cursor uint64) uint64 {
	offset, skip := (hash>>32)%M, ((hash&0xffffffff)%(M-1) + 1)
	return (offset + skip*cursor) % M
}

func populate(hashes []uint64, M uint64, dead []int) []int {
	N := len(hashes)
	cursors := make([]uint64, N)
	assignments := make([]int, M)
	for partition := range assignments {
		assignments[partition] = -1
	}

	var assigned int
	for {
		d := dead
		for node := 0; node < N; node++ {
			if len(d) > 0 && d[0] == node {
				d = d[1:]
				continue
			}
			partition := permute(hashes[node], M, cursors[node])
			for assignments[partition] >= 0 {
				cursors[node]++
				partition = permute(hashes[node], M, cursors[node])
			}
			assignments[partition] = node
			cursors[node]++
			assigned++
			if uint64(assigned) == M {
				return assignments
			}
		}
	}
}
