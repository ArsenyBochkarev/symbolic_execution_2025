package internal

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"strconv"

	"github.com/go-toolsmith/astcopy"
	"golang.org/x/tools/go/ast/astutil"
)

// TODO: Other relation operators
// TODO: Other operations for induction variable (`+=`, `-=`, etc)
// TODO: Non-invariant end values
func analyzeForStmt(forStmt ast.ForStmt) (string, int, int, int) {
	var indVarName string
	var startValue, endValue int

	if assignStmt, ok := forStmt.Init.(*ast.AssignStmt); ok {
		if len(assignStmt.Lhs) == 1 {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
				indVarName = ident.Name
			}
		} else {
			panic("unsupported assign stmt for induction var")
		}
		if len(assignStmt.Rhs) == 1 {
			if basicLit, ok := assignStmt.Rhs[0].(*ast.BasicLit); ok {
				if basicLit.Kind == token.INT {
					startValue, _ = strconv.Atoi(basicLit.Value)
				}
			}
		} else {
			panic("non-invariant RHS for induction var in assignment")
		}
	} else {
		panic("unsupported (non-assign) stmt for induction var")
	}

	var isIncluding bool
	if binaryExpr, ok := forStmt.Cond.(*ast.BinaryExpr); ok {
		switch binaryExpr.Op {
		case token.LSS: // "<"
			isIncluding = true
		case token.LEQ: // "<="
			isIncluding = false
		default:
			panic("unsupported relation operator")
		}

		// RHS
		if basicLit, ok := binaryExpr.Y.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
			endValue, _ = strconv.Atoi(basicLit.Value)
		} else {
			panic("non-invariant RHS in loop condition")
		}
	} else {
		panic("non-binary condition expr")
	}

	var iterationCount int
	if isIncluding {
		iterationCount = endValue - startValue
	} else {
		iterationCount = endValue - startValue + 1
	}

	var stepValue int
	if incStmt, ok := forStmt.Post.(*ast.IncDecStmt); ok {
		switch incStmt.Tok {
		case token.INC:
			stepValue = 1
		case token.DEC:
			stepValue = -1
		default:
			panic("unsupported operation for induction variable step")
		}
	}

	return indVarName, startValue, iterationCount, stepValue
}

func applyForCursor(cursor *astutil.Cursor) bool {
	if forStmt, ok := cursor.Node().(*ast.ForStmt); ok {
		indVarName, startValue, iterationCount, stepValue := analyzeForStmt(*forStmt)

		var newStatements []ast.Stmt
		for i := range iterationCount {
			currentValue := startValue + i*stepValue
			for _, stmt := range forStmt.Body.List {
				copiedStmt := astcopy.Stmt(stmt)
				newStmt := replaceIdentWithValue(copiedStmt, indVarName, currentValue)
				newStatements = append(newStatements, newStmt)
			}
		}

		newBlock := &ast.BlockStmt{
			Lbrace: forStmt.Body.Lbrace,
			Rbrace: forStmt.Body.Rbrace,
			List:   newStatements,
		}
		cursor.Replace(newBlock)
	}
	return true
}

// Replace all varName in node with literal value
func replaceIdentWithValue(node ast.Stmt, varName string, value int) ast.Stmt {
	return astutil.Apply(
		node,
		func(cursor *astutil.Cursor) bool {
			if ident, ok := cursor.Node().(*ast.Ident); ok && ident.Name == varName {
				newLit := &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(value)}
				cursor.Replace(newLit)
			}
			return true
		},
		nil).(ast.Stmt)
}

// Tries to unroll loop on AST level
func tryUnrollLoops(source string) string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		panic(err)
	}

	newNode := astutil.Apply(node, applyForCursor, nil)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, newNode); err != nil {
		log.Fatal(err)
	}
	return buf.String()
}
