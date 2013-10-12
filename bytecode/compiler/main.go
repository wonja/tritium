package main

import (
	"io/ioutil"
	"code.google.com/p/goprotobuf/proto"
	"os"

	"tritium/parser"
	tp "tritium/proto"
	k "tritium/constants"
	"tritium/bytecode"
)

var namesToIds map[string]int

func main() {
	wd, _ := os.Getwd()
	lib := parser.ParseFile(wd, "", "lib.ts", false)
	src := parser.ParseFile(wd, "", "main.ts", false)

	namesToIds = make(map[string]int, 0)
	for i, f := range lib.Functions {
		namesToIds[f.GetName()] = i
	}
	println("compiling sample code")
	for _, f := range lib.Functions {
		link(f.GetInstruction(), lib.Functions)
	}
	link(src.GetRoot(), lib.Functions)
	println()
	println("generating bytecode")
	g := &bytecode.Generator{
		Script: src.GetRoot(),
		Functions: lib.Functions,
		FuncOffsets: make([]int, 0),
		Code: make([]byte, 0),
		StringData: make([]byte, 0),
	}
	g.Generate()
	ioutil.WriteFile("main.bs", g.Code, os.FileMode(0666))
}

func link(ins *tp.Instruction, functions []*tp.Function) {
	if ins == nil {
		return
	}
	switch ins.GetType() {
	case k.Instruction_FUNCTION_CALL:
		println("linking call", ins.GetValue(), "to", namesToIds[ins.GetValue()])
		ins.FunctionId = proto.Int32(int32(namesToIds[ins.GetValue()]))
	case k.Instruction_IMPORT:
	case k.Instruction_TEXT:
	case k.Instruction_LOCAL_VAR:
	case k.Instruction_POSITION:
	case k.Instruction_COMMENT:
	case k.Instruction_BLOCK:
	default:
	}
	for _, child := range ins.Children {
		link(child, functions)
	}
}