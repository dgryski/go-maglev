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
	nodes       []hashed
	assignments []int16
	mod         uint64
	strength    int
}

type hashed struct {
	name   string
	hash   uint64
	offset uint32
	skip   uint32
}

func hashNames(names []string, M uint64) []hashed {
	hashes := make([]hashed, len(names))
	for i, name := range names {
		hash := siphash.Hash(0xdeadbeefcafebabe, 0, []byte(name))
		hashes[i].name, hashes[i].hash = name, hash
		hashes[i].offset, hashes[i].skip = uint32((hash>>32)%M), uint32((hash&0xffffffff)%(M-1)+1)
	}
	return hashes
}

func sortNodes(nodes []hashed) {
	sort.Slice(nodes, func(i, j int) bool {
		hi, hj := nodes[i].hash, nodes[j].hash
		return hi < hj || (hi == hj && nodes[i].name < nodes[j].name)
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
		nodes:       hashNames(names, M),
		assignments: make([]int16, size),
		mod:         M,
		strength:    strength,
	}
	sortNodes(t.nodes)
	t.assign()
	return t
}

func (t *Table) Lookup(key uint64) string {
	return t.nodes[t.assignments[key%uint64(len(t.assignments))]].name
}

func (t *Table) Rebuild(dead []string) {
	deadNodes := hashNames(dead, t.mod)
	sortNodes(deadNodes)
	deadIndexes := make([]int, len(deadNodes))
	N := len(t.nodes)
	nextIndex := 0
	for i, deadNode := range deadNodes {
		for j := nextIndex; j < N; j++ {
			if t.nodes[j] == deadNode {
				deadIndexes[i] = j
				nextIndex = j + 1
				break
			}
		}
	}
	t.assign()
	if len(dead) > 0 {
		t.reassign(deadIndexes)
	}
}

func (t *Table) assign() {
	numPartitions := len(t.assignments)
	N := len(t.nodes)
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
	N := len(t.nodes)
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
	partition := t.permute(cursors[node], node)
	cursors[node]++
	for partition > numPartitions-1 || t.assignments[partition] >= 0 {
		partition = t.permute(cursors[node], node)
		cursors[node]++
	}
	return partition
}

func (t *Table) permute(cursor uint32, node int) uint {
	c := uint64(cursor)
	for round := 0; round < t.strength; round++ {
		h := t.nodes[(node+round)%len(t.nodes)]
		c = (uint64(h.offset) + uint64(h.skip)*c) % t.mod
	}
	return uint(c)
}
