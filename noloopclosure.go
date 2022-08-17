package noloopclosure

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noloopclosure",
	Doc:      "noloopclosure is an analyzer that disallow reference capture of loop variable inside of a closure",
	Run:      run,
	Flags:    flags(),
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var errMsgFormat = "found captured reference to loop variable inside a closure"

type analyzerState struct {
	pass *analysis.Pass

	loopVars []types.Object
	issues   map[string]token.Pos
}

func (state *analyzerState) processRangeStmt(stmt *ast.RangeStmt) {
	if val := stmt.Value; val != nil {
		state.markAsLoopVars(val)
	}

	if key := stmt.Key; key != nil {
		state.markAsLoopVars(key)
	}
}

func (state *analyzerState) processForStmt(stmt *ast.ForStmt) {
	if stmt.Init != nil {
		state.processForStmtClause(stmt.Init)
	}

	if stmt.Post != nil {
		state.processForStmtClause(stmt.Post)
	}
}

func (state *analyzerState) processForStmtClause(clause ast.Stmt) {
	switch stmt := clause.(type) {
	case *ast.AssignStmt:
		for _, lhs := range stmt.Lhs {
			state.markAsLoopVars(lhs)
		}
	case *ast.IncDecStmt:
		state.markAsLoopVars(stmt.X)
	}
}

func (state *analyzerState) markAsLoopVars(expr ast.Expr) {
	obj := state.pass.TypesInfo.ObjectOf(state.getIdent(expr))
	if obj == nil {
		return
	}

	for _, o := range state.loopVars {
		if obj == o {
			return
		}
	}

	state.loopVars = append(state.loopVars, obj)
}

func (state *analyzerState) getIdent(expr ast.Expr) *ast.Ident {
	switch ee := expr.(type) {
	case *ast.ParenExpr:
		return state.getIdent(ee.X)
	case *ast.Ident:
		return ee
	case *ast.SelectorExpr:
		return ee.Sel
	case *ast.IndexExpr:
		return state.getIdent(ee.X)
	case *ast.StarExpr:
		return state.getIdent(ee.X)
	default:
		//  Note: if you reach this state, please raise an issue, thanks!
		return nil
	}
}

func (state *analyzerState) processBody(body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		for _, loopVarObj := range state.loopVars {
			if state.pass.TypesInfo.ObjectOf(ident) == loopVarObj {
				state.markAsIssue(ident.Pos())
			}
		}

		return true
	})
}

func (state *analyzerState) markAsIssue(pos token.Pos) {
	if state.issues == nil {
		state.issues = map[string]token.Pos{}
	}

	pp := state.pass.Fset.Position(pos)
	state.issues[fmt.Sprintf("%s:%d", pp.Filename, pp.Line)] = pos
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

func run(pass *analysis.Pass) (interface{}, error) {
	// It's a common usecase to ignore tests by default as it's a common place to capture for-loop variables inside
	// a closure, especially for table-based tests and benchmarks.
	shouldIncludeTestFiles := pass.Analyzer.Flags.Lookup("t").Value.(flag.Getter).Get().(bool)

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	filter := []ast.Node{(*ast.ForStmt)(nil), (*ast.RangeStmt)(nil)}
	inspector.Preorder(filter, func(n ast.Node) {
		if !shouldIncludeTestFiles && isTestFile(pass.Fset.Position(n.Pos()).Filename) {
			return
		}

		state := analyzerState{pass: pass}

		var body *ast.BlockStmt
		switch nn := n.(type) {
		case *ast.RangeStmt:
			state.processRangeStmt(nn)
			body = nn.Body
		case *ast.ForStmt:
			state.processForStmt(nn)
			body = nn.Body
		}

		if state.loopVars == nil {
			return
		}

		ast.Inspect(body, func(n ast.Node) bool {
			if flit, ok := n.(*ast.FuncLit); ok {
				state.processBody(flit.Body)
			}
			return true
		})

		for _, v := range state.issues {
			pass.Reportf(v, errMsgFormat)
		}
	})

	return nil, nil
}

func flags() flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Bool("t", false, "Include checking test files")
	return *flags
}
