package maglev

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// TestLookup tests the lookup field in table
func TestLookup(t *testing.T) {
	table := getTestingMaglevTable()

	if !reflect.DeepEqual(table.lookup, []int{
		1, 2, 0, 2, 0, 1, 0,
	}) {
		t.Errorf("table lookup field not the same")
	}
}

func TestPopulate(t *testing.T) {
	table := getTestingMaglevTable()

	var tests = []struct {
		dead               []int
		wantEntry          []int
		wantCurrentOffsets []uint64
	}{
		{dead: nil, wantEntry: []int{1, 2, 0, 2, 0, 1, 0}, wantCurrentOffsets: []uint64{1, 3, 5}},
		{dead: []int{1}, wantEntry: []int{0, 2, 0, 2, 0, 2, 0}, wantCurrentOffsets: []uint64{1, 0, 0}},
	}

	permutations := [][]uint64{
		{2, 6, 3, 0, 4, 1, 5},
		{0, 5, 3, 1, 6, 4, 2},
		{1, 3, 5, 0, 2, 4, 6},
	}
	newPermutations := [][]uint64{
		make([]uint64, 7),
		make([]uint64, 7),
		make([]uint64, 7),
	}
	table.resetOffsets()
	for i := 0; i < 3; i++ {
		for j := 0; j < 7; j++ {
			newPermutations[i][j] = table.nextOffset(i)
		}
	}

	if !reflect.DeepEqual(permutations, newPermutations) {
		t.Errorf("permutations=%v, want %v", newPermutations, permutations)
		t.Errorf("1")
	}

	for _, tt := range tests {
		table.resetOffsets()
		if got := table.populate(tt.dead); !reflect.DeepEqual(got, tt.wantEntry) {
			t.Errorf("populate(...,%v)=%v, want %v", tt.dead, got, tt.wantEntry)
		}

		if !reflect.DeepEqual(table.currentOffsets, tt.wantCurrentOffsets) {
			t.Errorf("currentOffsets=%v, want %v", table.currentOffsets, tt.wantCurrentOffsets)
		}
	}
}

func TestDistribution(t *testing.T) {
	const size = 125

	var names []string
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("backend-%d", i))
	}

	table := New(names, SmallM)

	r := make([]int, size)
	rand.Seed(0)
	for i := 0; i < 1e6; i++ {
		idx := table.Lookup(uint64(rand.Int63()))
		r[idx]++
	}

	var total int
	var max = 0
	for _, v := range r {
		total += v
		if v > max {
			max = v
		}
	}

	mean := float64(total) / size
	t.Logf("max=%v, mean=%v, peak-to-mean=%v", max, mean, float64(max)/mean)
}

func getTestingMaglevTable() *Table {
	return New([]string{
		"backend-0",
		"backend-1",
		"backend-2",
	}, 7)
}
