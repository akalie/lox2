package inner

import "fmt"

const (
	VAL_BOOL ValueType = iota
	VAL_NIL
	VAL_NUMBER
	VAL_OBJ
)

var ValTypeMap = map[ValueType]string{
	VAL_NIL:    "nil",
	VAL_NUMBER: "num",
	VAL_BOOL:   "bool",
	VAL_OBJ:    "obj",
}

type iValue interface {
	GetValue() float64
	GetObj() Obj
}

type FloatValue struct {
	v float64
}

func (fv FloatValue) GetValue() float64 {
	return fv.v
}

func (fv FloatValue) GetObj() Obj {
	panic("We should never be here")
}

type BoolValue struct {
	v bool
}

func (bv BoolValue) GetValue() float64 {
	if bv.v {
		return 1
	}
	return 0
}

func (bv BoolValue) GetObj() Obj {
	panic("We should never be here")
}

type ObjValue struct {
	v    Obj
	next *ObjValue
}

func (ov ObjValue) GetObj() Obj {
	return ov.v
}

func (ov ObjValue) GetValue() float64 {
	panic("We should never be here")
}

func (ov ObjValue) GetObjTypeName() string {
	return ov.GetObj().GetTypeName()
}

type Value struct {
	ttype ValueType
	v     iValue
}

func (v Value) GetValue() float64 {
	if v.v == nil {
		return 0 // todo gotcha
	}
	return v.v.GetValue()
}

func (v Value) GetObj() Obj {
	switch t := v.v.(type) {
	case *ObjValue:
		return t.GetObj()
	default:
		panic(fmt.Sprintf("We should never be here %#v", t))
	}
}

func toStringObj(value Value) ObjString {
	if !value.isObjType(OBJ_STRING) {
		panic("We should never be here, toString")
	}
	switch t := value.GetObj().(type) {
	case ObjString:
		return t
	default:
		panic("We should never be here, toString, not objString")
	}
}

func toFuncObj(value Value) ObjFunction {
	if !value.isObjType(OBJ_FUNCTION) {
		panic("We should never be here, toFunction")
	}
	switch t := value.GetObj().(type) {
	case ObjFunction:
		return t
	case *ObjFunction:
		return *t
	default:
		panic("We should never be here, toFuncObj, not function")
	}
}

func toNativeObj(value Value) ObjNative {
	if !value.isObjType(OBJ_NATIVE) {
		panic("We should never be here, toFunction")
	}
	switch t := value.GetObj().(type) {
	case ObjNative:
		return t
	case *ObjNative:
		return *t
	default:
		panic("We should never be here, toFuncObj, not function")
	}
}

func (v Value) isBool() bool {
	return v.ttype == VAL_BOOL
}

func (v Value) isNil() bool {
	return v.ttype == VAL_NIL
}

func (v Value) isNumber() bool {
	return v.ttype == VAL_NUMBER
}

func (v Value) isObj() bool {
	return v.ttype == VAL_OBJ
}

func (v Value) isObjType(ttype ObjType) bool {
	return v.ttype == VAL_OBJ && v.v.GetObj().GetType() == ttype
}

func boolVal(v bool) Value {
	return Value{ttype: VAL_BOOL, v: BoolValue{v: v}}
}

func numberVal(v float64) Value {
	return Value{ttype: VAL_NUMBER, v: FloatValue{v: v}}
}

func nilVal() Value {
	return Value{ttype: VAL_NIL, v: nil}
}

func objVal(v any, next *ObjValue) Value {
	switch t := v.(type) {
	case Obj:
		return Value{ttype: VAL_OBJ, v: &ObjValue{v: t, next: next}}
	case ObjString:
		return Value{ttype: VAL_OBJ, v: &ObjValue{v: ObjString{ttype: OBJ_STRING}, next: next}}
	default:
		panic(fmt.Sprintf("We should never be here(objVal): %#v", t))
	}
}

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
