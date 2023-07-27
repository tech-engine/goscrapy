package metadata

type MetaData map[string]any

func (m MetaData) Set(key string, val any) {
	m[key] = val
}

func (m MetaData) Get(key string) any {
	val, ok := m[key]
	if !ok {
		return nil
	}
	return val
}
