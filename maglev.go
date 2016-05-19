// Package maglev implements maglev consistent hashing
/*
   http://research.google.com/pubs/pub44824.html
*/
package maglev

import "github.com/dchest/siphash"

const (
	SmallM = 65537
	BigM   = 655373
)

type Table struct {
	n      int
	lookup []int
}

func New(names []string, m uint64) *Table {
	permutations := generatePermutations(names, m)
	lookup := populate(permutations)
	return &Table{
		n:      len(names),
		lookup: lookup,
	}
}

func (t *Table) Lookup(key uint64) int {
	return t.lookup[key%uint64(len(t.lookup))]
}

func generatePermutations(names []string, M uint64) [][]uint64 {
	permutations := make([][]uint64, len(names))

	for i, name := range names {
		b := []byte(name)
		h := siphash.Hash(0xdeadbeefcafebabe, 0, b)
		offset, skip := (h>>32)%M, ((h&0xffffffff)%(M-1) + 1)
		p := make([]uint64, M)
		for j := uint64(0); j < M; j++ {
			p[j] = uint64((offset + j*skip) % M)
		}
		permutations[i] = p
	}

	return permutations
}

func populate(permutation [][]uint64) []int {
	M := len(permutation[0])
	N := len(permutation)

	next := make([]uint64, N)
	entry := make([]int, M)
	for j := range entry {
		entry[j] = -1
	}

	var n int
	for {
		for i := 0; i < N; i++ {
			c := permutation[i][next[i]]
			for entry[c] >= 0 {
				next[i]++
				c = permutation[i][next[i]]
			}
			entry[c] = i
			next[i]++
			n++
			if n == M {
				return entry
			}
		}
	}
}
