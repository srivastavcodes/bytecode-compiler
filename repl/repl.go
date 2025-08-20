package repl

import (
	"Flint-v2/compiler"
	"Flint-v2/object"
	"Flint-v2/parser"
	"Flint-v2/vm"
	"bufio"
	"fmt"
	"io"

	"Flint-v2/lexer"
)

const PROMPT = ">>"

func Start(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)
	// env := object.NewEnvironment()

	for {
		fmt.Print(PROMPT)
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
		/*		evaluated := evaluator.Evaluate(root, env)
				if evaluated != nil {
					_, _ = io.WriteString(output, evaluated.Inspect())
					_, _ = io.WriteString(output, "\n")
				}
		*/
		cmp := compiler.NewCompiler()
		err := cmp.Compile(root)
		if err != nil {
			fmt.Fprintf(output, "Compilation failed:\n %s\n", err)
			continue
		}

		vrm := vm.NewVM(cmp.ByteCode())
		err = vrm.RunVM()
		if err != nil {
			fmt.Fprintf(output, "Executing bytecode failed:\n %s\n", err)
			continue
		}

		stackTop := vrm.StackTop()
		io.WriteString(output, stackTop.Inspect())
		io.WriteString(output, "\n")
	}
}

func printParserErrors(output io.Writer, errors []string) {
	errMsg := fmt.Sprintf("%sParser ERROR::%s\n", object.COLOR_RED, object.COLOR_RESET)
	io.WriteString(output, errMsg)

	for _, err := range errors {
		io.WriteString(output, "\t"+err+"\n")
	}
}
