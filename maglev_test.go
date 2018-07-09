package maglev

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestDistribution(t *testing.T) {
	const size = 125

	var names []string
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("backend-%d", i))
	}

	table := New(names, 1<<13)

	r := make(map[string]int, size)
	rand.Seed(0)
	for i := 0; i < 1e6; i++ {
		name := table.Lookup(uint64(rand.Int63()))
		r[name]++
	}

	var total int
	var max int
	for _, v := range r {
		total += v
		if v > max {
			max = v
		}
	}

	mean := float64(total) / size
	t.Logf("max=%v, mean=%v, peak-to-mean=%v", max, mean, float64(max)/mean)

	r = make(map[string]int, size)
	for _, node := range table.assignments {
		r[table.nodes[node].name]++
	}

	max = 0
	min := 1 << 30
	for _, v := range r {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	t.Logf("max-assignment=%v, min-assignment=%v max-to-min=%v", max, min, float64(max)/float64(min))

	originalAssignments := make([]int16, len(table.assignments))
	copy(originalAssignments, table.assignments)

	table.Rebuild([]string{"backend-13"})

	var reassigned int
	for partition, node := range table.assignments {
		if originalAssignments[partition] != node {
			reassigned++
		}
	}

	t.Logf("reassigned=%v/%v=%v", reassigned, len(originalAssignments), float64(reassigned)/float64(len(originalAssignments)))

	r = make(map[string]int, size)
	for i := 0; i < 1e6; i++ {
		name := table.Lookup(uint64(rand.Int63()))
		r[name]++
	}

	total = 0
	max = 0
	for _, v := range r {
		total += v
		if v > max {
			max = v
		}
	}

	mean = float64(total) / size
	t.Logf("max=%v, mean=%v, peak-to-mean=%v", max, mean, float64(max)/mean)

	r = make(map[string]int, size)
	for _, node := range table.assignments {
		if table.nodes[node].name == "backend-13" {
			t.Fatal("Dead node was not reassigned after rebuild")
		}
		r[table.nodes[node].name]++
	}

	max = 0
	min = 1 << 30
	for _, v := range r {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	t.Logf("max-assignment=%v, min-assignment=%v max-to-min=%v", max, min, float64(max)/float64(min))

	if nextPrime(104312) != 104323 {
		t.Fatal("nextPrime is broken")
	}
}
