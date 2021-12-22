package main

func setNineByte(rule byte, tocken int64) []byte {
	var buf []byte = make([]byte, 9)
	buf[0] = rule
	for iter := 1; iter < 9; iter++ {
		buf[iter] = byte(tocken >> ((8 - iter) * 8))
	}
	return buf
}

func setNineByteAll(rule byte, tocken []byte) []byte {
	return append([]byte{rule}, tocken...)
}

func equals(buf []byte, tocken int64) bool {
	if len(buf) != 8 {
		return false
	}
	for index, value := range buf {
		if value != byte(tocken>>((7-index)*8)) {
			return false
		}
	}
	return true
}

func comper(first []byte, second []byte) bool {
	if len(first) != len(second) {
		return false
	}
	for index, value := range first {
		if value != second[index] {
			return false
		}
	}
	return true
}

func convertToBytes(tocken int64) []byte {
	buf := make([]byte, 8)
	for iter := 0; iter < 8; iter++ {
		buf[iter] = byte(tocken >> ((7 - iter) * 8))
	}
	return buf
}

func convertToTocken(buf []byte) int64 {
	var tocken int64
	for iter := 0; iter < 8; iter++ {
		tocken <<= 8
		tocken |= int64(buf[iter])
	}
	return tocken
}
