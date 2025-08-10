package repl

import (
	"Flint-v2/evaluator"
	"Flint-v2/object"
	"Flint-v2/parser"
	"bufio"
	"fmt"
	"io"

	"Flint-v2/lexer"
)

const PROMPT = ">>"

func Start(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)
	env := object.NewEnvironment()

	for {
		fmt.Printf(PROMPT)
		ok := scanner.Scan()
		if !ok {
			return
		}
		scanned := scanner.Text()
		lxr := lexer.NewLexer(scanned)
		psr := parser.NewParser(lxr)

		root := psr.ParseRootStatement()
		if len(psr.Errors()) != 0 {
			printParserErrors(output, psr.Errors())
			continue
		}
		evaluated := evaluator.Evaluate(root, env)
		if evaluated != nil {
			_, _ = io.WriteString(output, evaluated.Inspect())
			_, _ = io.WriteString(output, "\n")
		}
	}
}

func printParserErrors(output io.Writer, errors []string) {
	errMsg := fmt.Sprintf("%sParser ERROR::%s\n", object.COLOR_RED, object.COLOR_RESET)
	_, _ = io.WriteString(output, errMsg)

	for _, err := range errors {
		_, _ = io.WriteString(output, "\t"+err+"\n")
	}
}
