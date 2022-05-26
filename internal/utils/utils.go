package utils

import (
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
