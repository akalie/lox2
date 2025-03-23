package inner

import (
	"bytes"
	"fmt"
	"os"
)

type InterpretResult uint8

const (
	INTERPRET_OK            InterpretResult = iota
	INTERPRET_COMPILE_ERROR InterpretResult = iota
	INTERPRET_RUNTIME_ERROR InterpretResult = iota
)

type ValueType uint8

const STACK_MAX = 256

type Vm struct {
	Chunk    *Chunk
	Ip       uint
	Stack    [STACK_MAX]Value
	StackTop uint8
	debug    bool
	objects  *ObjValue
	globals  *Table
}

func NewVm(chunk *Chunk, debug bool) *Vm {
	return &Vm{
		Chunk:    chunk,
		Ip:       0,
		Stack:    [256]Value{},
		StackTop: 0,
		debug:    debug,
		globals:  NewTable(),
	}
}

func (vm *Vm) Interpret(source string) InterpretResult {
	c := NewCompiler(vm.debug, vm)

	if !c.Compile(source) {
		return INTERPRET_RUNTIME_ERROR
	}

	vm.Chunk = c.chunk
	vm.Ip = 0
	result := vm.Run()

	if vm.debug {
		count := 0
		if vm.objects != nil {
			next := vm.objects
			//println(fmt.Sprintf("%#v", next))
			count = 1
			for next.next != nil {
				next = next.next
				//println(fmt.Sprintf("%#v", next))
				count++
			}
		}
		fmt.Println(fmt.Sprintf("Number of objects: %d", count))
	}

	vm.free()

	return result
}

func (vm *Vm) readByte() byte {
	vm.Ip++
	return vm.Chunk.Code[vm.Ip-1]
}

func (vm *Vm) readConstant() Value {
	return vm.Chunk.constants.Values[vm.readByte()]
}

func (vm *Vm) readString() ObjString {
	t := vm.readConstant().GetObj()
	switch k := t.(type) {
	case ObjString:
		return k
	default:
		panic("We should never be here")
	}
}

func (vm *Vm) Init() {
	vm.ResetStack()
	vm.objects = nil
}

func (vm *Vm) free() {
	freeObjects(vm)
}

func (vm *Vm) Run() InterpretResult {
	var result InterpretResult
	for {
		result = INTERPRET_OK
		switch instruction := vm.readByte(); instruction {
		case OP_NEGATE:
			if !vm.Peek(0).isNumber() {
				vm.runtimeError("Operand must be a number.")
				return INTERPRET_RUNTIME_ERROR
			}
			vm.Push(numberVal(-vm.Pop().GetValue()))
		case OP_CONSTANT:
			constant := vm.readConstant()
			vm.Push(constant)
		case OP_ADD:
			result = vm.binaryOp(numberVal, add)
		case OP_SUB:
			result = vm.binaryOp(numberVal, sub)
		case OP_MUL:
			result = vm.binaryOp(numberVal, mul)
		case OP_DIV:
			result = vm.binaryOp(numberVal, div)
		case OP_NIL:
			vm.Push(nilVal())
		case OP_TRUE:
			vm.Push(boolVal(true))
		case OP_FALSE:
			vm.Push(boolVal(false))
		case OP_POP:
			vm.Pop()
		case OP_GET_LOCAL:
			slot := vm.readByte()
			vm.Push(vm.Stack[slot])
		case OP_SET_LOCAL:
			slot := vm.readByte()
			vm.Stack[slot] = vm.Peek(0)
		case OP_NOT:
			vm.Push(boolVal(vm.isFalsy(vm.Pop())))
		case OP_EQUAL:
			b := vm.Pop()
			a := vm.Pop()
			vm.Push(boolVal(valuesEqual(a, b)))
		case OP_GREATER:
			result = vm.binaryOpBool(boolVal, greater)
		case OP_LESS:
			result = vm.binaryOpBool(boolVal, greater)
		case OP_PRINT:
			t := vm.Pop()
			if t.isNumber() {
				fmt.Printf("#> %v\n", t.GetValue())
			} else {
				fmt.Printf("#> %s\n", toStringObj(t).chars)
			}
		case OP_DEFINE_GLOBAL:
			name := vm.readString()
			vm.globals.Set(&name, vm.Peek(0))
			vm.Pop()
		case OP_GET_GLOBAL:
			name := vm.readString()
			val, ok := vm.globals.Get(&name)
			if !ok {
				vm.runtimeError("Undefined variable '%s'.", name.chars)
				return INTERPRET_RUNTIME_ERROR
			}
			vm.Push(val)
		case OP_SET_GLOBAL:
			name := vm.readString()
			if vm.globals.Set(&name, vm.Peek(0)) {
				vm.globals.Delete(&name)
				vm.runtimeError("Undefined variable '%s'.", name.chars)
				return INTERPRET_RUNTIME_ERROR
			}
		default:
			return INTERPRET_COMPILE_ERROR
		}

		if result != INTERPRET_OK {
			return result
		}
	}
}

func valuesEqual(a Value, b Value) bool {
	if a.ttype != b.ttype {
		return false
	}

	switch a.ttype {
	case VAL_BOOL, VAL_NUMBER:
		return a.GetValue() == b.GetValue()
	case VAL_NIL:
		return true
	case VAL_OBJ:
		switch a.GetObj().GetType() {
		case OBJ_STRING:
			aOb := a.GetObj().(ObjString)
			bOb := b.GetObj().(ObjString)
			if aOb.length != bOb.length {
				return false
			}

			return bytes.Equal(aOb.chars, bOb.chars)
		default:
			return true
		}
	default:
		return false // Unreachable.
	}
}

func (vm *Vm) ResetStack() {
	vm.StackTop = 0
}

func (vm *Vm) Push(value Value) {
	vm.Stack[vm.StackTop] = value
	vm.StackTop++
	if vm.debug {
		vm.DebugStack()
	}
}

func (vm *Vm) Pop() Value {
	vm.StackTop--
	return vm.Stack[vm.StackTop]
}

func (vm *Vm) Peek(distance uint8) Value {
	return vm.Stack[vm.StackTop-distance-1]
}

func (vm *Vm) DebugStack() {
	fmt.Print("[ ")

	for pointer := uint8(0); pointer < vm.StackTop; pointer++ {
		val := vm.Stack[pointer]

		ttypeName := ValTypeMap[val.ttype]
		var value any
		if val.ttype == VAL_NIL {
			value = "nil"
		} else if val.ttype == VAL_OBJ {
			t := val.GetObj()
			ttypeName = ttypeName + " " + t.GetTypeName()
			switch k := t.(type) {
			case ObjString:
				value = string(k.chars)
			default:
				panic("AAAA")
			}
			//value = val.GetObj().getTypeName()
		} else {
			value = val.GetValue()
		}
		l := fmt.Sprintf("%d: (%s) %#v", pointer, ttypeName, value)
		if pointer+1 != vm.StackTop {
			l = l + ", "
		}
		print(l)
	}

	println(" ]")
}

func (vm *Vm) binaryOp(valueType func(v float64) Value, op opFunc) InterpretResult {
	if vm.Peek(0).isObjType(OBJ_STRING) && vm.Peek(1).isObjType(OBJ_STRING) {
		vm.Concatenate()
	} else if vm.Peek(0).isNumber() || vm.Peek(1).isNumber() {
		vm.Push(valueType(op(vm.Pop().GetValue(), vm.Pop().GetValue())))
	} else {
		vm.runtimeError("Operands must be numbers.")
		return INTERPRET_RUNTIME_ERROR
	}

	return INTERPRET_OK
}
func (vm *Vm) binaryOpBool(valueTypeBool func(v bool) Value, opBool opFuncBool) InterpretResult {
	if !vm.Peek(0).isNumber() || !vm.Peek(1).isNumber() {
		vm.runtimeError("Operands must be numbers.")
		return INTERPRET_RUNTIME_ERROR
	}
	vm.Push(valueTypeBool(opBool(vm.Pop().GetValue(), vm.Pop().GetValue())))

	return INTERPRET_OK
}

func (vm *Vm) isFalsy(val Value) bool {
	switch val.ttype {
	case VAL_NIL:
		return true
	case VAL_BOOL:
		return val.GetValue() == 0
	case VAL_NUMBER:
		return val.GetValue() == 0
	default:
		return false
	}
}

func (vm *Vm) runtimeError(format string, vals ...any) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, vals...))
	//
	// instruction = vm.Ip - vm.Chunk.Code[0] - 1;
	instruction := vm.Ip - 1
	line := vm.Chunk.lines[instruction]
	fmt.Fprintln(os.Stderr, fmt.Sprintf("[line %d] in script\n", line))
	vm.ResetStack()
}

func (vm *Vm) Concatenate() {
	b := toStringObj(vm.Pop())
	a := toStringObj(vm.Pop())
	length := a.length + b.length
	t := make([]byte, length)
	s := make([]byte, length)
	s = append(a.chars, b.chars...)
	copy(t, s)

	newStringVal := objVal(ObjString{
		ttype:  OBJ_STRING,
		length: length,
		chars:  t,
	}, vm.objects)

	switch t := newStringVal.v.(type) {
	case *ObjValue:
		vm.objects = t
	}

	vm.Push(newStringVal)
}

type opFunc func(b, a float64) float64
type opFuncBool func(b, a float64) bool

func add(b, a float64) float64 {
	return a + b
}

func sub(b, a float64) float64 {
	return a - b
}

func mul(b, a float64) float64 {
	return a * b
}

func div(b, a float64) float64 {
	return a / b
}

func greater(b, a float64) bool {
	return a > b
}
func less(b, a float64) bool {
	return a < b
}
