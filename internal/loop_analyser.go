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

// Loop in form:
// for i := a; i rel b; op, where
// - a should be const
// - rel can be `<`, `>`, `<=`, `>=`
// - b can be non-invariant
// - op can be `++`, `--`, `+=`, `-=`
func analyzeForStmt(forStmt ast.ForStmt, basicUnrollValue int) (string, int, int, int, bool) {
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
	var direction int // 1 -- incr, -1 -- decr
	if binaryExpr, ok := forStmt.Cond.(*ast.BinaryExpr); ok {
		switch binaryExpr.Op {
		case token.LSS: // "<"
			isIncluding = true
			direction = 1
		case token.LEQ: // "<="
			isIncluding = false
			direction = 1
		case token.GTR: // ">"
			isIncluding = true
			direction = -1
		case token.GEQ: // ">="
			isIncluding = false
			direction = -1
		case token.EQL: // "=="
			panic("`==` operator in for loop condition is not supported for unrolling")
		case token.NEQ: // "!="
			panic("`!=` operator in for loop condition is not supported for unrolling")
		default:
			panic("unsupported relation operator")
		}

		// RHS
		if basicLit, ok := binaryExpr.Y.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
			endValue, _ = strconv.Atoi(basicLit.Value)
		} else if _, ok := binaryExpr.Y.(*ast.Ident); ok {
			endValue = basicUnrollValue
		} else {
			panic("unsupported RHS in loop condition")
		}
	} else {
		panic("non-binary condition expr")
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
	} else if assignStmt, ok := forStmt.Post.(*ast.AssignStmt); ok {
		// +=, -=
		if len(assignStmt.Lhs) == 1 {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok && ident.Name == indVarName {
				if len(assignStmt.Rhs) == 1 {
					switch assignStmt.Tok {
					case token.ADD_ASSIGN:
						if e, ok := assignStmt.Rhs[0].(*ast.BasicLit); ok {
							stepValue, _ = strconv.Atoi(e.Value)
						} else {
							panic("non-invariant steps unsupported")
						}
					case token.SUB_ASSIGN:
						if e, ok := assignStmt.Rhs[0].(*ast.BasicLit); ok {
							stepValue, _ = strconv.Atoi(e.Value)
							stepValue = (-1) * stepValue
						} else {
							panic("non-invariant steps unsupported")
						}
					default:
						panic("unsupported operation")
					}
				}
			} else {
				panic("post statement modifies different variable")
			}
		}
	} else if forStmt.Post != nil {
		panic("unsupported post statement type")
	}

	var iterationCount int
	if direction > 0 {
		if isIncluding {
			iterationCount = (endValue - startValue + stepValue) / stepValue
		} else {
			iterationCount = (endValue - startValue) / stepValue
		}
	} else {
		if isIncluding {
			iterationCount = (startValue - endValue + (-stepValue)) / (-stepValue)
		} else {
			iterationCount = (startValue - endValue) / (-stepValue)
		}
	}
	if iterationCount < 0 {
		iterationCount = 0
	}

	return indVarName, startValue, iterationCount, stepValue, direction > 0
}

func applyForCursor(cursor *astutil.Cursor) bool {
	if forStmt, ok := cursor.Node().(*ast.ForStmt); ok {
		indVarName, startValue, iterationCount, stepValue, increasing := analyzeForStmt(*forStmt, 10 /*=basicUnrollValue*/)

		var newStatements []ast.Stmt
		for i := range iterationCount {
			var currentValue int
			if increasing {
				currentValue = startValue + i*stepValue
			} else {
				currentValue = startValue - i*(-stepValue)
			}

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
