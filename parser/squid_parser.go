package parser

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	tp "tritium/proto"
	. "tritium/parser/tokenizer"
)

type Parser struct {
	*Tokenizer
	ProjectPath string // the project path (probably absolute)
	ScriptPath  string // the folder containing the script file being parsed (relative to the project path)
	FileName    string // the base-name of the script file being parsed
	FullPath    string
	Lookahead   *Token
	counter     int
	header      bool
	RootFile    bool
}

func (p *Parser) gensym() string {
	p.counter++
	return p.FullPath + string(p.counter)
}

func (p *Parser) peek() *Token {
	return p.Lookahead
}

func (p *Parser) pop() *Token {
	val := p.Lookahead
	p.Lookahead = p.Tokenizer.Pop()
	return val
}

func (p *Parser) error(msg string) {
	formatString := "%s:%d -- %s; found unexpected %s"
	numVals := 0
	val1 := ""
	val2 := ""
	if p.peek().Value != "" {
		formatString += ": %s"
		numVals++
		val1 = p.peek().Value
		if p.peek().Lexeme == STRING {
			val1 = strconv.Quote(val1)
		}
		if p.peek().ExtraValue != "" {
			formatString += ", %s"
			numVals++
			val2 = p.peek().ExtraValue
		}
	}
	var fullMsg string
	switch numVals {
	case 0:
		fullMsg = fmt.Sprintf(formatString,
			p.FullPath,
			p.peek().LineNumber,
			msg,
			LexemeName[p.peek().Lexeme])
	case 1:
		fullMsg = fmt.Sprintf(formatString,
			p.FullPath,
			p.peek().LineNumber,
			msg,
			LexemeName[p.peek().Lexeme],
			val1)
	case 2:
		fullMsg = fmt.Sprintf(formatString,
			p.FullPath,
			p.peek().LineNumber,
			msg,
			LexemeName[p.peek().Lexeme],
			val1,
			val2)
	}
	panic(fullMsg)
}

func MakeParser(src, projectPath, scriptPath, fileName string, isRootFile bool) *Parser {
	fullpath := filepath.Join(projectPath, scriptPath, fileName)
	fullpath, _ = filepath.Abs(fullpath)
	scriptPath = filepath.Clean(scriptPath)
	projectPath = filepath.Clean(projectPath)
	p := &Parser{
		Tokenizer:   MakeTokenizer([]byte(src)),
		ProjectPath: projectPath, // the project path (probably absolute)
		ScriptPath:  scriptPath,  // the folder containing the script file being parsed (relative to the project path)
		FileName:    fileName,    // the base-name of the script file being parsed
		FullPath:    fullpath,
		Lookahead:   nil,
		counter:     0,
		RootFile:    isRootFile,
	}
	p.pop()
	return p
}

func (p *Parser) Parse() *tp.ScriptObject {
	script := new(tp.ScriptObject)
	// script.Name = proto.String(p.FullPath)
	if !p.RootFile {
		script.Name = proto.String(filepath.Join(p.ScriptPath, p.FileName))
	} else {
		script.Name = proto.String("__rewriter__")
	}

	stmts := tp.ListInstructions()
	defs := make([]*tp.Function, 0) // Add a new constructor in instruction.go

	for p.peek().Lexeme != EOF {
		switch p.peek().Lexeme {
		case FUNC:
			defs = append(defs, p.definition())

			// Is this still necessary, now that new doc tools are being developed?
			if len(stmts) > 0 {
				previousStatement := stmts[len(stmts)-1]
				if *previousStatement.Type == tp.Instruction_TEXT {
					defs[len(defs)-1].Description = previousStatement.Value
					if len(stmts) > 1 {
						stmts = stmts[:len(stmts)-2]
					} else if len(stmts) == 1 {
						stmts = stmts[0:0]
					}
				}
			}
		default:
			stmt := p.statement()
			stmts = append(stmts, stmt)
			// need to intersperse imports with definitions
			if tp.Instruction_InstructionType_name[int32(stmt.GetType())] == "IMPORT" {
				// Make a special function stub that represents the import.
				// Need to do this because we can't mix definitions and instructions in
				// the same array.
				imp := new(tp.Function)
				imp.Name = proto.String("@import")
				imp.Description = proto.String(stmt.GetValue())
				defs = append(defs, imp)
			}
		}
	}

	if len(defs) == 0 {
		defs = nil
	}

	var line int32
	if len(stmts) == 0 {
		stmts = nil
	} else {
		line = *stmts[0].LineNumber
	}

	script.Functions = defs
	script.Root = tp.MakeBlock(stmts, line)

	return script

	// switch p.peek().Lexeme {
	// case FUNC:
	//  for p.peek().Lexeme != EOF {
	//    defs = append(defs, p.definition())
	//  }
	//  if len(defs) == 0 {
	//    defs = nil
	//  }
	//  script.Functions = defs
	// default:
	//  for p.peek().Lexeme != EOF {
	//    stmts = append(stmts, p.statement())
	//  }
	//  line := int32(0)
	//  if len(stmts) == 0 {
	//    stmts = nil
	//  } else {
	//    line = *stmts[0].LineNumber
	//  }
	//  script.Root = tp.MakeBlock(stmts, line)
	// }
	// return script
}

func (p *Parser) statement() (node *tp.Instruction) {
	switch p.peek().Lexeme {
	case IMPORT:
		token := p.pop() // pop the "@import" token (includes importee)
		scriptLocationInProject := filepath.Clean(filepath.Join(p.ScriptPath, token.Value))

		// extract the root script folder from the relative path of the importee
		// (would be easier if filepath.FromSlash worked as advertised)
		dir, base := filepath.Split(p.ScriptPath)
		if len(dir) == 0 {
			dir = base
			base = ""
		}
		if dir[len(dir)-1] == os.PathSeparator {
			dir = dir[0:len(dir)-1]
		}
		for len(base) > 0 {
			dir, base = filepath.Split(dir)
			if len(dir) == 0 {
				dir = base
				base = ""
			}
			if dir[len(dir)-1] == os.PathSeparator {
				dir = dir[0:len(dir)-1]
			}
		}
		// if dir == "." {
		// 	wd, wdErr := os.Getwd()
		// 	if wdErr != nil {
		// 		msg := fmt.Sprintf("%s:%d -- @import could not determine current working directory!", p.FileName, token.Lexeme)
		// 		panic(msg)
		// 	}
		// 	dir = wd
		// }
		// make sure that the importee is under the right subfolder
		if !strings.HasPrefix(scriptLocationInProject, dir) {
			msg := fmt.Sprintf("%s:%d -- imported file must exist under the `%s` folder", p.FileName, token.LineNumber, dir)
			panic(msg)
		}
		node = tp.MakeImport(scriptLocationInProject, token.LineNumber)
		// println("PROCESSED IMPORT", dir, base, scriptLocationInProject)
	case STRING, REGEXP, POS, READ, ID, TYPE, GVAR, LVAR, LPAREN:
		node = p.expression()
	default:
		p.error("statement must consist of import or expression")
	}
	return node
}

func (p *Parser) expression() (node *tp.Instruction) {
	node = p.term()
	rest := tp.ListInstructions()
	for p.peek().Lexeme == PLUS {
		p.pop() // pop the plus sign
		switch p.peek().Lexeme {
		case STRING, REGEXP, POS, READ, ID, TYPE, GVAR, LVAR, LPAREN:
			rest = append(rest, p.term())
		default:
			p.error("argument to + must be a self-contained expression")
		}
	}
	if len(rest) > 0 {
		node = tp.FoldLeft("concat", node, rest)
	}
	return node
}

func (p *Parser) term() (node *tp.Instruction) {
	switch p.peek().Lexeme {
	case STRING, REGEXP, POS:
		node = p.literal()
	case READ:
		node = p.read()
	case ID:
		node = p.call()
	case TYPE:
		node = p.cast()
	case GVAR, LVAR:
		node = p.variable()
	case LPAREN:
		p.pop() // pop the lparen
		node = p.expression()
		if p.peek().Lexeme != RPAREN {
			p.error("unclosed parenthesis")
		}
		p.pop() // pop the rparen
	default:
		// will never get to this point because p.statement() and p.expression() check first
		p.error("expression begins with invalid element")
	}
	return node
}

func (p *Parser) literal() (node *tp.Instruction) {
	token := p.pop()
	switch token.Lexeme {
	case STRING:
		node = tp.MakeText(token.Value, token.LineNumber)
	case REGEXP:
		node = tp.MakeFunctionCall("regexp",
			tp.ListInstructions(tp.MakeText(token.Value, token.LineNumber),
				tp.MakeText(token.ExtraValue, token.LineNumber)),
			nil,
			token.LineNumber)
	case POS:
		node = tp.MakePosition(token.Value, token.LineNumber)
	}
	return node
}

func (p *Parser) read() (node *tp.Instruction) {
	p.pop() // pop the "read" keyword
	readLineNo := p.peek().LineNumber
	if p.peek().Lexeme != LPAREN {
		p.error("argument list expected for read")
	}
	p.pop() // pop the lparen
	if p.peek().Lexeme != STRING {
		p.error("read requires a literal string argument")
	}
	readPath := p.pop().Value
	if p.peek().Lexeme != RPAREN {
		p.error("unterminated argument list in read")
	}
	p.pop() // pop the rparen

	// make sure we're not trying to read outside the project folder
	fullReadPath := filepath.Clean(filepath.Join(p.ProjectPath, p.ScriptPath, readPath))
	absReadPath, err := filepath.Abs(fullReadPath)
	if err != nil {
		msg := fmt.Sprintf("%s:%d -- `read` could not resolve the full path to %s", p.FileName, readLineNo, readPath)
		panic(msg)
	}
	if !strings.HasPrefix(absReadPath, filepath.Join(p.ProjectPath)) {
		msg := fmt.Sprintf("%s:%d -- `read` cannot open files outside the project folder", p.FileName, readLineNo)
		panic(msg)
	}

	contents, err := ioutil.ReadFile(filepath.Join(p.ProjectPath, p.ScriptPath, readPath))
	if err != nil { // can't use p.error because it's not a syntax error
		msg := fmt.Sprintf("%s:%d -- `read` could not open %s", p.FileName, readLineNo, readPath)
		panic(msg)
	}
	node = tp.MakeText(string(contents), readLineNo)
	return node
}

func (p *Parser) call() (node *tp.Instruction) {
	funcName := p.pop().Value // grab the function name
	funcLineNo := p.peek().LineNumber
	if p.peek().Lexeme != LPAREN {
		p.error("parenthesized argument list expected in call to " + funcName)
	}
	p.pop() // pop the lparen

	ords, kwdnames, kwdvals := p.arguments(funcName) // gather the arguments
	numArgs := len(ords)

	// TO DO: integrate this block for better variadic concat/log expansions
	// if funcName == "concat" && numArgs > 2 {
	//   // expand variadic concat into nested binary concats
	//   lhs := tp.FoldLeft("concat", ords[0], ords[1:numArgs-1])
	//   rhs := ords[numArgs-1]
	//   node = tp.MakeFunctionCall("concat", tp.ListInstructions(lhs,rhs), block, funcLineNo)
	// } else if funcName == "log" && numArgs > 1 {
	//   // expand variadic log into composition of log and concat
	//   cats := tp.FoldLeft("concat", ords[0], ords[1:])
	//   node = tp.MakeFunctionCall("log", tp.ListInstructions(cats), block, funcLineNo)
	// }

	// this will never happen because p.arguments() only returns when it encounters an rparen
	if p.peek().Lexeme != RPAREN {
		p.error("unterminated argument list in call to " + funcName)
	}
	p.pop() // pop the rparen
	var block []*tp.Instruction
	if p.peek().Lexeme == LBRACE {
		block = p.block()
	}

	// Expand keyword args
	if kwdnames != nil && kwdvals != nil {
		kwdToGensym := make(map[string]string, len(kwdnames))
		outer := tp.ListInstructions()
		for i, k := range kwdnames {
			tempname := p.gensym()
			tempvar := tp.MakeFunctionCall("var",
				tp.ListInstructions(tp.MakeText(tempname, funcLineNo),
					kwdvals[i]),
				nil, funcLineNo)
			outer = append(outer, tempvar)
			kwdToGensym[k] = tempname
		}
		inner := tp.ListInstructions()
		for _, k := range kwdnames {
			getter := tp.MakeFunctionCall("var",
				tp.ListInstructions(tp.MakeText(kwdToGensym[k], funcLineNo)),
				nil, funcLineNo)
			setter := tp.MakeFunctionCall("set",
				tp.ListInstructions(tp.MakeText(k, funcLineNo), getter),
				nil, funcLineNo)
			inner = append(inner, setter)
		}
		if block != nil {
			for _, v := range block {
				inner = append(inner, v)
			}
		}
		theCall := tp.MakeFunctionCall(funcName, ords, inner, funcLineNo)
		outer = append(outer, theCall)
		node = tp.MakeBlock(outer, funcLineNo)

	} else if funcName == "concat" && numArgs > 2 {
		// expand variadic concat into nested binary concats
		lhs := tp.FoldLeft("concat", ords[0], ords[1:numArgs-1])
		rhs := ords[numArgs-1]
		node = tp.MakeFunctionCall("concat", tp.ListInstructions(lhs, rhs), block, funcLineNo)
	} else if funcName == "log" && numArgs > 1 {
		// expand variadic log into composition of log and concat
		cats := tp.FoldLeft("concat", ords[0], ords[1:])
		node = tp.MakeFunctionCall("log", tp.ListInstructions(cats), block, funcLineNo)
	} else {
		node = tp.MakeFunctionCall(funcName, ords, block, funcLineNo)
	}
	return node
}

func (p *Parser) arguments(funcName string) (ords []*tp.Instruction, kwdnames []string, kwdvals []*tp.Instruction) {
	ords, kwdnames, kwdvals = make([]*tp.Instruction, 0), make([]string, 0), make([]*tp.Instruction, 0)
	counter := 0
	for p.peek().Lexeme != RPAREN {
		if counter > 0 {
			if p.peek().Lexeme != COMMA {
				p.error("arguments must be separated by commas")
			}
			p.pop() // pop the comma
		}
		if p.peek().Lexeme == KWD {
			k := p.pop().Value
			kwdnames = append(kwdnames, k)
			// TO DO: CHECK EXPRESSION FIRST-SET
			switch p.peek().Lexeme {
			case STRING, REGEXP, POS, READ, ID, TYPE, GVAR, LVAR, LPAREN:
				kwdvals = append(kwdvals, p.expression())
			default:
				p.error("value for keyword argument " + strconv.Quote(k) + " in call to " + funcName + " is not a valid expression")
			}
		} else {
			switch p.peek().Lexeme {
			case STRING, REGEXP, POS, READ, ID, TYPE, GVAR, LVAR, LPAREN:
				ords = append(ords, p.expression())
			default:
				p.error("value for ordinal argument in call to " + funcName + " is not a valid expression")
			}
		}
		counter++
	}
	if len(ords) == 0 {
		ords = nil
	}
	if len(kwdnames) == 0 || len(kwdvals) == 0 {
		kwdnames = nil
		kwdvals = nil
	}
	return ords, kwdnames, kwdvals
}

func (p *Parser) cast() (node *tp.Instruction) {
	typeName := p.pop().Value // grab the function name
	typeLineNo := p.peek().LineNumber
	if p.peek().Lexeme != LPAREN {
		p.error("parenthesized argument needed for typecast to " + typeName)
	}
	p.pop() // pop the lparen
	expr := p.expression()
	if p.peek().Lexeme != RPAREN {
		p.error("single argument to " + typeName + " typecast is missing closing parenthesis")
	}
	p.pop() // pop the rparen
	var block []*tp.Instruction
	if p.peek().Lexeme == LBRACE {
		block = p.block()
	}

	node = tp.MakeFunctionCall(typeName, tp.ListInstructions(expr), block, typeLineNo)
	return node
}

func (p *Parser) variable() (node *tp.Instruction) {
	token := p.pop()
	lexeme, name, lineNo := token.Lexeme, token.Value, token.LineNumber
	sigil := "$"
	if lexeme == LVAR {
		sigil = "%"
	}
	var val *tp.Instruction
	var block []*tp.Instruction
	if p.peek().Lexeme == EQUAL {
		p.pop() // pop the equal sign
		switch p.peek().Lexeme {
		case STRING, REGEXP, POS, READ, ID, TYPE, GVAR, LVAR, LPAREN:
			val = p.expression()
		default:
			p.error("invalid expression in assignment to " + sigil + name)
		}
	}
	if p.peek().Lexeme == LBRACE {
		block = p.block()
	}
	if lexeme == LVAR {
		node = tp.MakeLocalVar(name, val, block, lineNo)
	} else {
		args := tp.ListInstructions(tp.MakeText(name, lineNo))
		if val != nil {
			args = append(args, val)
		}
		node = tp.MakeFunctionCall("var", args, block, lineNo)
	}
	return node
}

func (p *Parser) block() (stmts []*tp.Instruction) {
	stmts = tp.ListInstructions()
	p.pop() // pop the lbrace
	for p.peek().Lexeme != RBRACE {
		stmts = append(stmts, p.statement())
	}
	p.pop() // pop the rbrace
	if len(stmts) == 0 {
		stmts = nil
	}
	return stmts
}

func (p *Parser) definition() *tp.Function {
	isSignature := false
	node := new(tp.Function)

	funcLineNo := p.pop().LineNumber // pop the `@func` keyword
	contextType := ""
	if p.peek().Lexeme == TYPE {
		contextType = p.pop().Value
		if p.peek().Lexeme != DOT {
			p.error("function context and function name must be separated by '.'")
		}
		p.pop() // pop the dot
	}

	if p.peek().Lexeme != ID {
		p.error("invalid function name in definition")
	}

	funcName := p.pop().Value
	funcFile := ""
	if len(p.ScriptPath) > 0 && p.ScriptPath != "." {
		funcFile = filepath.Join(p.ScriptPath, p.FileName)
	}

	if p.peek().Lexeme != LPAREN {
		p.error("parenthesized parameter list expected in function declaration")
	}
	p.pop() // pop the lparen
	params := p.parameters(funcName)
	if len(params) == 0 {
		params = nil
	}
	p.pop() // pop the rparen

	returnType := ""
	opensType := ""
	if p.peek().Lexeme == TYPE {
		isSignature = true
		returnType = p.pop().Value
		if p.peek().Lexeme == TYPE {
			opensType = p.pop().Value
		}
	}

	node.Name = proto.String(funcName)
	if len(funcFile) > 0 {
		node.Filename = proto.String(funcFile)
	}
	node.LineNumber = proto.Int32(funcLineNo)
	node.Args = params
	node.ScopeType = proto.String(contextType)
	node.ReturnType = proto.String(returnType)
	node.OpensType = proto.String(opensType)

	if isSignature {
		if p.peek().Lexeme == LBRACE {
			p.error("body not permitted in signature for " + funcName)
		}
		node.BuiltIn = proto.Bool(true)
		return node
	}
	node.BuiltIn = proto.Bool(false)
	if p.peek().Lexeme != LBRACE {
		p.error("definition for " + funcName + " is missing a body")
	}
	funcBody := &tp.Instruction{
		Type: tp.Instruction_BLOCK.Enum(),
		// Children: p.block(),
		// use the wrapper to get a better error message
		Children: p.function_body(*node.Name),
	}
	node.Instruction = funcBody
	return node
}

func (p *Parser) function_body(funcName string) (stmts []*tp.Instruction) {
	// catch a parsing error and add extra error info about the surrounding definition
	defer func() {
		if r := recover(); r != nil {
			// pull out the actual message without the filename/line-no
			msg := strings.Split(strings.Split(r.(string), "-- ")[1], "; ")[0]
			// re-throw the error with the current filename, line-no, and function-name
			p.error(fmt.Sprintf("in function '%s': %s", funcName, msg))
		}
	}()

	stmts = p.block()
	return stmts
}

func (p *Parser) parameters(funcName string) []*tp.Function_Argument {
	params := make([]*tp.Function_Argument, 0)
	counter := 0
	for p.peek().Lexeme != RPAREN {
		if counter > 0 {
			if p.peek().Lexeme != COMMA {
				p.error("parameter list for " + funcName + " must be separated by commas")
			}
			p.pop() // pop the comma
		}
		if p.peek().Lexeme != TYPE {
			p.error("parameter for " + funcName + " is missing a type")
		}
		param := &tp.Function_Argument{
			TypeString: proto.String(p.pop().Value),
		}
		if p.peek().Lexeme != LVAR {
			p.error("parameter for " + funcName + " has invalid name")
		}
		param.Name = proto.String(p.pop().Value)
		params = append(params, param)
		counter++
	}
	return params
}
