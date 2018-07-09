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
	nodes       []hashedName
	assignments []int16
	mod         uint64
	strength    int
}

type hashedName struct {
	name   string
	hashes []hash
}

type hash struct {
	hash         uint64
	offset, skip uint32
}

func hashNames(names []string, M uint64, strength int) []hashedName {
	hashedNames := make([]hashedName, len(names))
	for i, name := range names {
		hashedNames[i].name = name
		hashedNames[i].hashes = make([]hash, strength)
		for j := 0; j < strength; j++ {
			h := siphash.Hash(0xdeadbeefcafebabe, uint64(j), []byte(name))
			hashedNames[i].hashes[j] = hash{h, uint32((h >> 32) % M), uint32((h&0xffffffff)%(M-1) + 1)}
		}
	}
	return hashedNames
}

func sortNodes(nodes []hashedName) {
	sort.Slice(nodes, func(i, j int) bool {
		hi, hj := nodes[i].hashes[0].hash, nodes[j].hashes[0].hash
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
		nodes:       hashNames(names, M, strength),
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
	deadNodes := hashNames(dead, t.mod, t.strength)
	sortNodes(deadNodes)
	deadIndexes := make([]int, len(deadNodes))
	N := len(t.nodes)
	nextIndex := 0
	for i, deadNode := range deadNodes {
		for j := nextIndex; j < N; j++ {
			if t.nodes[j].name == deadNode.name {
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
	partition := t.permute(cursors[node], t.nodes[node])
	cursors[node]++
	for partition > numPartitions-1 || t.assignments[partition] >= 0 {
		partition = t.permute(cursors[node], t.nodes[node])
		cursors[node]++
	}
	return partition
}

func (t *Table) permute(cursor uint32, node hashedName) uint {
	c := uint64(cursor)
	for _, h := range node.hashes {
		c = (uint64(h.offset) + uint64(h.skip)*c) % t.mod
	}
	return uint(c)
}
