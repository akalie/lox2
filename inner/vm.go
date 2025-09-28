package inner

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

type InterpretResult uint8

const (
	INTERPRET_OK            InterpretResult = iota
	INTERPRET_COMPILE_ERROR InterpretResult = iota
	INTERPRET_RUNTIME_ERROR InterpretResult = iota
)

type ValueType uint8

const STACK_MAX = 256
const FRAME_MAX = 64

type CallFrame struct {
	function *ObjFunction
	ip       uint16
	slots    uint8
}

func NewCallFrame() *CallFrame {
	return &CallFrame{
		function: nil,
		ip:       0,
		slots:    0,
	}
}

type Vm struct {
	Stack           [STACK_MAX]Value
	StackTop        uint8
	debug           bool
	objects         *ObjValue
	globals         *Table
	CallFrames      [FRAME_MAX]*CallFrame
	FrameCount      uint8
	currentCompiler *Compiler
}

func NewVm(chunk *Chunk, debug bool) *Vm {
	t := [FRAME_MAX]*CallFrame{}
	for i := 0; i < FRAME_MAX; i++ {
		t[i] = NewCallFrame()
	}

	vm := &Vm{
		Stack:      [STACK_MAX]Value{},
		StackTop:   0,
		debug:      debug,
		globals:    NewTable(),
		CallFrames: t,
		FrameCount: 0,
	}

	vm.DefineNative("clock", clockNative, 0)
	vm.DefineNative("add", addNative, 3)
	vm.DefineNative("addVariadic", addNative, -1)

	return vm
}

func (vm *Vm) Interpret(source string) InterpretResult {
	vm.currentCompiler = NewCompiler(vm.debug, vm, FUNK_TYPE_SCRIPT, nil, nil)

	f := vm.currentCompiler.Compile(source)
	if f == nil {
		return INTERPRET_RUNTIME_ERROR
	}

	vm.Push(objVal(f, vm.objects))
	frame := vm.CallFrames[vm.FrameCount]
	if frame == nil {
		vm.CallFrames[vm.FrameCount] = NewCallFrame()
		frame = vm.CallFrames[vm.FrameCount]
	}
	vm.FrameCount++
	frame.function = f
	frame.ip = 0
	frame.slots = vm.StackTop

	result := vm.Run()

	//if vm.debug {
	//	count := 0
	//	if vm.objects != nil {
	//		next := vm.objects
	//		//println(fmt.Sprintf("%#v", next))
	//		count = 1
	//		for next.next != nil {
	//			next = next.next
	//			//println(fmt.Sprintf("%#v", next))
	//			count++
	//		}
	//	}
	//	fmt.Println(fmt.Sprintf("Number of objects: %d", count))
	//}

	vm.free()

	return result
}

func (vm *Vm) readByte() byte {
	frame := vm.CallFrames[vm.FrameCount-1]
	frame.ip++

	return frame.function.Chunk.Code[frame.ip-1]
}

func (vm *Vm) readConstant() Value {
	frame := vm.CallFrames[vm.FrameCount-1]
	return frame.function.Chunk.constants.Values[vm.readByte()]
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
	frame := vm.CallFrames[vm.FrameCount-1]
	for {
		result = INTERPRET_OK
		instruction := vm.readByte()
		switch instruction {
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
			//vm.Push(vm.Stack[stack.Slots])
			vm.Push(vm.Stack[frame.slots+slot])
		case OP_SET_LOCAL:
			slot := vm.readByte()
			//frame.slots[slot] = vm.Peek(0)
			vm.Stack[frame.slots+slot] = vm.Peek(0)
		case OP_NOT:
			vm.Push(boolVal(vm.isFalsy(vm.Pop())))
		case OP_EQUAL:
			b := vm.Pop()
			a := vm.Pop()
			vm.Push(boolVal(valuesEqual(a, b)))
		case OP_GREATER:
			result = vm.binaryOpBool(boolVal, greater)
		case OP_LESS:
			result = vm.binaryOpBool(boolVal, less)
		case OP_PRINT:
			t := vm.Pop()
			if t.isNumber() {
				fmt.Printf("#> %v\n", t.GetValue())
			} else if t.isObj() {
				fmt.Printf("#> %s\n", toStringObj(t).chars)
			} else {
				fmt.Printf("#> %v <- this is unexpected btw \n", t)
			}
		case OP_DEFINE_GLOBAL:
			name := vm.readString()
			vm.globals.Set(&name, vm.Peek(0))
			vm.Pop()
		case OP_GET_GLOBAL:
			name := vm.readString()

			val, ok := vm.globals.Get(&name)
			if !ok {
				vm.runtimeError("Undefined variable1 '%s'.", name.chars)
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
		case OP_JUMP_IF_FALSE:
			offset := vm.readShort()
			if vm.isFalsy(vm.Peek(0)) {
				frame.ip += offset
			}
		case OP_JUMP:
			offset := vm.readShort()
			frame.ip += offset
		case OP_LOOP:
			offset := vm.readShort()
			frame.ip -= offset
		case OP_CALL:
			argCount := vm.readByte()
			//fmt.Printf("Arg call count %d \n", argCount)
			if !vm.CallValue(vm.Peek(argCount), argCount) {
				return INTERPRET_RUNTIME_ERROR
			}
			frame = vm.CallFrames[vm.FrameCount-1]
		case OP_RETURN:
			res := vm.Pop()
			vm.FrameCount--
			if vm.FrameCount == 0 {
				vm.Pop()
				return INTERPRET_OK
			}
			vm.StackTop = frame.slots
			vm.Push(res)
			frame = vm.CallFrames[vm.FrameCount-1]
		default:
			return INTERPRET_COMPILE_ERROR
		}
		if vm.debug {
			fmt.Printf("instruction %s: \n", Maaa[instruction])
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
	vm.FrameCount = 0
}

func (vm *Vm) Push(value Value) {
	if value.isBool() {
		//fmt.Printf("AAAAAAAAAAAbb")

	}
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
			case *ObjFunction:
				value = string(k.Name.chars)
			case *ObjNative:
				value = k.Name
			default:
				fmt.Printf("%#v\n", k)
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
		vm.runtimeError("Operands must be numbers (vrament).")
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

	//frame := vm.CallFrames[vm.FrameCount-1]
	//instruction := frame.ip - uint16(frame.function.Chunk.Code[0]) - 1
	//line := frame.function.Chunk.lines[instruction]

	fmt.Fprintln(os.Stderr, fmt.Sprintf("[line TODO%d] in script\n", -12))

	for i := int(vm.FrameCount) - 1; i >= 0; i-- {
		frame := vm.CallFrames[i]
		f := frame.function
		//instruction := frame.ip - uint16(frame.function.Chunk.Code[0]) - 1
		//fmt.Println(fmt.Sprintf("[line %d] in ", f.Chunk.lines[instruction]))
		if len(f.Name.chars) == 0 {
			fmt.Println("script")
		} else {
			fmt.Printf("%s()\n", string(f.Name.chars))
		}
	}
	vm.ResetStack()
}

/*
DefineNative defines a native function in the VM's global scope.
*/
func (vm *Vm) DefineNative(name string, fn NativeFn, arity int) {
	vm.Push(objVal(NewObjString([]byte(name)), vm.objects))
	vm.Push(objVal(NewObjNative(fn, name, arity), vm.objects))
	t := toStringObj(vm.Peek(1))
	vm.globals.Set(&t, vm.Peek(0))
	vm.Pop()
	vm.Pop()
}

func clockNative(argCount byte, args ...Value) Value {
	return numberVal(float64(time.Now().Unix()))
}

func addNative(argCount byte, args ...Value) Value {
	sum := 0.0
	for _, arg := range args {
		if !arg.isNumber() {
			return numberVal(0)
		}
		sum += arg.GetValue()
	}
	return numberVal(sum)
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

func (vm *Vm) readShort() uint16 {
	frame := vm.CallFrames[vm.FrameCount-1]
	frame.ip += 2
	tt := (uint16(frame.function.Chunk.Code[frame.ip-2]) << 8) | uint16(frame.function.Chunk.Code[frame.ip-1])

	return tt
}

func (vm *Vm) CallValue(callee Value, argCount byte) bool {
	if callee.isObj() {
		switch true {
		case callee.isObjType(OBJ_FUNCTION):
			t := toFuncObj(callee)
			return vm.Call(&t, argCount)
		case callee.isObjType(OBJ_NATIVE):
			native := toNativeObj(callee)
			args := make([]Value, argCount)
			if native.Arity != -1 && int(argCount) != native.Arity {
				vm.runtimeError(
					"Expected %d arguments but got %d.",
					native.Arity,
					argCount,
				)
				return false
			}
			for i := byte(0); i < argCount; i++ {
				args[i] = vm.Stack[vm.StackTop-argCount+i]
			}
			result := native.NFunk(argCount, args...)
			vm.StackTop -= argCount + 1
			vm.Push(result)

			return true
		default:
			break // Non-callable object type.
		}

	}
	vm.runtimeError("Can only call functions and classes.")
	return false
}

func (vm *Vm) Call(function *ObjFunction, argCount byte) bool {
	if int(argCount) != function.Arity {
		vm.runtimeError(
			"Expected %d arguments but got %d.",
			function.Arity,
			argCount,
		)
		return false
	}

	frame := vm.CallFrames[vm.FrameCount]
	vm.FrameCount += 1
	if vm.FrameCount == FRAME_MAX {
		vm.runtimeError("Stack overflow.")
		return false
	}
	frame.function = function
	//frame.ip = uint16(int(vm.StackTop) - int(argCount) - 1)
	frame.ip = 0
	frame.slots = vm.StackTop - argCount - 1
	//todo ?
	return true
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
