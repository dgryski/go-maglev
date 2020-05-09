package maglev

import (
	"fmt"
	"testing"
)

var total int

func BenchmarkGenerate(b *testing.B) {
	const size = 125

	var names []string
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("backend-%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		offsets, _ := generateOffsetAndSkips(names, SmallM)
		total += len(offsets)
	}
}

func BenchmarkNew(b *testing.B) {
	const size = 125

	var names []string
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("backend-%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		table := New(names, SmallM)
		total += len(table.offsets)
	}
}
