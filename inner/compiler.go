package inner

import (
	"bytes"
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

type ParseFn func(c *Compiler, canAssign bool)

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
		TOKEN_BANG:          {unary, nil, PREC_NONE},
		TOKEN_BANG_EQUAL:    {nil, binary, PREC_EQUALITY},
		TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {nil, binary, PREC_EQUALITY},
		TOKEN_GREATER:       {nil, binary, PREC_COMPARISON},
		TOKEN_GREATER_EQUAL: {nil, binary, PREC_COMPARISON},
		TOKEN_LESS:          {nil, binary, PREC_COMPARISON},
		TOKEN_LESS_EQUAL:    {nil, binary, PREC_COMPARISON},
		TOKEN_IDENTIFIER:    {variable, nil, PREC_NONE},
		TOKEN_STRING:        {str, nil, PREC_NONE},
		TOKEN_NUMBER:        {number, nil, PREC_NONE},
		TOKEN_AND:           {nil, nil, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, nil, PREC_NONE},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {nil, nil, PREC_NONE},
		TOKEN_TRUE:          {literal, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func unary(c *Compiler, canAssign bool) {
	operatorType := c.parser.previous.Type
	// Compile the operand.
	c.parsePrecedence(PREC_UNARY)

	// Emit the operator instruction.
	switch operatorType {
	case TOKEN_MINUS:
		c.EmitByte(OP_NEGATE)
	case TOKEN_BANG:
		c.EmitByte(OP_NOT)
	default:
		return
	}
}

func binary(c *Compiler, canAssign bool) {
	operatorType := c.parser.previous.Type
	rule := c.getRule(operatorType)
	c.parsePrecedence(Precedence(int(rule.precedence) + 1))

	switch operatorType {
	case TOKEN_PLUS:
		c.EmitByte(OP_ADD)
	case TOKEN_MINUS:
		c.EmitByte(OP_SUB)
	case TOKEN_STAR:
		c.EmitByte(OP_MUL)
	case TOKEN_SLASH:
		c.EmitByte(OP_DIV)
	case TOKEN_BANG_EQUAL:
		c.EmitBytes(OP_EQUAL, OP_NOT)
	case TOKEN_EQUAL_EQUAL:
		c.EmitByte(OP_EQUAL)
	case TOKEN_GREATER:
		c.EmitByte(OP_GREATER)
	case TOKEN_GREATER_EQUAL:
		c.EmitBytes(OP_LESS, OP_NOT)
	case TOKEN_LESS:
		c.EmitByte(OP_LESS)
	case TOKEN_LESS_EQUAL:
		c.EmitBytes(OP_GREATER, OP_NOT)
	default:
		return
	}
}

func literal(c *Compiler, canAssign bool) {
	switch c.parser.previous.Type {
	case TOKEN_FALSE:
		c.EmitByte(OP_FALSE)
	case TOKEN_NIL:
		c.EmitByte(OP_NIL)
	case TOKEN_TRUE:
		c.EmitByte(OP_TRUE)
	default:
		return // Unreachable.
	}
}

func grouping(c *Compiler, canAssign bool) {
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}
func number(c *Compiler, canAssign bool) {
	t, _ := strconv.ParseFloat(string(c.parser.previous.GetSource()), 64)
	c.emitConstant(numberVal(t))
}

func str(c *Compiler, canAssign bool) {
	s := make([]byte, len(c.parser.previous.Source)-2)
	ddt := c.parser.previous.Source[c.parser.previous.Start+1 : len(c.parser.previous.Source)-1]
	copy(s, ddt)
	sVal := NewObjString(s)
	wrap := objVal(sVal, c.vm.objects)
	switch t := wrap.v.(type) {
	case *ObjValue:
		c.vm.objects = t
	}
	c.emitConstant(wrap)
}

func variable(c *Compiler, canAssign bool) {
	c.namedVariable(c.parser.previous, canAssign)
}

type Local struct {
	name  Token
	depth int
}

func NewLocal() *Local {
	return &Local{
		depth: -1,
	}
}

type Compiler struct {
	parser     *Parser
	chunk      *Chunk
	scanner    *Scanner
	debug      bool
	vm         *Vm
	locals     [256]*Local
	localCount int
	scopeDepth int
}

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

func NewCompiler(debug bool, vm *Vm) *Compiler {
	locals := [256]*Local{}
	for i := 0; i < len(locals); i++ {
		locals[i] = NewLocal()
	}
	return &Compiler{
		parser: &Parser{},
		chunk:  NewChunk(),
		debug:  debug,
		vm:     vm,
		locals: locals,
	}
}

func (c *Compiler) Compile(source string) bool {
	c.scanner = NewScanner(source)
	c.Advance()

	for !c.Match(TOKEN_EOF) {
		c.declaration()
	}
	c.endCompiler()

	return !c.parser.hadError
}

func (c *Compiler) declaration() {
	if c.Match(TOKEN_VAR) {
		c.VarDeclaration()
	} else {
		c.Statement()
	}
	if c.parser.panicMode {
		c.synchronize()
	}
}

func (c *Compiler) Statement() {
	if c.Match(TOKEN_PRINT) {
		c.PrintStatement()
	} else if c.Match(TOKEN_LEFT_BRACE) {
		c.BeginScope()
		c.Block()
		c.EndScope()
	} else {
		c.ExpressionStatement()
	}
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
	c.currentChunk().Write(byte2, 1)
}

func (c *Compiler) EmitBytes(byte2 byte, byte3 byte) {
	c.currentChunk().Write(byte2, 1)
	c.currentChunk().Write(byte3, 1)
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
		fmt.Printf(" at '%s'", token.Source)
		fmt.Printf(" at line %d symbol %d", token.Line, token.Start)
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
	c.EmitByte(OP_RETURN)
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

	canAssign := precedence <= PREC_ASSIGNMENT
	prefixRule(c, canAssign)

	for precedence <= c.getRule(c.parser.current.Type).precedence {
		c.Advance()
		infixRule := c.getRule(c.parser.previous.Type).infix
		infixRule(c, canAssign)
	}

	if canAssign && c.Match(TOKEN_EQUAL) {
		c.Error("Invalid assignment target.")
	}
}

func (c *Compiler) emitConstant(value Value) {
	c.EmitBytes(OP_CONSTANT, c.makeConstant(value))
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

func (c *Compiler) Match(ttype TokenType) bool {
	if !c.Check(ttype) {
		return false
	}
	c.Advance()
	return true
}

func (c *Compiler) Check(ttype TokenType) bool {
	return c.parser.current.Type == ttype
}

func (c *Compiler) PrintStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect '' after value.")
	c.EmitByte(OP_PRINT)
}

func (c *Compiler) ExpressionStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect '' after expression.")
	c.EmitByte(OP_POP)
}

func (c *Compiler) synchronize() {
	c.parser.panicMode = false

	for c.parser.current.Type != TOKEN_EOF {
		if c.parser.previous.Type == TOKEN_SEMICOLON {
			return
		}
		switch c.parser.current.Type {
		case TOKEN_CLASS, TOKEN_FUN, TOKEN_VAR, TOKEN_FOR, TOKEN_IF, TOKEN_WHILE, TOKEN_PRINT, TOKEN_RETURN:
			return
		default:
			// Do nothing.
		}

		c.Advance()
	}
}

func (c *Compiler) VarDeclaration() {
	global := c.parseVariable("Expect variable name.")

	if c.Match(TOKEN_EQUAL) {
		c.expression()
	} else {
		c.EmitByte(OP_NIL)
	}
	c.consume(TOKEN_SEMICOLON, "Expect '' after variable declaration.")

	c.DefineVariable(global)
}

func (c *Compiler) DefineVariable(global byte) {
	if c.scopeDepth > 0 {
		c.markInitialized()
		return
	}

	c.EmitBytes(OP_DEFINE_GLOBAL, global)
}

func (c *Compiler) parseVariable(errorMessage string) byte {
	c.consume(TOKEN_IDENTIFIER, errorMessage)

	c.declareVariable()
	if c.scopeDepth > 0 {
		return 0
	}

	return c.identifierConstant(c.parser.previous)
}

func (c *Compiler) identifierConstant(name Token) byte {
	return c.makeConstant(objVal(NewObjString(name.Source), c.vm.objects))
}

func (c *Compiler) namedVariable(name Token, canAssign bool) {
	argt := c.resolveLocal(name)
	var getOp, setOp, arg byte
	if argt != -1 {
		arg = uint8(argt)
		getOp = OP_GET_LOCAL
		setOp = OP_SET_LOCAL
	} else {
		arg = c.identifierConstant(name)
		getOp = OP_GET_GLOBAL
		setOp = OP_SET_GLOBAL
	}

	if canAssign && c.Match(TOKEN_EQUAL) {
		c.expression()
		c.EmitBytes(setOp, arg)
	} else {
		c.EmitBytes(getOp, arg)
	}
}

func (c *Compiler) Block() {
	for !c.Check(TOKEN_RIGHT_BRACE) && !c.Check(TOKEN_EOF) {
		c.declaration()
	}

	c.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (c *Compiler) BeginScope() {
	c.scopeDepth++
}

func (c *Compiler) EndScope() {
	c.scopeDepth--
	for c.localCount > 0 && c.locals[c.localCount-1].depth > c.scopeDepth {
		c.EmitByte(OP_POP)
		c.localCount--
	}
}

func (c *Compiler) declareVariable() {
	if c.scopeDepth == 0 {
		return
	}

	name := c.parser.previous

	for i := c.localCount - 1; i >= 0; i-- {
		local := c.locals[i]
		if local.depth != -1 && local.depth < c.scopeDepth {
			break
		}

		if c.identifiersEqual(name, local.name) {
			c.errorAtCurrent("Already a variable with this name in this scope.")
		}
	}

	c.addLocal(name)
}

func (c *Compiler) addLocal(name Token) {
	if c.localCount == STACK_MAX {
		c.errorAtCurrent("Too many local variables in function.")

		return
	}

	local := c.locals[c.localCount]
	c.localCount++
	local.name = name
	local.depth = c.scopeDepth
}

func (c *Compiler) identifiersEqual(a Token, b Token) bool {
	if len(a.Source) != len(b.Source) {
		return false
	}
	return bytes.Equal(a.Source, b.Source)
}

func (c *Compiler) resolveLocal(name Token) int {
	for i := c.localCount - 1; i >= 0; i-- {
		local := c.locals[i]
		if c.identifiersEqual(name, local.name) {
			if local.depth == -1 {
				c.errorAtCurrent("Can't read local variable in its own initializer.")
			}
			return i
		}
	}

	return -1
}

func (c *Compiler) markInitialized() {
	c.locals[c.localCount-1].depth = c.scopeDepth
}
