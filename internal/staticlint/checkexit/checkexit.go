// Package checkexit  - statict checker for os.Exit calls in main package in main function
package checkexit

import (
	"go/ast"

	"go/token"

	"golang.org/x/tools/go/analysis"
)

type checkExit struct {
	tokenErr []token.Position
	mainPak  bool
	mainFun  bool
	mainErr  bool
}

var ErrCheckExit = &analysis.Analyzer{
	Name: "checkExit",
	Doc:  "check for os.Exit calls",
	Run:  run,
}

var testExit checkExit

func run(pass *analysis.Pass) (interface{}, error) {
	exprPkg := func(x *ast.File) {
		if x.Name.Name == "main" {
			testExit.mainPak = true
		} else {
			testExit.mainPak = false
		}
	}

	exprFuncMain := func(x *ast.FuncDecl) {
		if x.Name.Name == "main" {
			testExit.mainFun = true
		} else {
			testExit.mainFun = false
		}
	}

	exprCallExpr := func(x *ast.CallExpr) {
		isExit := isPkgDot(x.Fun, "os", "Exit")
		if isExit && testExit.mainPak && testExit.mainFun {
			testExit.mainErr = true
			testExit.tokenErr = append(testExit.tokenErr, pass.Fset.Position(x.Fun.Pos()))
			pass.Reportf(x.Pos(), "call os.Exit in main pkg in main function")
		}
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				exprPkg(x)
			case *ast.FuncDecl:
				exprFuncMain(x)
			case *ast.CallExpr:
				exprCallExpr(x)
			}
			return true
		})
	}
	return nil, nil
}

func isPkgDot(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && isIdent(sel.X, pkg) && isIdent(sel.Sel, name)
}

func isIdent(expr ast.Expr, ident string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == ident
}
