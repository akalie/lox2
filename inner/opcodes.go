package inner

type Mword uint8

const (
	OP_NIL      Mword = iota
	OP_RETURN   Mword = iota
	OP_CONSTANT Mword = iota
	OP_NEGATE   Mword = iota
	OP_ADD      Mword = iota
	OP_SUB      Mword = iota
	OP_MUL      Mword = iota
	OP_DIV      Mword = iota
)
