// Package noexit implement static analyzer which checks for use of 'os.Exit'
// in the main function of main package.
package noexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "forbids usage of os.Exit in the main function of main package",
	Run:  run,
}

// isMainFunc returns true if the specified node declares 'func main'.
func isMainFunc(node *ast.FuncDecl) bool {
	if node.Name == nil || node.Name.Obj == nil {
		return false
	}

	decl := node.Name.Obj

	return decl.Kind == ast.Fun && decl.Name == "main"
}

func validateCallExpr(node ast.Node, pass *analysis.Pass) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	var pkgName string
	if pkg, ok := selector.X.(*ast.Ident); ok {
		pkgName = pkg.Name
	}

	if pkgName == "os" && selector.Sel.Name == "Exit" {
		pass.Reportf(node.Pos(), "os.Exit in main function is forbidden")
		return
	}
}

func validateExprStmt(node *ast.ExprStmt, pass *analysis.Pass) {
	validateCallExpr(node.X, pass)
}

func validateGoStmt(node *ast.GoStmt, pass *analysis.Pass) {
	if node.Call != nil {
		validateCallExpr(node.Call, pass)
	}
}

func validateDeferStmt(node *ast.DeferStmt, pass *analysis.Pass) {
	if node.Call != nil {
		validateCallExpr(node.Call, pass)
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// NB (alkurbatov): Ignore files not ended with the .go suffix, otherwise we will parse files like:
		// ~/Library/Caches/go-build/48/48d9cb68dc485871876a9702c960476cf6c7ca2a69e155317a78269c73427141-d:51:2
		filename := pass.Fset.Position(file.Pos()).Filename
		if strings.HasSuffix(filename, "_test.go") || !strings.HasSuffix(filename, ".go") {
			continue
		}

		if file.Name != nil && file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				return isMainFunc(x)

			case *ast.ExprStmt:
				validateExprStmt(x, pass)

			case *ast.GoStmt:
				validateGoStmt(x, pass)

			case *ast.DeferStmt:
				validateDeferStmt(x, pass)

			case *ast.FuncLit:
				return false
			}

			return true
		})
	}

	return nil, nil //nolint: nilnil
}
