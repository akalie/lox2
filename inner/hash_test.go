package inner

import (
	"strconv"
	"testing"
)

func TestSimple(t *testing.T) {
	h := NewTable()
	if h == nil {
		t.Error("Failed to create hash table")
	}

	key := NewObjString([]byte("hello"))
	value := objVal(ObjString{
		ttype:  OBJ_STRING,
		length: 5,
		chars:  []byte("hello"),
	}, nil)
	if !h.Set(&key, value) {
		t.Error("Failed to set key 1")
	}
	if h.count != 1 {
		t.Error("Failed to set key 2")
	}
	if h.capacity != 8 {
		t.Error("Failed init hash properly")
	}
	newVal, ok := h.Get(&key)
	if !ok {
		t.Error("Failed to get value")
	}
	if newVal.GetObj().GetTypeName() != value.GetObj().GetTypeName() {
		t.Error("Failed to get value")
	}
	if toStringObj(newVal).length != toStringObj(value).length {
		t.Error("Failed to get value")
	}

	newNewVal := objVal(ObjString{
		ttype:  OBJ_STRING,
		length: 6,
		chars:  []byte("hello!"),
	}, nil)

	if toStringObj(newNewVal).length == toStringObj(value).length {
		t.Error("Something went wrong")
	}

	h.Set(&key, newNewVal)

	if h.count != 1 {
		t.Error("Counter is wrong")
	}

	retNewNewVal, ok := h.Get(&key)
	if !ok {
		t.Error("Failed to get value")
	}
	if toStringObj(retNewNewVal).length != toStringObj(newNewVal).length {
		t.Error("Failed to get value")
	}

	h.Delete(&key)
	_, ok = h.Get(&key)
	if ok {
		t.Error("Failed to delete key")
	}

	if h.count != 1 {
		t.Error("Counter is wrong, we should have a tombstone")
	}
	h.adjustCapacity(16)
	if h.count != 0 {
		t.Error("Failed to adjust capacity with tombstones correctly")
	}
}

func TestGrow(t *testing.T) {
	h := NewTable()
	for i := 0; i < 100; i++ {
		key := NewObjString([]byte("hello" + strconv.Itoa(i)))
		value := objVal(ObjString{
			ttype:  OBJ_STRING,
			length: 5 + len(strconv.Itoa(i)),
			chars:  []byte("hello" + strconv.Itoa(i)),
		}, nil)
		if !h.Set(&key, value) {
			t.Error("Failed to set key 1")
		}
		if h.count != i+1 {
			t.Error("Failed to set key 2")
		}
	}

	if h.capacity != 256 {
		t.Error("Failed to grow hash properly")
	}
}

func TestOneMilSameKeyEntries(t *testing.T) {
	h := NewTable()
	for i := 0; i < 1000000; i++ {
		key := NewObjString([]byte("hello"))
		value := objVal(ObjString{
			ttype:  OBJ_STRING,
			length: 5,
			chars:  []byte("hello"),
		}, nil)
		h.Set(&key, value)
	}
	if h.count != 1 {
		panic("we should have only one entry")
	}
}

func TestOneMilDiffKeyEntries(t *testing.T) {
	h := NewTable()
	for i := 0; i < 1000000; i++ {
		key := NewObjString([]byte("hello" + strconv.Itoa(i)))
		value := objVal(ObjString{
			ttype:  OBJ_STRING,
			length: 5,
			chars:  []byte("hello"),
		}, nil)
		h.Set(&key, value)
	}
	if h.count != 1000000 {
		panic("we should have only one entry")
	}
}

func TestLookup(t *testing.T) {
	h := NewTable()
	newValue := Value{}
	for i := 0; i < 1000000; i++ {
		key := NewObjString([]byte("hello" + strconv.Itoa(i)))
		value := objVal(ObjString{
			ttype:  OBJ_STRING,
			length: 5 + len(strconv.Itoa(i)),
			chars:  []byte("hello" + strconv.Itoa(i)),
		}, nil)
		if newValue == value {
			t.Error("Values should be different")
		}
		h.Set(&key, value)
		newValue, _ := h.Get(&key)
		if newValue != value {
			t.Error("Failed to get just set value")
		}
	}
	if h.count != 1000000 {
		panic("we should have only one entry")
	}
}
