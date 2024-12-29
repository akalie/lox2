package inner

//type Mword uint8

const (
	OP_NIL      byte = iota
	OP_TRUE     byte = iota
	OP_FALSE    byte = iota
	OP_RETURN   byte = iota
	OP_CONSTANT byte = iota
	OP_NEGATE   byte = iota
	OP_ADD      byte = iota
	OP_SUB      byte = iota
	OP_MUL      byte = iota
	OP_DIV      byte = iota
	OP_NOT      byte = iota
	OP_EQUAL    byte = iota
	OP_GREATER  byte = iota
	OP_LESS     byte = iota
)
