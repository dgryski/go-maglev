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

	table := New(names, SmallM)

	r := make(map[string]int, size)
	rand.Seed(0)
	for i := 0; i < 1e6; i++ {
		name := table.Lookup(uint64(rand.Int63()))
		r[name]++
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
