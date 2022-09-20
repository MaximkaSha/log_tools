// Package notmain for testing of osexit.
package notmain

import (
	"fmt"
	"os"
)

func mulfunc(i int) (int, error) {
	return i * 2, nil
}

func errCheckFunc() {
	res, err := mulfunc(5)

	os.Exit(1)
	if err != nil {
		fmt.Println(res)
	}
	fmt.Println(res)
}
