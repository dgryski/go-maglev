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
	mod         uint64
	strength    int
	hashes      [][]hash
}

type hash struct {
	offset, skip uint32
}

func hashString(s string, seed uint64) uint64 {
	return siphash.Hash(0xdeadbeefcafebabe, seed, []byte(s))
}

func sortNames(names []string) {
	sort.Slice(names, func(i, j int) bool {
		hi, hj := hashString(names[i], 0), hashString(names[j], 0)
		return hi < hj || (hi == hj && names[i] < names[j])
	})
}

func nextPrime(num uint) uint {
	num++
	for !isPrime(num) {
		num++
	}
	return num
}

func isPrime(n uint) bool {
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	i, w := uint(5), uint(2)
	for i*i <= n {
		if n%i == 0 {
			return false
		}
		i += w
		w = 6 - w
	}
	return true
}

func New(names []string, size uint) *Table {
	return NewWithPermutationStrength(names, size, 3)
}

func NewWithPermutationStrength(names []string, size uint, strength int) *Table {
	if strength < 1 {
		strength = 1
	}
	M := uint64(nextPrime(size - 1))
	t := &Table{
		names:       append([]string{}, names...),
		hashes:      make([][]hash, len(names)),
		assignments: make([]int16, size),
		mod:         M,
		strength:    strength,
	}
	sortNames(t.names)
	for i, name := range t.names {
		t.hashes[i] = make([]hash, strength)
		for j := 0; j < strength; j++ {
			h := hashString(name, uint64(j))
			t.hashes[i][j] = hash{uint32((h >> 32) % M), uint32((h&0xffffffff)%(M-1) + 1)}
		}
	}
	t.assign()
	return t
}

func (t *Table) Lookup(key uint64) string {
	return t.names[t.assignments[key%uint64(len(t.assignments))]]
}

func (t *Table) Rebuild(dead []string) {
	t.assign()
	if len(dead) == 0 {
		return
	}
	deadSorted := append([]string{}, dead...)
	sortNames(deadSorted)
	deadIndexes := make([]int, len(deadSorted))
	N := len(t.names)
	nextIndex := 0
	found := 0
	for i, deadNode := range deadSorted {
		for j := nextIndex; j < N && found < len(deadSorted); j++ {
			if t.names[j] == deadNode {
				deadIndexes[i] = j
				nextIndex = j + 1
				found++
				break
			}
		}
	}
	t.reassign(deadIndexes)
}

func (t *Table) assign() {
	numPartitions := len(t.assignments)
	N := len(t.names)
	assigned := 0
	cursors := make([]uint32, N)
	for partition := range t.assignments {
		t.assignments[partition] = -1
	}
	for {
		for node := 0; node < N; node++ {
			t.assignments[t.nextAvailablePartition(cursors, node)] = int16(node)
			assigned++
			if assigned == numPartitions {
				return
			}
		}
	}
}

func (t *Table) reassign(dead []int) {
	numPartitions := len(t.assignments)
	N := len(t.names)
	assigned := numPartitions
	cursors := make([]uint32, N)
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
			t.assignments[t.nextAvailablePartition(cursors, node)] = int16(node)
			assigned++
			if assigned == numPartitions {
				return
			}
		}
	}
}

func (t *Table) nextAvailablePartition(cursors []uint32, node int) uint {
	numPartitions := uint(len(t.assignments))
	partition := t.permute(cursors[node], t.hashes[node])
	cursors[node]++
	for partition >= numPartitions || t.assignments[partition] >= 0 {
		partition = t.permute(cursors[node], t.hashes[node])
		cursors[node]++
	}
	return partition
}

func (t *Table) permute(cursor uint32, hashes []hash) uint {
	c := uint64(cursor)
	for _, h := range hashes {
		c = (uint64(h.offset) + uint64(h.skip)*c) % t.mod
	}
	return uint(c)
}
