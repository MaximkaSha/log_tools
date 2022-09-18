package osexitcheck

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
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
						pass.Reportf(x.Fun.Pos(), "found os.Exit in main()")
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
	return strings.Contains(src, sub)
}
