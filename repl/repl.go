package repl

import (
	"bufio"
	"comp/compiler"
	"comp/object"
	"comp/parser"
	"comp/vm"
	"fmt"
	"io"

	"comp/lexer"
)

const PROMPT = ">>"

// TODO: add file support with extension .sc?

func Start(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)
	// env := object.NewEnvironment()

	var (
		constants   []object.Object
		globals     = make([]object.Object, vm.GlobalsSize)
		symbolTable = compiler.NewSymbolTable()
	)
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
		cmp := compiler.NewWithState(symbolTable, constants)
		err := cmp.Compile(root)
		if err != nil {
			_, _ = fmt.Fprintf(output, "Compilation failed:\n %s\n", err)
			continue
		}
		bytecode := cmp.ByteCode()
		constants = bytecode.Constants

		vrm := vm.NewVMWithGlobalsStore(bytecode, globals)

		err = vrm.RunVM()
		if err != nil {
			_, _ = fmt.Fprintf(output, "Executing bytecode failed:\n %s\n", err)
			continue
		}
		stackTop := vrm.LastPoppedStackElement()

		_, _ = io.WriteString(output, stackTop.Inspect())
		_, _ = io.WriteString(output, "\n")
	}
}

func printParserErrors(output io.Writer, errors []string) {
	errMsg := fmt.Sprintf("%sParser ERROR::%s\n", object.COLOR_RED, object.COLOR_RESET)
	_, _ = io.WriteString(output, errMsg)

	for _, err := range errors {
		_, _ = io.WriteString(output, "\t"+err+"\n")
	}
}
