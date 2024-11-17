package main

import (
	"bufio"
	"fmt"
	"lox2/inner"
	"os"
)

func main() {
	vm := inner.NewVm(inner.NewChunk())

	switch len(os.Args) {
	case 1:
		repl(vm)
	case 2:
		runFile(vm, os.Args[1])
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

		//line = strings.Trim(line, ";\n")
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
