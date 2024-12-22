package inner

import (
	"fmt"
	"strconv"
)

type Precedence int

const (
	PREC_NONE       Precedence = iota
	PREC_ASSIGNMENT Precedence = iota // =
	PREC_OR         Precedence = iota // or
	PREC_AND        Precedence = iota // and
	PREC_EQUALITY   Precedence = iota // == !=
	PREC_COMPARISON Precedence = iota // < > <= >=
	PREC_TERM       Precedence = iota // + -
	PREC_FACTOR     Precedence = iota // * /
	PREC_UNARY      Precedence = iota // ! -
	PREC_CALL       Precedence = iota // . ()
	PREC_PRIMARY    Precedence = iota
)

type ParseFn func(c *Compiler)

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence Precedence
}

var rules []ParseRule

func init() {
	rules = []ParseRule{
		TOKEN_LEFT_PAREN:    {grouping, nil, PREC_NONE},
		TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		TOKEN_COMMA:         {nil, nil, PREC_NONE},
		TOKEN_DOT:           {nil, nil, PREC_NONE},
		TOKEN_MINUS:         {unary, binary, PREC_TERM},
		TOKEN_PLUS:          {nil, binary, PREC_TERM},
		TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		TOKEN_SLASH:         {nil, binary, PREC_FACTOR},
		TOKEN_STAR:          {nil, binary, PREC_FACTOR},
		TOKEN_BANG:          {nil, nil, PREC_NONE},
		TOKEN_BANG_EQUAL:    {nil, nil, PREC_NONE},
		TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {nil, nil, PREC_NONE},
		TOKEN_GREATER:       {nil, nil, PREC_NONE},
		TOKEN_GREATER_EQUAL: {nil, nil, PREC_NONE},
		TOKEN_LESS:          {nil, nil, PREC_NONE},
		TOKEN_LESS_EQUAL:    {nil, nil, PREC_NONE},
		TOKEN_IDENTIFIER:    {nil, nil, PREC_NONE},
		TOKEN_STRING:        {nil, nil, PREC_NONE},
		TOKEN_NUMBER:        {number, nil, PREC_NONE},
		TOKEN_AND:           {nil, nil, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {nil, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {nil, nil, PREC_NONE},
		TOKEN_OR:            {nil, nil, PREC_NONE},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {nil, nil, PREC_NONE},
		TOKEN_TRUE:          {nil, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func unary(c *Compiler) {
	operatorType := c.parser.previous.Type
	// Compile the operand.
	c.parsePrecedence(PREC_UNARY)

	// Emit the operator instruction.
	switch operatorType {
	case TOKEN_MINUS:
		c.EmitByte(byte(OP_NEGATE))
	default:
		return
	}
}

func binary(c *Compiler) {
	operatorType := c.parser.previous.Type
	rule := c.getRule(operatorType)
	c.parsePrecedence(Precedence(int(rule.precedence) + 1))

	switch operatorType {
	case TOKEN_PLUS:
		c.EmitByte(byte(OP_ADD))
	case TOKEN_MINUS:
		c.EmitByte(byte(OP_SUB))
	case TOKEN_STAR:
		c.EmitByte(byte(OP_MUL))
	case TOKEN_SLASH:
		c.EmitByte(byte(OP_DIV))
	default:
		return
	}
}

func grouping(c *Compiler) {
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}
func number(c *Compiler) {
	t, _ := strconv.ParseFloat(string(c.parser.previous.GetSource()), 64)
	c.emitConstant(Value(t))
}

type Compiler struct {
	parser  *Parser
	chunk   *Chunk
	scanner *Scanner
	debug   bool
}

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

func NewCompiler(debug bool) *Compiler {
	return &Compiler{
		parser: &Parser{},
		chunk:  NewChunk(),
		debug:  debug,
	}
}

func (c *Compiler) Compile(source string) bool {
	c.scanner = NewScanner(source)
	c.Advance()
	c.expression()
	c.consume(TOKEN_EOF, "Expect end of expression.")

	c.endCompiler()

	return !c.parser.hadError
}

func (c *Compiler) Advance() {
	c.parser.previous = c.parser.current

	for {
		c.parser.current = c.scanner.scanToken()
		if c.parser.current.Type != TOKEN_ERROR {
			break
		}

		c.errorAtCurrent("")
	}
}

func (c *Compiler) EmitByte(byte2 byte) {
	c.currentChunk().Write(Mword(byte2), 1)
}

func (c *Compiler) EmitBytes(byte2 byte, byte3 byte) {
	c.currentChunk().Write(Mword(byte2), 1)
	c.currentChunk().Write(Mword(byte3), 1)
}

func (c *Compiler) currentChunk() *Chunk {
	return c.chunk
}

func (c *Compiler) errorAtCurrent(message string) {
	c.ErrorAt(c.parser.current, message)
}

func (c *Compiler) Error(message string) {
	c.ErrorAt(c.parser.previous, message)
}

func (c *Compiler) ErrorAt(token Token, message string) {

	if c.parser.panicMode {
		return
	}
	c.parser.panicMode = true

	fmt.Printf("[line %d] Error", token.Line)

	if token.Type == TOKEN_EOF {
		fmt.Print(" at end")
	} else if token.Type == TOKEN_ERROR {
		// Nothing.
	} else {
		fmt.Printf(" at '%.*s'", len(token.Source), token.Start)
	}

	fmt.Printf(": %s\n", message)
	c.parser.hadError = true

}

func (c *Compiler) consume(ttype TokenType, message string) {
	if c.parser.current.Type == ttype {
		c.Advance()
		return
	}

	c.errorAtCurrent(message)
}

func (c *Compiler) endCompiler() {
	c.emitReturn()
	if !c.parser.hadError {
		if c.debug {
			c.currentChunk().Disassemble("code")
		}
	}
}

func (c *Compiler) emitReturn() {
	c.EmitByte(byte(OP_RETURN))
}

func (c *Compiler) expression() {
	c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) parsePrecedence(precedence Precedence) {
	c.Advance()
	prefixRule := c.getRule(c.parser.previous.Type).prefix
	if prefixRule == nil {
		c.Error("Expect expression.")
		return
	}

	prefixRule(c)

	for precedence <= c.getRule(c.parser.current.Type).precedence {
		c.Advance()
		infixRule := c.getRule(c.parser.previous.Type).infix
		infixRule(c)
	}
}

func (c *Compiler) emitConstant(value Value) {
	c.EmitBytes(byte(OP_CONSTANT), c.makeConstant(value))
}

func (c *Compiler) makeConstant(value Value) byte {
	constant := c.currentChunk().AddConstant(value)
	if constant > 255 {
		c.Error("Too many constants in one chunk.")
		return 0
	}

	return uint8(constant)
}

func (c *Compiler) getRule(operatorType TokenType) ParseRule {
	return rules[operatorType]
}
