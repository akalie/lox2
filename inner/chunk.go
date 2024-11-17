package inner

import "fmt"

type Chunk struct {
	code      []Mword
	constants *ValueArray
	lines     []int
}

func (c *Chunk) Write(b Mword, line int) {
	c.code = append(c.code, b)
	c.lines = append(c.lines, line)
}

func NewChunk() *Chunk {
	return &Chunk{
		code:      []Mword{},
		constants: NewValueArray(),
		lines:     []int{},
	}
}

func (c *Chunk) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(c.code); {
		offset = c.disassembleInstruction(offset)
	}
}

func (c *Chunk) disassembleInstruction(offset int) int {
	fmt.Printf("%04d  ", offset)

	if offset > 0 && c.lines[offset] == c.lines[offset-1] {
		fmt.Print("   |")
	} else {
		fmt.Printf("l%d  ", c.lines[offset])
	}

	instruction := c.code[offset]
	switch instruction {
	case OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)
	case OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", c, offset)
	case OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case OP_ADD:
		return simpleInstruction("OP_ADD", offset)
	case OP_SUB:
		return simpleInstruction("OP_SUBTRACT", offset)
	case OP_MUL:
		return simpleInstruction("OP_MULTIPLY", offset)
	case OP_DIV:
		return simpleInstruction("OP_DIVIDE", offset)
	case OP_NIL:
		return offset + 1
	default:
		fmt.Printf("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func (c *Chunk) AddConstant(v Value) Mword {
	c.constants.Write(v)

	return Mword(len(c.constants.Values) - 1)
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func constantInstruction(name string, chunk *Chunk, offset int) int {
	constant := chunk.code[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	printValue(chunk.constants.Values[constant])

	return offset + 2
}

func printValue(value Value) {
	fmt.Printf("%v\n", value)
}
