package bytecode

import (
  // "code.google.com/p/goprotobuf/proto"
  tp "tritium/proto"
  "tritium/constants"
)

const (
	LOAD_PTR = iota
	LOAD_STRING
	LOAD_POS
	LOCAL
	SAVE_CTX
	JUMP
	CALL
	YIELD
	UNYIELD
	RETURN
	HALT
)

const (
	TOP = iota
	BOTTOM
	BEFORE
	AFTER
)

type Generator struct {
	Script        *tp.Instruction
	Functions     []*tp.Function

	FuncOffsets   []int
	Code          []byte
	StringData    []byte
}

func (g *Generator) Generate() {

	g.emitBytes([]byte("Moovweb Bytecode")) // signature
	g.emitBytes(make([]byte, 8))           // space for segment pointers

	for _, f := range g.Functions {
		g.FuncOffsets = append(g.FuncOffsets, len(g.Code))
		g.generateFunction(f)
	}
	println("saving code segment pointer:", word(len(g.Code)))
	g.emitValueInto(16, word(len(g.Code)), 4)

	for _, toplevel := range g.Script.Children {
		g.generateInstruction(toplevel)
	}


	println("saving string segment pointer:", word(len(g.Code)))
	g.emitValueInto(20, word(len(g.Code)), 4)
	g.emitBytes(g.StringData)
}

func (g *Generator) generateInstruction(ins *tp.Instruction) {
	switch ins.GetType() {
	case constants.Instruction_BLOCK:
		g.generateBlock(ins)
	case constants.Instruction_FUNCTION_CALL:
		g.generateCall(ins)
	case constants.Instruction_IMPORT:
		g.generateImport(ins)
	case constants.Instruction_TEXT:
		g.generateText(ins)
	case constants.Instruction_LOCAL_VAR:
		g.generateLocal(ins)
	case constants.Instruction_POSITION:
		g.generatePos(ins)
	default:
		// should probably panic or something
	}
}

func (g *Generator) generateText(ins *tp.Instruction) {
	g.emitOpcode(LOAD_STRING)
	g.emitValue(word(len(g.StringData)), 3)
	theString := ins.GetValue()
	g.accumulateString(theString)
}

func (g *Generator) generatePos(ins *tp.Instruction) {
	g.emitOpcode(LOAD_POS)
	switch ins.GetValue() {
	case "top":
		g.emitValue(TOP, 3)
	case "bottom":
		g.emitValue(BOTTOM, 3)
	case "before":
		g.emitValue(BEFORE, 3)
	case "after":
		g.emitValue(AFTER, 3)
	default:
		g.emitValue(BOTTOM, 3)
	}
}

func (g *Generator) generateCall(ins *tp.Instruction) {
	g.generateBlock(ins)
	for _, arg := range ins.Arguments {
		g.generateInstruction(arg)
	}
	id := int(ins.GetFunctionId())
	f := g.Functions[id]
	if f.GetBuiltIn() {
		g.emitOpcode(YIELD)
		g.emitValue(word(0), 3)
	} else {
		g.emitOpcode(CALL)
		g.emitValue(word(g.FuncOffsets[id]), 3)
	}
}

func (g *Generator) generateBlock(ins *tp.Instruction) {
	beforeBlock := len(g.Code) + 4
	insideBlock := len(g.Code) + 8
	g.emitOpcode(LOAD_PTR)
	if len(ins.Children) == 0 {
		g.emitValue(word(0), 3)
	} else {
		g.emitValue(word(insideBlock), 3)
		g.emitOpcode(JUMP)
		g.emitValue(word(0), 3)
		for _, child := range ins.Children {
			g.generateInstruction(child)
		}
		g.emitOpcode(UNYIELD)
		g.emitValue(word(0), 3) // let's just keep it 32 bits per instruction
		// jump over the block
		g.emitValueInto(beforeBlock+1, word(len(g.Code)), 3)
	}
}

func (g *Generator) generateFunction(f *tp.Function) {
	if f.GetBuiltIn() {
		return
	}
	g.emitOpcode(LOAD_STRING)
	g.emitValue(word(len(g.StringData)), 3)
	g.accumulateString("header for " + f.GetName()) // load the function's name, for now
	for _, bodyIns := range f.GetInstruction().Children {
		g.generateInstruction(bodyIns)
	}
	g.emitOpcode(RETURN)
	g.emitValue(word(0), 3)
}

func (g *Generator) generateLocal(ins *tp.Instruction) {
	// just store the name for now
	g.emitOpcode(LOCAL)
	g.emitValue(word(len(g.StringData)), 3)
	theString := ins.GetValue()
	g.accumulateString(theString)
}

func (g *Generator) generateImport(ins *tp.Instruction) {

}