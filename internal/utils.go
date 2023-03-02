package internal

func Difference[E comparable](a, b []E) (diff []E) {
	m := make(map[E]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
