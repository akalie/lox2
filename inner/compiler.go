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

type FuncType int

const (
	FUNK_TYPE_FUNCTION FuncType = iota
	FUNK_TYPE_SCRIPT   FuncType = iota
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
		TOKEN_LEFT_PAREN:    {grouping, call, PREC_CALL},
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
		TOKEN_AND:           {nil, and, PREC_AND},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, or, PREC_OR},
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

func and(c *Compiler, assign bool) {
	endJump := c.emitJump(OP_JUMP_IF_FALSE)

	c.EmitByte(OP_POP)
	c.parsePrecedence(PREC_AND)

	c.PatchJump(endJump)
}

func or(c *Compiler, assign bool) {
	elseJump := c.emitJump(OP_JUMP_IF_FALSE)
	endJump := c.emitJump(OP_JUMP)

	c.PatchJump(elseJump)
	c.EmitByte(OP_POP)

	c.parsePrecedence(PREC_OR)
	c.PatchJump(endJump)
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

func call(c *Compiler, canAssign bool) {
	argCount := argumentList(c)
	if argCount > 255 {
		c.Error("Can't have more than 255 arguments.")
	}
	c.EmitBytes(OP_CALL, byte(argCount))
}

func argumentList(c *Compiler) int {
	argCount := 0
	if !c.Check(TOKEN_RIGHT_PAREN) {
		for {
			c.expression()
			argCount = argCount + 1
			if !c.Match(TOKEN_COMMA) {
				break
			}
		}
	}
	if !c.Check(TOKEN_RIGHT_PAREN) {
		c.errorAtCurrent("Expect `)` after expression")

	}
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after arguments.")

	return argCount
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
	function   *ObjFunction
	funcType   FuncType
	parser     *Parser
	chunk      *Chunk
	scanner    *Scanner
	debug      bool
	vm         *Vm
	locals     [256]*Local
	localCount int
	scopeDepth int
	Enclosing  *Compiler
}

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

func NewCompiler(debug bool, vm *Vm, funcType FuncType, scanner *Scanner, c *Compiler) *Compiler {
	locals := [256]*Local{}
	for i := 0; i < len(locals); i++ {
		locals[i] = NewLocal()
	}
	local := locals[0]
	local.depth = 0
	local.name.Start = 0
	local.name.Source = []byte("")
	parser := &Parser{}
	if c != nil {
		parser = c.parser
	}
	name := []byte{}
	if funcType != FUNK_TYPE_SCRIPT {
		name = c.parser.previous.Source
	}
	return &Compiler{
		parser:     parser,
		chunk:      NewChunk(),
		debug:      debug,
		vm:         vm,
		locals:     locals,
		funcType:   funcType,
		function:   NewObjFunction(0, NewObjString(name)),
		localCount: 1,
		scanner:    scanner,
		Enclosing:  c,
	}
}

//func AddLocalCompiler(c *Compiler, funcType FuncType, vm *Vm) *Compiler {
//	c.Enclosing = vm.currentCompiler
//	c.function = nil
//	c.funcType = funcType
//	c.localCount = 0
//	c.scopeDepth = 0
//
//	c.function = NewObjFunction(0, NewObjString([]byte("")))
//	vm.currentCompiler = c

//
//	local := vm.currentCompiler.locals[vm.currentCompiler.localCount]
//	vm.currentCompiler.localCount++
//	local.depth = 0
//	//local.isCaptured = false
//	local.name.Start = 0
//	local.name.Source = []byte("")
//	if funcType != FUNK_TYPE_FUNCTION {
//		//	local.name.Source = "this"
//		//	local.name.length = 4
//		//} else {
//
//	}
//	return c
//}

func (c *Compiler) Compile(source string) *ObjFunction {
	c.scanner = NewScanner(source)
	c.Advance()

	for !c.Match(TOKEN_EOF) {
		c.declaration()
	}
	f := c.endCompiler()

	if c.parser.hadError {
		return nil
	} else {
		return f
	}
}

func (c *Compiler) declaration() {
	if c.Match(TOKEN_FUN) {
		c.FunDeclaration()
	} else if c.Match(TOKEN_VAR) {
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
	} else if c.Match(TOKEN_IF) {
		c.IfStatement()
	} else if c.Match(TOKEN_WHILE) {
		c.whileStatement()
	} else if c.Match(TOKEN_RETURN) {
		c.returnStatement()
	} else if c.Match(TOKEN_FOR) {
		c.forStatement()
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

		c.errorAtCurrent("Error!")
	}
}

func (c *Compiler) EmitByte(byte2 byte) {
	if byte2 == OP_NIL {
		fmt.Printf("AAAAAAAAa")
	}
	c.currentChunk().Write(byte2, 1)
}

func (c *Compiler) EmitBytes(byte2 byte, byte3 byte) {
	c.currentChunk().Write(byte2, 1)
	c.currentChunk().Write(byte3, 1)
}

func (c *Compiler) currentChunk() *Chunk {
	return c.function.Chunk
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
		fmt.Printf(" at line %d symbol %d", token.Line, token.CurrentChar)
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

func (c *Compiler) endCompiler() *ObjFunction {
	c.emitReturn()
	if !c.parser.hadError {
		if c.debug {
			if len(c.function.Name.chars) == 0 {
				c.currentChunk().Disassemble("<script>")
			} else {
				c.currentChunk().Disassemble(string(c.function.Name.chars))
			}
		}
	}
	c.vm.currentCompiler = c.Enclosing

	return c.function
}

func (c *Compiler) emitReturn() {
	c.EmitByte(OP_NIL)
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
	c.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	c.EmitByte(OP_PRINT)
}

func (c *Compiler) ExpressionStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
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
	c.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration.")

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
	if c.scopeDepth == 0 {
		return
	}
	c.locals[c.localCount-1].depth = c.scopeDepth
}

func (c *Compiler) IfStatement() {
	c.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	thenJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.EmitByte(OP_POP)
	c.Statement()
	elseJump := c.emitJump(OP_JUMP)

	c.PatchJump(thenJump)
	if c.Match(TOKEN_ELSE) {
		c.Statement()
	}
	c.PatchJump(elseJump)
}

func (c *Compiler) emitJump(instruction byte) int {
	c.EmitByte(instruction)
	c.EmitByte(0xff)
	c.EmitByte(0xff)
	return len(c.currentChunk().Code) - 2
}

func (c *Compiler) PatchJump(offset int) {
	jump := len(c.currentChunk().Code) - offset - 2

	if jump > 255 {
		c.Error("Too much code to jump over.")
	}

	c.currentChunk().Code[offset] = byte((jump >> 8) & 0xff)
	c.currentChunk().Code[offset+1] = byte(jump & 0xff)
}

func (c *Compiler) whileStatement() {
	loopStart := len(c.currentChunk().Code)
	c.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'while'.")
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.EmitByte(OP_POP)
	c.Statement()
	c.EmitLoop(loopStart)

	c.PatchJump(exitJump)
	c.EmitByte(OP_POP)
}

func (c *Compiler) EmitLoop(loopStart int) {
	c.EmitByte(OP_LOOP)

	offset := len(c.currentChunk().Code) - loopStart + 2
	if offset > 65535 {
		c.Error("Loop body too large.")
	}

	c.EmitByte(byte((offset >> 8) & 0xff))
	c.EmitByte(byte(offset & 0xff))
}

func (c *Compiler) forStatement() {
	c.BeginScope()
	c.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'for'.")
	if c.Match(TOKEN_SEMICOLON) {
		// No initializer.
	} else if c.Match(TOKEN_VAR) {
		c.VarDeclaration()
	} else {
		c.ExpressionStatement()
	}

	loopStart := len(c.currentChunk().Code)
	//c.consume(TOKEN_SEMICOLON, "Expect ';'.")

	exitJump := -1
	if !c.Match(TOKEN_SEMICOLON) {
		c.expression()
		c.consume(TOKEN_SEMICOLON, "Expect ';' after loop condition.")

		// Jump out of the loop if the condition is false.
		exitJump = c.emitJump(OP_JUMP_IF_FALSE)
		c.EmitByte(OP_POP) // Condition.
	}

	//c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")
	if !c.Match(TOKEN_RIGHT_PAREN) {
		bodyJump := c.emitJump(OP_JUMP)
		incrementStart := len(c.currentChunk().Code)
		c.expression()
		c.EmitByte(OP_POP)
		c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")

		c.EmitLoop(loopStart)
		loopStart = incrementStart
		c.PatchJump(bodyJump)
	}
	c.Statement()
	c.EmitLoop(loopStart)
	if exitJump != -1 {
		c.PatchJump(exitJump)
		c.EmitByte(OP_POP) // Condition.
	}
	c.EndScope()
}

func (c *Compiler) FunDeclaration() {
	funName := c.parseVariable("Expect function name")
	c.markInitialized()
	c.fun(FUNK_TYPE_FUNCTION)
	c.DefineVariable(funName)
}

func (c *Compiler) fun(ttype FuncType) {
	cmp := NewCompiler(c.debug, c.vm, ttype, c.scanner, c)
	cmp.BeginScope()

	cmp.consume(TOKEN_LEFT_PAREN, "Expect '(' after function name.")
	if !cmp.Check(TOKEN_RIGHT_PAREN) {
		for {
			cmp.function.Arity++
			if cmp.function.Arity > 255 {
				c.errorAtCurrent("Can't have more than 255 parameters.")
			}
			constant := cmp.parseVariable("Expect parameter name.")
			cmp.DefineVariable(constant)
			if !cmp.Match(TOKEN_COMMA) {
				break
			}
		}
	}
	cmp.consume(TOKEN_RIGHT_PAREN, "Expect ')' after parameters.")
	cmp.consume(TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	cmp.Block()

	function := cmp.endCompiler()

	c.EmitBytes(OP_CONSTANT, c.makeConstant(objVal(function, c.vm.objects)))
}

func (c *Compiler) returnStatement() {
	if c.funcType == FUNK_TYPE_SCRIPT {
		c.Error("Can't return from top-level code.")
	}
	if c.Match(TOKEN_SEMICOLON) {
		c.emitReturn()
		return
	}
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect ';' after return value.")
	c.EmitByte(OP_RETURN)
}
