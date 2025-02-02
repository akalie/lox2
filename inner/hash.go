package inner

import (
	"bytes"
	"fmt"
)

const TABLE_MAX_LOAD = 0.75

type Table struct {
	count    int
	capacity uint32
	entries  []*Entry
}

type Entry struct {
	key   *ObjString
	value Value
}

func NewTable() *Table {
	return &Table{
		count:    0,
		capacity: 0,
		entries:  nil,
	}
}

func (t *Table) Set(key *ObjString, value Value) bool {
	if float64(t.count+1) > float64(t.capacity)*TABLE_MAX_LOAD {
		capacity := grow(t.capacity)
		fmt.Sprintf("Growing table to %d", capacity)
		t.adjustCapacity(capacity)
	}

	entry := findEntry(t.entries, t.capacity, key)
	isNewKey := entry.key == nil
	if isNewKey && entry.value.v == nil {
		t.count++
	}
	entry.key = key
	entry.value = value

	return isNewKey
}

func (t *Table) Delete(key *ObjString) bool {
	if t.count == 0 {
		return false
	}

	entry := findEntry(t.entries, t.capacity, key)
	if entry.key == nil {
		return false
	}

	// toombstone
	entry.key = nil
	entry.value = boolVal(true)

	return true
}

func (t *Table) Get(key *ObjString) (Value, bool) {
	if t.capacity == 0 {
		return Value{}, false
	}

	entry := findEntry(t.entries, t.capacity, key)
	return entry.value, entry.key != nil
}

func (t *Table) adjustCapacity(newCap uint32) {
	newSlice := make([]*Entry, newCap)
	t.count = 0
	for _, entry := range t.entries {
		if entry == nil || entry.key == nil {
			continue
		}
		dest := findEntry(newSlice, newCap, entry.key)
		dest.key = entry.key
		dest.value = entry.value
		t.count++
	}
	t.entries = newSlice
	t.capacity = newCap
}

func findEntry(entries []*Entry, capacity uint32, key *ObjString) *Entry {
	index := key.hash % capacity
	var tombstone *Entry

	for {
		// check if entries[index] exists
		if entries[index] == nil {
			newEntry := &Entry{key: nil, value: Value{}}
			entries[index] = newEntry
			return newEntry
		}
		entry := entries[index]

		if entry.key == nil {
			if entry.value.v == nil {
				if tombstone == nil {
					return tombstone
				}
				return entry
			}
			if tombstone == nil {
				tombstone = entry
			}
		} else if bytes.Equal(entry.key.chars, key.chars) {
			return entry
		}

		index = (index + 1) % capacity
	}
}

func grow(capacity uint32) uint32 {
	if capacity < 8 {
		return 8
	}
	return capacity * 2
}

func copyTable(from *Table, to *Table) {
	for _, entry := range from.entries {
		if entry.key != nil {
			to.Set(entry.key, entry.value)
		}
	}
}
