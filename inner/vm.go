package inner

import "fmt"

type InterpretResult uint8

const (
	INTERPRET_OK            InterpretResult = iota
	INTERPRET_COMPILE_ERROR InterpretResult = iota
	INTERPRET_RUNTIME_ERROR InterpretResult = iota
)
const STACK_MAX = 256

type Vm struct {
	Chunk    *Chunk
	Ip       uint
	Stack    [STACK_MAX]Value
	StackTop uint8
}

func NewVm(chunk *Chunk) *Vm {
	return &Vm{
		Chunk:    chunk,
		Ip:       0,
		Stack:    [256]Value{},
		StackTop: 0,
	}
}

func (vm *Vm) Interpret(source string) InterpretResult {
	Compile(source)

	return INTERPRET_OK
}

func (vm *Vm) readByte() Mword {
	defer func() { vm.Ip++ }()
	return vm.Chunk.code[vm.Ip]
}

func (vm *Vm) readConstant() Value {
	return vm.Chunk.constants.Values[vm.readByte()]
}

func (vm *Vm) Init() {
	vm.ResetStack()
}

func (vm *Vm) Run() InterpretResult {
	for {
		switch instruction := vm.readByte(); instruction {
		case OP_RETURN:
			printValue(vm.Pop())
			return INTERPRET_OK
		case OP_NEGATE:
			vm.Stack[vm.StackTop-1] = -vm.Stack[vm.StackTop-1]
		case OP_CONSTANT:
			constant := vm.readConstant()
			vm.Push(constant)
		case OP_ADD:
			vm.Push(Value(wrapper(float64(vm.Pop()), float64(vm.Pop()), add)))
		case OP_SUB:
			vm.Push(Value(wrapper(float64(vm.Pop()), float64(vm.Pop()), sub)))
		case OP_MUL:
			vm.Push(Value(wrapper(float64(vm.Pop()), float64(vm.Pop()), mul)))
		case OP_DIV:
			vm.Push(Value(wrapper(float64(vm.Pop()), float64(vm.Pop()), div)))
		default:
			return INTERPRET_COMPILE_ERROR
		}
	}
}

func (vm *Vm) ResetStack() {
	vm.StackTop = 0
}

func (vm *Vm) Push(value Value) {
	vm.Stack[vm.StackTop] = value
	vm.StackTop++
}

func (vm *Vm) Pop() Value {
	vm.StackTop--
	return vm.Stack[vm.StackTop]
}

func (vm *Vm) DebugStack() {
	println("          ")
	for pointer := uint8(0); pointer < vm.StackTop; pointer++ {
		println(pointer)
		fmt.Print("[ ")
		printValue(vm.Stack[pointer])
		println(" ]")
	}
	println("")
}

func wrapper(a float64, b float64, op opFunc) float64 {
	return op(a, b)
}

type opFunc func(b, a float64) float64

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