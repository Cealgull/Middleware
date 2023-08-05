package utils

func Map[T any, U any](data []T, f func(rune T) U) []U {
	r := make([]U, len(data))
	for i, x := range data {
		r[i] = f(x)
	}
	return r
}

func Filter[T any](data []T, f func(rune T) bool) []T {
	r := make([]T, 0)
	for _, x := range data {
		if f(x) {
			r = append(r, x)
		}
	}
	return r
}

func Reduce[T any](data []T, f func(x T, y T) T) T {
	var r T
	l := len(data)
	if l == 0 {
		return r
	} else {
		r = data[0]
		for i := 1; i < l; i++ {
			r = f(r, data[i])
		}
		return r
	}
}
