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
	names             []string
	assignments       []int16
	mod               uint64
	hashKey           uint64
	permutationRounds int
	dead              []int
	hashes            [][]hash
}

type hash struct {
	offset, skip uint32
}

type options struct {
	permutationRounds int
	mod               uint64
	hashKey           uint64
}

type Option func(*options)

func WithPermutationRounds(rounds int) Option {
	return func(opts *options) { opts.permutationRounds = rounds }
}

func WithModulus(mod uint64) Option {
	return func(opts *options) { opts.mod = mod }
}

func WithHashKey(key uint64) Option {
	return func(opt *options) { opt.hashKey = key }
}

func hashString(s string, hashKey uint64, seed uint64) uint64 {
	return siphash.Hash(hashKey, seed, []byte(s))
}

func sortNames(names []string, hashKey uint64) {
	sort.Slice(names, func(i, j int) bool {
		hi, hj := hashString(names[i], hashKey, 0), hashString(names[j], hashKey, 0)
		return hi < hj || (hi == hj && names[i] < names[j])
	})
}

func permute(cursor uint32, hashes []hash, mod uint64) uint {
	c := uint64(cursor)
	for _, h := range hashes {
		c = (uint64(h.offset) + uint64(h.skip)*c) % mod
	}
	return uint(c)
}

func deduplicate(existing []string, extra []string, excludeExisting bool) []string {
	sliceFrom := 0
	if excludeExisting {
		sliceFrom = len(existing)
	}
	m := make(map[string]bool, len(extra))
	for _, name := range extra {
		m[name] = true
	}
	found := 0
	for _, name := range existing {
		if m[name] {
			m[name] = false
			found++
		}
	}
	if found == len(extra) {
		return existing[sliceFrom:]
	}
	deduped := make([]string, 0, len(existing)+len(extra)-found-sliceFrom)
	deduped = append(deduped, existing[sliceFrom:]...)
	for name, ok := range m {
		if ok {
			deduped = append(deduped, name)
		}
	}
	return deduped
}

func nextPrime(n uint) uint {
outer:
	for {
		n++
		if n%2 == 0 || n%3 == 0 {
			continue
		}
		i, w := uint(5), uint(2)
		for i*i <= n {
			if n%i == 0 {
				continue outer
			}
			i += w
			w = 6 - w
		}
		return n
	}
}

func New(names []string, partitions uint, args ...Option) *Table {
	opts := &options{
		permutationRounds: 3,
		mod:               SmallM,
		hashKey:           0xdeadbeefcafebabe,
	}
	for _, arg := range args {
		arg(opts)
	}
	if opts.mod < uint64(partitions) {
		opts.mod = uint64(nextPrime(partitions - 1))
	}
	t := &Table{
		names:             append([]string{}, names...),
		assignments:       make([]int16, partitions),
		mod:               opts.mod,
		hashKey:           opts.hashKey,
		permutationRounds: opts.permutationRounds,
		hashes:            make([][]hash, len(names)),
	}
	t.initialize()
	t.assign()
	return t
}

func (t *Table) Lookup(key uint64) string {
	return t.names[t.assignments[key%uint64(len(t.assignments))]]
}

func (t *Table) PartitionOwner(partition int) string {
	return t.Lookup(uint64(partition))
}

func (t *Table) Add(names ...string) {
	originalNameCount := len(t.names)
	t.names = deduplicate(t.names, names, false)
	if len(t.names) == originalNameCount {
		return
	}
	t.initialize()
	t.Rebuild(deduplicate(names, t.getDeadNames(), true))
}

func (t *Table) Remove(names ...string) {
	originalDeadCount := len(t.dead)
	dead := deduplicate(t.getDeadNames(), names, false)
	if len(dead) == originalDeadCount {
		return
	}
	t.Rebuild(dead)
}

func (t *Table) Rebuild(dead []string) {
	t.assign()
	if len(dead) == 0 {
		return
	}
	sorted := make([]string, len(dead))
	copy(sorted, dead)
	sortNames(sorted, t.hashKey)
	indexedDead := make([]int, len(dead))
	N := len(t.names)
	nextIndex := 0
	found := 0
	for i, deadNode := range sorted {
		for j := nextIndex; j < N && found < len(dead); j++ {
			if t.names[j] == deadNode {
				indexedDead[i] = j
				nextIndex = j + 1
				found++
				break
			}
		}
	}
	t.dead = indexedDead
	t.reassign()
}

func (t *Table) initialize() {
	sortNames(t.names, t.hashKey)
	if len(t.hashes) != len(t.names) {
		t.hashes = make([][]hash, len(t.names))
	}
	for i, name := range t.names {
		if t.hashes[i] == nil || len(t.hashes[i]) != t.permutationRounds {
			t.hashes[i] = make([]hash, t.permutationRounds)
		}
		for j := 0; j < t.permutationRounds; j++ {
			h64 := hashString(name, t.hashKey, uint64(j))
			t.hashes[i][j] = hash{
				offset: uint32((h64 >> 32) % t.mod),
				skip:   uint32((h64&0xffffffff)%(t.mod-1) + 1),
			}
		}
	}
}

func (t *Table) getDeadNames() []string {
	if len(t.dead) == 0 {
		return nil
	}
	deadNames := make([]string, len(t.dead))
	for i, node := range t.dead {
		deadNames[i] = t.names[node]
	}
	return deadNames
}

func (t *Table) assign() {
	assigned := 0
	cursors := make([]uint32, len(t.names))
	for partition := range t.assignments {
		t.assignments[partition] = -1
	}
	for {
		for node := 0; node < len(t.names); node++ {
			t.assignments[t.nextAvailablePartition(cursors, node)] = int16(node)
			assigned++
			if assigned == len(t.assignments) {
				return
			}
		}
	}
}

func (t *Table) reassign() {
	assigned := len(t.assignments)
	cursors := make([]uint32, len(t.names))
	deadMap := make(map[int]bool, len(t.dead))
	for _, node := range t.dead {
		deadMap[node] = true
	}
	for partition, node := range t.assignments {
		if deadMap[int(node)] {
			t.assignments[partition] = -1
			assigned--
		}
	}
	for {
		d := len(t.dead) - 1
		for node := len(t.names) - 1; node >= 0; node-- {
			if d >= 0 && t.dead[d] == node {
				d--
				continue
			}
			t.assignments[t.nextAvailablePartition(cursors, node)] = int16(node)
			assigned++
			if assigned == len(t.assignments) {
				return
			}
		}
	}
}

func (t *Table) nextAvailablePartition(cursors []uint32, node int) uint {
	partition := permute(cursors[node], t.hashes[node], t.mod)
	cursors[node]++
	for partition >= uint(len(t.assignments)) || t.assignments[partition] >= 0 {
		partition = permute(cursors[node], t.hashes[node], t.mod)
		cursors[node]++
	}
	return partition
}
