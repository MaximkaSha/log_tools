// Package utils provide common utils for agent and server.
package utils

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
)

// CheckIfStringIsNumber - return true if string is float64 or int64 value.
func CheckIfStringIsNumber(v string) bool {
	if _, err1 := strconv.Atoi(v); err1 == nil {
		return true
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return true
	}
	return false
}

// Float64ToByte - convert float64 value to bytes.Buffer.
func Float64ToByte(f float64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, f)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

// CheckError - logs error.
func CheckError(err error) {
	if err != nil {
		log.Printf("error: %s", err)
	}
}
