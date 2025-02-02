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
	var debug = true
	var lineRun = false

	for _, arg := range os.Args[1:] {
		argTrimmed := strings.ToLower(strings.TrimSpace(arg))
		if argTrimmed == "-d" || argTrimmed == "--debug" {
			debug = true
		} else if argTrimmed == "-l" {
			lineRun = true
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
		repl(vm)
	case 1:
		if lineRun {
			runLine(vm, newArgs[0])
			return
		}
		runFile(vm, newArgs[0])
	default:
		print("Usage: glox [path]\n")
		os.Exit(64)
	}
}

func repl(vm *inner.Vm) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("=> ")
		line, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error while reading line: " + err.Error())
		}

		vm.Interpret(line)
	}
}

func runFile(vm *inner.Vm, path string) {
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

func runLine(vm *inner.Vm, code string) {
	result := vm.Interpret(code + "\n")

	if result == inner.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if result == inner.INTERPRET_RUNTIME_ERROR {
		os.Exit(70)
	}
}
