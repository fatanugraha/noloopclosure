package noloopclosure

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noloopclosure",
	Doc:      "noloopclosure is an analyzer that disallow reference capture of loop variable inside of a closure",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var errMsgFormat = "found reference to loop variable `%s`. Consider to duplicate variable `%s` before using it inside the function closure."

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	filter := []ast.Node{(*ast.ForStmt)(nil), (*ast.RangeStmt)(nil)}
	inspector.Preorder(filter, func(n ast.Node) {
		var loopVarObjs []types.Object
		var body *ast.BlockStmt
		switch nn := n.(type) {
		case *ast.RangeStmt:
			loopVarObjs = getRangeStmtLoopVars(pass, nn)
			body = nn.Body
		case *ast.ForStmt:
			loopVarObjs = getForStmtLoopVars(pass, nn)
			body = nn.Body
		}

		if loopVarObjs == nil {
			return
		}

		ast.Inspect(body, func(n ast.Node) bool {
			if flit, ok := n.(*ast.FuncLit); ok {
				issues := getIssues(pass, loopVarObjs, flit.Body)

				for _, v := range issues {
					pass.Reportf(v.pos, errMsgFormat, v.varName, v.varName)
				}
			}
			return true
		})
	})

	return nil, nil
}

type issue struct {
	pos     token.Pos
	varName string
}

func getRangeStmtLoopVars(pass *analysis.Pass, stmt *ast.RangeStmt) []types.Object {
	var loopVars []types.Object
	if val := stmt.Value; val != nil {
		loopVars = append(loopVars, pass.TypesInfo.ObjectOf(getIdent(val)))
	}
	if key := stmt.Key; key != nil {
		loopVars = append(loopVars, pass.TypesInfo.ObjectOf(getIdent(key)))
	}
	return loopVars

}

func getForStmtLoopVars(pass *analysis.Pass, stmt *ast.ForStmt) []types.Object {
	assignStmt, ok := stmt.Init.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	var loopVars []types.Object
	for _, lhs := range assignStmt.Lhs {
		loopVars = append(loopVars, pass.TypesInfo.ObjectOf(getIdent(lhs)))
	}

	return loopVars
}

func getIdent(expr ast.Expr) *ast.Ident {
	switch ee := expr.(type) {
	case *ast.Ident:
		return ee
	case *ast.SelectorExpr:
		return ee.Sel
	}
	panic("lol")
}

func getIssues(pass *analysis.Pass, loopVarObjs []types.Object, body *ast.BlockStmt) []issue {
	var issues []issue

	ast.Inspect(body, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		for _, loopVarObj := range loopVarObjs {
			if pass.TypesInfo.ObjectOf(ident) == loopVarObj {
				issues = append(issues, issue{pos: ident.Pos(), varName: ident.Name})
			}
		}

		return true
	})

	return issues
}
