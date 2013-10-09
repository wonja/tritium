package bytecode

import (
  pb "code.google.com/p/goprotobuf/proto"

  tp "tritium/proto"
)

const (
	LOAD_PTR = iota
	LOAD_STRING
	LOAD_POS
	LOCAL
	SAVE_CTX
	CALL
	RETURN
)

const (
	TOP = iota
	BOTTOM
	BEFORE
	AFTER
)

var funcOffsets    []int // maps function ID to offset in bytecode
var stringData     []byte
var output         []byte
var entryPoints    []int

type Generator struct {
	transformers  []*tp.Transformer
	functions     []*tp.Function

	funcOffsets   []int
	scriptOffsets []int
	Code          []byte
	entryPoints   []byte
	stringData    []byte
}

func (g *Generator) Generate(slug *tp.Slug) {
	g.transformers := slug.GetTransformers()
	g.functions    := transformers[0].GetPkg().GetFunctions()

	g.emitBytes([]byte("Moovweb Bytecode")) // signature
	g.emitBytes(make([]byte, 20))           // space for entry points

	g.generateFunctions()

	g.entryPoints = make([]int, 5)
	for i, transformer := range g.transformers {
		g.scriptOffsets = make([]int, 0)
		g.entryPoints[i] = len(g.Code)
		g.generateTransformer(transformer)
	}

	entryPoints[4] = len(g.Code)
	g.emitBytes(stringData)

	for i, ep := range g.entryPoints {
		g.emitValueInto(16+i*4, ep, 4)
	}
}

func (g *Generator) generateFunctions() {
	for _, f := range g.functions {
		g.funcOffsets = append(g.funcOffsets, len(g.Code))
		g.generateFunction(f)
	}
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

func g.generateText(ins *tp.Instruction) {
	g.emitOpcode(LOAD_STRING)
	g.emitValue(word(len(g.stringData)), 3)
	theString := ins.GetValue()
	g.accumulateString(theString)
}

func generatePos(ins *tp.Instruction) {
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

func g.generateCall(ins *tp.Instruction) {
	g.generateBlock(ins)
	for _, arg := range ins.Arguments {
		g.generateInstruction(arg)
	}
	id := int(ins.GetFunctionId())
	g.emitOpcode(CALL)
	g.emitValue(word(g.funcOffsets[id]), 3)
}

func g.generateBlock(ins *tp.Instruction) {
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
		g.emitOpcode(RETURN)
		g.emitValue(word(0), 3) // let's just keep it 32 bits per instruction
		// jump over the block
		g.emitValueInto(beforeBlock+1, word(len(g.Code)), 3)
	}
}

func (g *Generator) generateFunction(f *tp.Function) {
	g.generateInstruction(f.GetInstruction())
	g.emitOpcode(RETURN)
	g.emitValue(word(0), 3)
}

