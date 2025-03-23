package inner

import "fmt"

type Chunk struct {
	Code      []byte
	constants *ValueArray
	lines     []int
}

func (c *Chunk) Write(b byte, line int) {
	c.Code = append(c.Code, b)
	c.lines = append(c.lines, line)
}

func NewChunk() *Chunk {
	return &Chunk{
		Code:      []byte{},
		constants: NewValueArray(),
		lines:     []int{},
	}
}

func (c *Chunk) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(c.Code); {
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

	instruction := c.Code[offset]
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
	case OP_TRUE:
		return simpleInstruction("OP_TRUE", offset)
	case OP_FALSE:
		return simpleInstruction("OP_FALSE", offset)
	case OP_NOT:
		return simpleInstruction("OP_NOT", offset)
	case OP_NIL:
		return simpleInstruction("OP_FALSE", offset)
	case OP_EQUAL:
		return simpleInstruction("OP_EQUAL", offset)
	case OP_GREATER:
		return simpleInstruction("OP_GREATER", offset)
	case OP_LESS:
		return simpleInstruction("OP_LESS", offset)
	case OP_PRINT:
		return simpleInstruction("OP_PRINT", offset)
	case OP_POP:
		return simpleInstruction("OP_POP", offset)
	case OP_DEFINE_GLOBAL:
		return constantInstruction("OP_DEFINE_GLOBAL", c, offset)
	case OP_GET_GLOBAL:
		return constantInstruction("OP_GET_GLOBAL", c, offset)
	case OP_SET_GLOBAL:
		return constantInstruction("OP_SET_GLOBAL", c, offset)
	case OP_GET_LOCAL:
		return byteInstruction("OP_GET_LOCAL", c, offset)
	case OP_SET_LOCAL:
		return byteInstruction("OP_SET_LOCAL", c, offset)
	default:
		fmt.Printf("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func (c *Chunk) AddConstant(v Value) byte {
	c.constants.Write(v)

	return byte(len(c.constants.Values) - 1)
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func byteInstruction(name string, chunk *Chunk, offset int) int {
	slot := chunk.Code[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}

func constantInstruction(name string, chunk *Chunk, offset int) int {
	constant := chunk.Code[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	printValue(chunk.constants.Values[constant])

	return offset + 2
}

func printValue(value Value) {
	//fmt.Printf("%v\n", value.GetValue())
	switch value.ttype {
	case VAL_BOOL:
		if value.GetValue() == 1 {
			fmt.Print("true")
		} else {
			fmt.Print("false")
		}
	case VAL_NIL:
		print("nil")
	case VAL_NUMBER:
		fmt.Printf("%g", value.GetValue())
	case VAL_OBJ:
		printObj(value)
	}
	println()
}

func printObj(value Value) {
	switch value.GetObj().GetType() {
	case OBJ_STRING:
		switch t := value.GetObj().(type) {
		case ObjString:
			print(string(t.chars))
		default:
			panic("AAAAAAAA")
		}
		break
	}
}
