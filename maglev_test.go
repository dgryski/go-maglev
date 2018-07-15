package maglev

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestDistribution(t *testing.T) {
	const size = 1000
	const partitions = size * 100

	rand.Seed(0)

	var names []string
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("backend-%d", i))
	}

	table := New(names, partitions)

	for i := 0; i < 1e8*6; i++ {
		if len(table.Lookup(uint64(rand.Int63()))) == 0 {
			t.Fatal("Failed lookup")
		}
	}

	t.Logf("[New]: names=%v, partitions=%v, modulus=%v", size, partitions, table.mod)

	var min, max int
	getMinMax := func(r map[string]int) {
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
	}

	r := make(map[string]int, size)
	for _, node := range table.assignments {
		r[table.names[node]]++
	}
	getMinMax(r)
	t.Logf("[New]: max-assignment=%v, min-assignment=%v max-to-min=%v", max, min, float64(max)/float64(min))

	originalTable := table
	originalTable.Rebuild(nil)

	table = New(originalTable.names, partitions)
	table.Remove("backend-13")

	reassigned := 0
	for partition, node := range table.assignments {
		if originalTable.names[originalTable.assignments[partition]] != table.names[node] {
			reassigned++
		}
	}
	t.Logf("[New -> Remove 1]: reassigned=%v/%v=%v", reassigned, partitions, float64(reassigned)/float64(partitions))

	r = make(map[string]int, size)
	for _, node := range table.assignments {
		if table.names[node] == "backend-13" {
			t.Fatal("Dead node was not reassigned after rebuild")
		}
		r[table.names[node]]++
	}
	getMinMax(r)
	t.Logf("[New -> Remove 1]: max-assignment=%v, min-assignment=%v max-to-min=%v", max, min, float64(max)/float64(min))

	table = New(originalTable.names, partitions)
	table.Add(append(originalTable.names, fmt.Sprintf("backend-%d", size))...)

	reassigned = 0
	for partition, node := range table.assignments {
		if originalTable.names[originalTable.assignments[partition]] != table.names[node] {
			reassigned++
		}
	}
	t.Logf("[New -> Add 1]: reassigned=%v/%v=%v", reassigned, partitions, float64(reassigned)/float64(partitions))

	r = make(map[string]int, size)
	for _, node := range table.assignments {
		r[table.names[node]]++
	}
	getMinMax(r)
	t.Logf("[New -> Add 1]: max-assignment=%v, min-assignment=%v max-to-min=%v", max, min, float64(max)/float64(min))

}
