package maglev

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestPopulate(t *testing.T) {
	table := New([]string{
		"backend-0",
		"backend-1",
		"backend-2",
	}, 7)

	var tests = []struct {
		dead []int
		want []int
	}{
		{nil, []int{1, 2, 0, 2, 0, 1, 0}},
		{[]int{1}, []int{0, 2, 0, 2, 0, 2, 0}},
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
			table.nextOffset(i, &newPermutations[i][j])
		}
	}

	if !reflect.DeepEqual(permutations, newPermutations) {
		t.Errorf("permutations=%v, want %v", newPermutations, permutations)
		t.Errorf("1")
	}

	for _, tt := range tests {
		if got := table.populate(7, tt.dead); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("populate(...,%v)=%v, want %v", tt.dead, got, tt.want)
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
