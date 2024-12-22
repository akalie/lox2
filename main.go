package main

import (
	"bufio"
	"fmt"
	"lox2/inner"
	"os"
	"strings"
)

func main() {
	var newArgs []string
	var debug = false
	// pare
	for _, arg := range os.Args[1:] {
		argTrimmed := strings.ToLower(strings.TrimSpace(arg))
		if argTrimmed == "-d" || argTrimmed == "--debug" {
			debug = true
		} else {
			newArgs = append(newArgs, argTrimmed)
		}
	}

	vm := inner.NewVm(inner.NewChunk(), debug)

	switch len(newArgs) {
	case 0:
		if debug {
			fmt.Println("Debug is on")
		}
		repl(vm, debug)
	case 1:
		runFile(vm, newArgs[0], debug)
	default:
		print("Usage: glox [path]\n")
		os.Exit(64)
	}
}

func repl(vm *inner.Vm, debug bool) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("=> ")
		line, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error while reading line: " + err.Error())
		}

		//line = strings.Trim(line, ";\n")
		vm.Interpret(line)
	}
}

func runFile(vm *inner.Vm, path string, debug bool) {
	source, err := os.ReadFile(path)
	if err != nil {
		println("No such file!")
		os.Exit(1)
	}

	result := vm.Interpret(string(source))

	if result == inner.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if result == inner.INTERPRET_RUNTIME_ERROR {
		os.Exit(70)
	}

}
