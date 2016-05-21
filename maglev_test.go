package maglev

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestPopulate(t *testing.T) {

	var tests = []struct {
		dead []int
		want []int
	}{
		{nil, []int{1, 0, 1, 0, 2, 2, 0}},
		{[]int{1}, []int{0, 0, 0, 0, 2, 2, 2}},
	}

	permutations := [][]uint64{
		{3, 0, 4, 1, 5, 2, 6},
		{0, 2, 4, 6, 1, 3, 5},
		{3, 4, 5, 6, 0, 1, 2},
	}

	for _, tt := range tests {
		if got := populate(permutations, tt.dead); !reflect.DeepEqual(got, tt.want) {
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
