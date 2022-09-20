// Package main for testing osexitanalyzer.
package main

import (
	"fmt"
	"os"
)

func mulfunc(i int) (int, error) {
	return i * 2, nil
}

func main() {
	res, _ := mulfunc(5)
	os.Exit(1) // want "found os.Exit in main()"
	fmt.Println(res)
}
