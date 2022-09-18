// Staticlint package is a static code analyzer.
// It utilaze all go/analysis, go-critic, ineffassign linters and osexit analyzer.
// To install, run.
// $ go get github.com/log_tools/internal/osexitcheck.
// and put the resulting binary in one of your PATH directories if $GOPATH/bin isn't already in your PATH.
// Usage:
// staticlint [<flag> ...] <Go file or directory> ...
// Examples:
// staticlint ../...
// staticlint main.go.

package main

import (
	"github.com/MaximkaSha/log_tools/internal/osexitcheck"
	"github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Describes staticcheck rules.
	checks := map[string]bool{
		"SA":     true, // SA rules
		"S1002":  true, // Omit comparison with boolean constant
		"ST1017": true, // Yoda conditions
		"QF1001": true, // De Morgan’s law
	}
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
	mychecks = append(mychecks, printf.Analyzer)
	// Package structtag defines an Analyzer that checks struct field tags are well formed.
	mychecks = append(mychecks, structtag.Analyzer)
	// Package asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
	mychecks = append(mychecks, asmdecl.Analyzer)
	// Package assign defines an Analyzer that detects useless assignments.
	mychecks = append(mychecks, assign.Analyzer)
	// Package atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
	mychecks = append(mychecks, atomic.Analyzer)
	// Package atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
	mychecks = append(mychecks, atomicalign.Analyzer)
	// Package bools defines an Analyzer that detects common mistakes involving boolean operators.
	mychecks = append(mychecks, bools.Analyzer)
	// Package buildssa defines an Analyzer that constructs the SSA representation of an error-free package and returns the set of all functions within it.
	mychecks = append(mychecks, buildssa.Analyzer)
	// Package buildtag defines an Analyzer that checks build tags.
	mychecks = append(mychecks, buildtag.Analyzer)
	// Package cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
	mychecks = append(mychecks, cgocall.Analyzer)
	// Package composite defines an Analyzer that checks for unkeyed composite literals.
	mychecks = append(mychecks, composite.Analyzer)
	// Package copylock defines an Analyzer that checks for locks erroneously passed by value.
	mychecks = append(mychecks, copylock.Analyzer)
	// Package ctrlflow is an analysis that provides a syntactic control-flow graph (CFG) for the body of a function.
	mychecks = append(mychecks, ctrlflow.Analyzer)
	// Package deepequalerrors defines an Analyzer that checks for the use of reflect.DeepEqual with error values.
	mychecks = append(mychecks, deepequalerrors.Analyzer)
	// The errorsas package defines an Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
	mychecks = append(mychecks, errorsas.Analyzer)
	// Package fieldalignment defines an Analyzer that detects structs that would use less memory if their fields were sorted.
	mychecks = append(mychecks, fieldalignment.Analyzer)
	// Package findcall defines an Analyzer that serves as a trivial example and test of the Analysis API.
	mychecks = append(mychecks, findcall.Analyzer)
	// Package framepointer defines an Analyzer that reports assembly code that clobbers the frame pointer before saving it.
	mychecks = append(mychecks, framepointer.Analyzer)
	// Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
	mychecks = append(mychecks, httpresponse.Analyzer)
	// Package ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
	mychecks = append(mychecks, ifaceassert.Analyzer)
	// Package inspect defines an Analyzer that provides an AST inspector (golang.org/x/tools/go/ast/inspector.Inspector) for the syntax trees of a package.
	mychecks = append(mychecks, inspect.Analyzer)
	// Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
	mychecks = append(mychecks, loopclosure.Analyzer)
	// Package lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
	mychecks = append(mychecks, lostcancel.Analyzer)
	// Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
	mychecks = append(mychecks, nilfunc.Analyzer)
	// Package nilness inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons.
	mychecks = append(mychecks, nilness.Analyzer)
	// The pkgfact package is a demonstration and test of the package fact mechanism.
	mychecks = append(mychecks, pkgfact.Analyzer)
	// Package reflectvaluecompare defines an Analyzer that checks for accidentally using == or reflect.DeepEqual to compare reflect.Value values.
	mychecks = append(mychecks, reflectvaluecompare.Analyzer)
	// Package shadow defines an Analyzer that checks for shadowed variables.
	mychecks = append(mychecks, shadow.Analyzer)
	// Package sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal.Notify.
	mychecks = append(mychecks, sigchanyzer.Analyzer)
	// Package sortslice defines an Analyzer that checks for calls to sort.Slice that do not use a slice type as first argument.
	mychecks = append(mychecks, sortslice.Analyzer)
	// Package stdmethods defines an Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
	mychecks = append(mychecks, stdmethods.Analyzer)
	// Package stringintconv defines an Analyzer that flags type conversions from integers to strings.
	mychecks = append(mychecks, stringintconv.Analyzer)
	// Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
	mychecks = append(mychecks, tests.Analyzer)
	// The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
	mychecks = append(mychecks, unmarshal.Analyzer)
	// Package unreachable defines an Analyzer that checks for unreachable code.
	mychecks = append(mychecks, unreachable.Analyzer)
	// Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
	mychecks = append(mychecks, unsafeptr.Analyzer)
	// Package unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
	mychecks = append(mychecks, unusedresult.Analyzer)
	// Package unusedwrite checks for unused writes to the elements of a struct or array object.
	mychecks = append(mychecks, unusedwrite.Analyzer)
	// Package usesgenerics defines an Analyzer that checks for usage of generic features added in Go 1.18.
	mychecks = append(mychecks, usesgenerics.Analyzer)
	// Package osexitcheck cheks for using os.Exit direct from main func main package.
	mychecks = append(mychecks, osexitcheck.OSExitAnalyzer)
	// Package go-critic 100 diagnostics that check for bugs, performance and style issues.
	mychecks = append(mychecks, analyzer.Analyzer)
	// Detect ineffectual assignments in Go code. An assignment is ineffectual if the variable assigned is not thereafter used.
	mychecks = append(mychecks, ineffassign.Analyzer)
	multichecker.Main(
		mychecks...,
	)

}
