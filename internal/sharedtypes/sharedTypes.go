package sharedtypes

import (
	"slices"
)

type Kvlm struct {
	*OrderedMap
}

type OrderedMap struct {
	Map         *map[string][][]byte
	OrderedKeys []string
}

func NewKvlm() *Kvlm {
	return &Kvlm{NewOrderedMap()}
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Map:         &map[string][][]byte{},
		OrderedKeys: []string{},
	}
}

func (orderedMap *OrderedMap) Insert(key string, val [][]byte) {
	if _, ok := (*orderedMap.Map)[key]; ok {
		i := slices.Index(orderedMap.OrderedKeys, key)
		if i >= 0 {
			orderedMap.OrderedKeys = slices.Delete(orderedMap.OrderedKeys, i, i+1)
		}
	}

	orderedMap.OrderedKeys = append(orderedMap.OrderedKeys, key)
	(*orderedMap.Map)[key] = val
}

func (orderedMap *OrderedMap) InsertAndSort(key string, val [][]byte) {
	orderedMap.Insert(key, val)
	orderedMap.Sort()
}

func (orderedMap *OrderedMap) Remove(key string) {
	if _, ok := (*orderedMap.Map)[key]; !ok {
		return
	}

	delete((*orderedMap.Map), key)
	i := slices.Index(orderedMap.OrderedKeys, key)
	if i >= 0 {
		orderedMap.OrderedKeys = slices.Delete(orderedMap.OrderedKeys, i, i+1)
	}
}

func (orderedMap *OrderedMap) Sort() {
	slices.Sort(orderedMap.OrderedKeys)
}
