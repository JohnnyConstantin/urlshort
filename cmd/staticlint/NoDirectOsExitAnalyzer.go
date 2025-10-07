package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NoDirectOsExitAnalyzer запрещает прямой вызов os.Exit в функции main пакета main.
//
// # Принцип работы
//
// 1. Идентифицирует функции с именем "main" в пакетах "main"
// 2. Проверяет, что функция не имеет параметров и возвращаемых значений
// 3. Обходит AST дерево функции в поиске вызовов os.Exit
// 4. Сообщает о найденных прямых вызовах
//
// В случае нахождения os.Exit, будет выдано предупреждение:
//   - direct call to os.Exit in main function of main package is forbidden
var NoDirectOsExitAnalyzer = &analysis.Analyzer{
	Name:     "nodirectosexit",
	Doc:      "forbid direct calls to os.Exit in main function of main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run является основной функцией выполнения анализатора.
// Она использует inspector для обхода AST и поиска запрещенных вызовов.
func run(pass *analysis.Pass) (interface{}, error) {
	inspectorLocal := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspectorLocal.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Проверка по ТЗ на функцию main в пакете main
		if funcDecl.Name.Name != "main" || pass.Pkg.Name() != "main" {
			return
		}

		// Проверяем, что у функции нет параметров и возвращаемых значений
		if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
			return
		}
		if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
			return
		}

		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkgIdent, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			// Проверяем, что идентификатор селектора равен os.Exit
			if pkgIdent.Name == "os" && selExpr.Sel.Name == "Exit" {
				pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function of main package is forbidden")
			}

			return true
		})
	})

	return nil, nil
}
