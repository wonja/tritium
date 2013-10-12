package bytecode

type word uint32
const bytesPerWord = 4

func appendBytes(destination []byte, value word, size int) []byte {
	for i := 0; i < size; i++ {
		destination = append(destination, byte(value >> uint(i*8)))
	}
	return destination
}

func encodeBytesAt(destination []byte, position int, value word, size int) {
	buf := make([]byte, 0)
	buf = appendBytes(buf, value, size)
	for i, b := range buf {
		destination[position+i] = b
	}
}

func DecodeBytesAt(buffer []byte, position int, size int) word {
	var value word
	for i := 0; i < size; i++ {
		value = value | (word(buffer[position+i]) << uint(i*8))
	}
	return value
}