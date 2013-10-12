package main

import (
	"io/ioutil"
	"strings"
	"tritium/bytecode"
)

func main() {
	indent := "    "
	prog, _ := ioutil.ReadFile("main.bs")
	sig := string(prog[0:16])
	println("leading bytes =",sig)
	if sig != "Moovweb Bytecode" {
		println("not Moovweb bytecode!")
		return
	}
	println("confirmed Moovweb bytecode")
	codeStart := bytecode.DecodeBytesAt(prog, 16, 4)
	stringStart := bytecode.DecodeBytesAt(prog, 20, 4)
	strs := prog[stringStart:len(prog)]

	pc := int(codeStart)
	stack := make([]int, 1024)
	sp := 0
	frames := 0
	prog = prog[0:stringStart]
	last := len(prog)
	println()

	for pc < last {
		opcode := bytecode.DecodeBytesAt(prog, pc, 1)
		value := bytecode.DecodeBytesAt(prog, pc+1, 3)
		switch opcode {
		case bytecode.LOAD_STRING:
			length := bytecode.DecodeBytesAt(strs, int(value), 4)
			the_string := string(strs[value+4:value+4+length])
			println(strings.Repeat(indent, frames), "STRING:", the_string)
			pc = pc + 4
		case bytecode.JUMP:
			println(strings.Repeat(indent, frames), "JUMP:", value)
			pc = int(value)
		case bytecode.CALL:
			stack[sp] = pc+4 // save return address
			sp++
			pc = int(value)
			println(strings.Repeat(indent, frames), "CALL:", value, "(return to", stack[sp-1], ")")
			frames++
		case bytecode.YIELD:
			sp--
			println(strings.Repeat(indent, frames), "YIELD:", stack[sp-2], "(return to", pc+4, ")")
			frames++
			stack[sp] = pc+4
			sp++
			pc = stack[sp-3]
		case bytecode.UNYIELD:
			println(strings.Repeat(indent, frames), "UNYIELD:", stack[sp-1])
			pc = stack[sp-1]
			sp--
			frames--
		case bytecode.RETURN:
			sp--
			pc = stack[sp]
			sp--
			println(strings.Repeat(indent, frames), "RETURN:", pc)
			frames--
		case bytecode.LOAD_PTR:
			stack[sp] = int(value)
			sp++
			println(strings.Repeat(indent, frames), "LOAD PTR:", value)
			pc = pc + 4
		default:
			println(strings.Repeat(indent, frames), "NOT IMPLEMENTED")
			pc = pc + 4
		}
	}
	println("stack size:", sp)
}

func decode4(quad []byte) int {
	val := uint32(quad[0])
	println(quad[0])
	println(quad[1])
	println(quad[2])
	println(quad[3])
	val = val + (uint32(quad[1]) << 8)
	val = val + (uint32(quad[2]) << 16)
	val = val + (uint32(quad[3]) << 24)
	return int(val)
}