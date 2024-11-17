package inner

import "fmt"

func Compile(source string) {
	s := NewScanner(source)

	line := -1

	for {
		token := s.scanToken()
		if token.Line != line {
			fmt.Printf("%4d ", token.Line)
			line = token.Line
		} else {
			print("   | ")
		}
		fmt.Printf("%2d '%s'\n", token.Type, token.Source)

		if token.Type == TOKEN_EOF {
			break
		}
	}
}
