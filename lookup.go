// +build !amd64

package maglev

func lookup(t *Table, key uint64) string {
	var node int
	if len(t.assignments) == 0 {
		goto notFound
	}
	node = int(t.assignments[key%uint64(len(t.assignments))])
	if node < 0 || node >= len(t.names) {
		goto notFound
	}
	return t.names[node]
notFound:
	return ""
}
