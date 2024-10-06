package main

import "fmt"

type mword uint8

type Value float64

const (
	OP_RETURN   = 1
	OP_CONSTANT = 2
)

type ValueArray struct {
	values []*Value
}

func NewValueArray() *ValueArray {
	return &ValueArray{
		values: make([]*Value, 8),
	}
}

func (va *ValueArray) write(v *Value) {
	va.values = append(va.values, v)
}

type Chunk struct {
	code      []uint8
	constants *ValueArray
}

func (c *Chunk) write(b uint8) {
	c.code = append(c.code, b)
}

func NewChunk() *Chunk {
	return &Chunk{
		code:      make([]uint8, 8),
		constants: NewValueArray(),
	}
}

func (c *Chunk) disassemble(name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(c.code); {
		offset = c.disassembleInstruction(offset)
	}
}

func (c *Chunk) disassembleInstruction(offset int) int {
	fmt.Printf("%04d  ", offset)

	instruction := c.code[offset]
	switch instruction {
	case OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)
	case OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", offset)
	default:
		println("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func (c *Chunk) addConstant(v *Value) int {
	c.constants.write(v)
	return len(c.constants.values) - 1
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func constantInstruction(name string) {

}

func main() {
	chunk := NewChunk()
	v := Value(1.2)

	cons := chunk.addConstant(&v)
	chunk.write(OP_CONSTANT)
	chunk.write(OP_RETURN)
	chunk.disassemble("test chunk")

}
