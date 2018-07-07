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

func New(names []string, M uint64) *Table {
	sortedNames := make([]string, len(names))
	copy(sortedNames, names)
	sort.Strings(sortedNames)
	return &Table{
		names:       sortedNames,
		assignments: populate(names, M, nil),
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
	t.assignments = populate(t.names, uint64(len(t.assignments)), deadIndexes)
}

func permutate(name string, M uint64, cursor uint64) uint64 {
	h := siphash.Hash(0xdeadbeefcafebabe, 0, []byte(name))
	offset, skip := (h>>32)%M, ((h&0xffffffff)%(M-1) + 1)
	return (offset + skip*cursor) % M
}

func populate(names []string, M uint64, dead []int) []int {
	N := len(names)
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
			partition := permutate(names[node], M, cursors[node])
			for assignments[partition] >= 0 {
				cursors[node]++
				partition = permutate(names[node], M, cursors[node])
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
