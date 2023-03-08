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

func Filter[T any](data []T, f func(T) bool) []T {
	res := make([]T, 0)
	for _, e := range data {
		if f(e) {
			res = append(res, e)
		}
	}
	return res
}

func Map[T, U any](data []T, f func(T) U) []U {
	res := make([]U, 0, len(data))

	for _, e := range data {
		res = append(res, f(e))
	}

	return res
}
