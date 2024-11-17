package inner

type Value float64

type ValueArray struct {
	Values []Value
}

func NewValueArray() *ValueArray {
	return &ValueArray{
		Values: []Value{},
	}
}

func (va *ValueArray) Write(v Value) {
	va.Values = append(va.Values, v)
}
