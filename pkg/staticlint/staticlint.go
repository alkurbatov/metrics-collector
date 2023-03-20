// Package staticlint implements multichecker static analyser with different linters.
// The packages uses golang.org/x/tools/go/analysis/multichecker under the hood
// to combine different linters in unified and convenient utility.
//
// Staticlint includes the following linters:
//
// From golang.org/x/tools/go/analysis/passes:
// assign, atomic, atomicalign, bools, composite, copylock, deepequalerrors,
// directive, errorsas, httpresponse, ifaceassert, loopclosure, lostcancel,
// nilfunc, nilness, reflectvaluecompare, shadow, shift, sigchanyzer, sortslice,
// stdmethods, stringintconv, structtag, tests, timeformat, unmarshal,
// unreachable, unsafeptr, unusedresult, unusedwrite.
//
// For additional details regarding particular linter please refer to:
// https://pkg.go.dev/golang.org/x/tools/go/analysis/passes
//
// From staticheck.io:
//   - All SA* (staticcheck) linters.
//   - All S* (simple) linters.
//   - All ST* (stylecheck) linters.
//   - All QF* (quickfix) linters.
//
// For additional details regarding particular linting group please refer to:
// https://staticcheck.io/docs/checks/
//
// Other publicly available linters:
//   - errcheck to check for unchecked errors in Go code, see
//     https://github.com/kisielk/errcheck
//   - bodyclose to check whether HTTP response body is closed and
//     a re-use of TCP connection is not blocked, see:
//     https://github.com/timakin/bodyclose
//   - testpackage that makes you use a separate _test package, see
//     https://github.com/maratori/testpackage
//
// Custom linters:
//   - noexit to check whether os.Exit is used in the main function of the main package.
package staticlint

import (
	"github.com/alkurbatov/metrics-collector/pkg/noexit"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/maratori/testpackage/pkg/testpackage"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// StaticLint is umbrella structure for collection of linters passed to
// golang.org/x/tools/go/analysis/multichecker.
type StaticLint struct {
	checkers []*analysis.Analyzer
}

// New initializes new instance of StaticLint linter.
func New() StaticLint {
	// Add analyzers from passes.
	checkers := []*analysis.Analyzer{
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
	}

	// Add staticcheck analyzers.
	for _, collection := range [][]*lint.Analyzer{
		staticcheck.Analyzers,
		simple.Analyzers,
		stylecheck.Analyzers,
		quickfix.Analyzers,
	} {
		for _, v := range collection {
			checkers = append(checkers, v.Analyzer)
		}
	}

	// Add more standalone analyzers.
	checkers = append(checkers, errcheck.Analyzer)
	checkers = append(checkers, bodyclose.Analyzer)
	checkers = append(checkers, testpackage.NewAnalyzer())

	// Add custom linter.
	checkers = append(checkers, noexit.Analyzer)

	return StaticLint{checkers}
}

// Run launches the linters against provided source code.
func (s StaticLint) Run() {
	multichecker.Main(s.checkers...)
}
