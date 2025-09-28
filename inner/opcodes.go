package inner

//type Mword uint8

const (
	OP_NIL           byte = iota
	OP_TRUE          byte = iota
	OP_FALSE         byte = iota
	OP_RETURN        byte = iota
	OP_CONSTANT      byte = iota
	OP_NEGATE        byte = iota
	OP_ADD           byte = iota
	OP_SUB           byte = iota
	OP_MUL           byte = iota
	OP_DIV           byte = iota
	OP_NOT           byte = iota
	OP_EQUAL         byte = iota
	OP_GREATER       byte = iota
	OP_LESS          byte = iota
	OP_PRINT         byte = iota
	OP_POP           byte = iota
	OP_DEFINE_GLOBAL byte = iota
	OP_GET_GLOBAL    byte = iota
	OP_SET_GLOBAL    byte = iota
	OP_GET_LOCAL     byte = iota
	OP_SET_LOCAL     byte = iota
	OP_JUMP_IF_FALSE byte = iota
	OP_JUMP          byte = iota
	OP_LOOP          byte = iota
	OP_CALL          byte = iota
)

var Maaa = map[byte]string{
	OP_NIL:           "OP_NIL",
	OP_TRUE:          "OP_TRUE",
	OP_FALSE:         "OP_FALSE",
	OP_RETURN:        "OP_RETURN",
	OP_CONSTANT:      "OP_CONSTANT",
	OP_NEGATE:        "OP_NEGATE",
	OP_ADD:           "OP_ADD",
	OP_SUB:           "OP_SUB",
	OP_MUL:           "OP_MUL",
	OP_DIV:           "OP_DIV",
	OP_NOT:           "OP_NOT",
	OP_EQUAL:         "OP_EQUAL",
	OP_GREATER:       "OP_GREATER",
	OP_LESS:          "OP_LESS",
	OP_PRINT:         "OP_PRINT",
	OP_POP:           "OP_POP",
	OP_DEFINE_GLOBAL: "OP_DEFINE_GLOBAL",
	OP_GET_GLOBAL:    "OP_GET_GLOBAL",
	OP_SET_GLOBAL:    "OP_SET_GLOBAL",
	OP_GET_LOCAL:     "OP_GET_LOCAL",
	OP_SET_LOCAL:     "OP_SET_LOCAL",
	OP_JUMP_IF_FALSE: "OP_JUMP_IF_FALSE",
	OP_JUMP:          "OP_JUMP",
	OP_LOOP:          "OP_LOOP",
	OP_CALL:          "OP_CALL",
}
