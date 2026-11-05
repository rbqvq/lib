package slices

func Keys[S ~[]E, E comparable](s S) map[E]struct{} {
	m := make(map[E]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}

func Unique[S ~[]E, E comparable](s S) S {
	m := make(map[E]struct{})
	result := make(S, 0, len(s))
	for _, v := range s {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func Diff[S ~[]E, E comparable](a, b S) S {
	m := Keys(b)

	out := make(S, 0, len(a))
	for _, v := range a {
		if _, ok := m[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}
