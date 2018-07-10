package maglev

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestDistribution(t *testing.T) {
	const size = 1000
	const partitions = size * 100

	var names []string
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("backend-%d", i))
	}

	table := New(names, partitions)

	t.Logf("names=%v, partitions=%v, modulus=%v", size, partitions, table.mod)

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
		r[table.names[node]]++
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
		if table.names[originalAssignments[partition]] != table.names[node] {
			reassigned++
		}
	}
	t.Logf("reassigned=%v/%v=%v", reassigned, len(originalAssignments), float64(reassigned)/float64(len(originalAssignments)))

	r = make(map[string]int, size)
	for _, node := range table.assignments {
		if table.names[node] == "backend-13" {
			t.Fatal("Dead node was not reassigned after rebuild")
		}
		r[table.names[node]]++
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

	originalTable := table
	originalTable.Rebuild(nil)

	table = New(originalTable.names, partitions)
	table.Add(fmt.Sprintf("backend-%d", size+1))

	reassigned = 0
	for partition, node := range table.assignments {
		if originalTable.names[originalTable.assignments[partition]] != table.names[node] {
			reassigned++
		}
	}
	t.Logf("reassigned=%v/%v=%v", reassigned, len(originalAssignments), float64(reassigned)/float64(len(originalAssignments)))

}
