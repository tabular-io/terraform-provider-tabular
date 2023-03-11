package internal

func Filter[T any](data []T, f func(T) bool) []T {
	res := make([]T, 0)
	for _, e := range data {
		if f(e) {
			res = append(res, e)
		}
	}
	return res
}

func Difference[E comparable](a, b []E) []E {
	m := make(map[E]bool)

	for _, item := range b {
		m[item] = true
	}

	return Filter(a, func(e E) bool {
		_, ok := m[e]
		return !ok
	})
}

func Intersection[E comparable](a, b []E) (intersection []E) {
	m := make(map[E]bool)

	for _, item := range b {
		m[item] = true
	}

	return Filter(a, func(e E) bool {
		_, ok := m[e]
		return ok
	})
}

func Map[T, U any](data []T, f func(T) U) []U {
	res := make([]U, 0, len(data))

	for _, e := range data {
		res = append(res, f(e))
	}

	return res
}
