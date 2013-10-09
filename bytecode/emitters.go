package bytecode

func (g *Generator) emitBytes(slice []byte) {
	g.Code = append(g.Code, slice...)
}

func (g *Generator) emitValue(value word, size int) {
	g.Code = appendBytes(g.Code, value, size)
}

func (g *Generator) emitOpcode(opcode int) {
	g.Code = append(g.Code, byte(opcode))
}

func (g *Generator) emitValueInto(position int, value word, size int) {
	encodeBytesAt(g.Code, position, value, size)
}

func (g *Generator) accumulateString(str string) {
	g.stringData = appendBytes(g.stringData, word(len(str)), 4)
	g.stringData = append(g.stringData, []byte(str)...)
}

func (g *Generator) emitStringSegment() {
	g.emitBytes(g.stringData)
}

// [opcode..][24-bit constant.........]
func (g *Generator) emitOpK(opcode, arg int) {
	g.Code = append(g.Code, byte(opcode))
	g.emitValue(word(arg), 3)
}

// [opcode..][reg1....]
func (g *Generator) emitOp1(opcode, arg int) {
	g.Code = append(g.Code, byte(opcode))
	g.emitValue(word(arg), 1)
}

// [opcode..][reg1....][reg2....]
func (g *Generator) emitOp2(opcode, arg1, arg2 int) {
	g.Code = append(g.Code, byte(opcode))
	g.emitValue(word(arg), 1)
	g.emitValue(word(arg), 1)
}