package common

type CountMap struct {
	Map  map[string]int
	Keys []string
}

type CountMapEntry struct {
	Key   string
	Count int
}

func (m *CountMap) Add(key string) {
	count, exists := m.Map[key]
	if exists {
		m.Map[key] = count + 1
	} else {
		m.Map[key] = 1
		m.Keys = append(m.Keys, key)
	}
}

func (m *CountMap) Entries() []CountMapEntry {
	keys := m.Keys
	countMap := m.Map
	entries := make([]CountMapEntry, len(keys))
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		entries[i] = CountMapEntry{Key: key, Count: countMap[key]}
	}
	return entries
}

func NewCountMap() *CountMap {
	return &CountMap{Map: make(map[string]int)}
}
