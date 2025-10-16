package filterMap

func New(data []string, keyFn func(key string) string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range data {
		m[keyFn(v)] = struct{}{}
	}
	return m
}
