package utils

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
)

func CheckIfStringIsNumber(v string) bool {
	if _, err1 := strconv.Atoi(v); err1 == nil {
		return true
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return true
	}
	return false
}

func Float64ToByte(f float64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, f)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}
