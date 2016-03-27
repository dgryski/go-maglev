package maglev

func populate(permutation [][]int) []int {

	M := len(permutation)
	N := len(permutation[0])

	// N is number of backends
	// M is the size of the lookup table (prime)

	// 2: for each i < N do next[i] ← 0 end for
	next := make([]int, N)

	// 3: for each j < M do entry[ j] ← −1 end for
	entry := make([]int, M)
	for j := range entry {
		entry[j] = -1
	}

	// 4: n ← 0
	n := 0

	// 5: while true do
	for {
		// 6: for each i < N do
		for i := 0; i < N; i++ {
			// 7: c ← permutation[i][next[i]]
			c := permutation[i][next[i]]

			// 8: while entry[c] ≥ 0 do
			for entry[c] >= 0 {
				// 9: next[i] ← next[i] +1
				next[i]++
				// 10: c ← permutation[i][next[i]]
				c = permutation[i][next[i]]
				// 11: end while
			}

			// 12: entry[c] ← i
			entry[c] = i
			// 13: next[i] ← next[i] +1
			next[i]++
			// 14: n ← n+1
			n++
			//15: if n = M then return end if
			if n == M {
				return entry
			}
			//16: end for
		}
		// 17: end while
	}
	//18: end function
}
