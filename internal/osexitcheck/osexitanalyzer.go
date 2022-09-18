package osexitcheck

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var OSExitAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for os.Exit in main.go",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		if findSubString(file.Name.String(), "main") {
			ast.Inspect(file, func(node ast.Node) bool {
				if x, ok := node.(*ast.CallExpr); ok {
					str := fmt.Sprintf("%v", pass.Fset.Position(x.Fun.Pos()))
					var buf bytes.Buffer
					printer.Fprint(&buf, pass.Fset, x.Fun)
					if findOSExitInMain(buf.String() + str) {
						fmt.Printf("Found os.Exit in main() %v: ", pass.Fset.Position(x.Fun.Pos()))
						printer.Fprint(os.Stdout, pass.Fset, x.Fun)
						fmt.Println()
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

func findOSExitInMain(src string) bool {
	if findSubString(src, "\\main.go") || findSubString(src, "/main.go") {
		if findSubString(src, "os.Exit") {
			return true
		}
	}
	return false
}

func findSubString(src string, sub string) bool {
	if strings.Contains(src, sub) {
		return true
	}
	return false
}
