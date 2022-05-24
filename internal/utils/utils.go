package utils

import (
	"strconv"
)

func CheckIfStringIsNumber(v string) bool {
	if _, err := strconv.Atoi(v); err == nil {
		return true
	}
	return false
}
