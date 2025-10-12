// Демонстрационная программа для тестирования символьных выражений
package main

import (
	"fmt"
	"log"
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"
)

func translateAndPrintRes(translator *translator.Z3Translator, expr symbolic.SymbolicExpression, name string) {
	z3Expr, _ := translator.TranslateExpression(expr)
	fmt.Printf("func %s: %s with type %T\n --------------------------------- \n", name, expr.String(), z3Expr)
}

func main() {
	// fmt.Println("=== Symbolic Expressions Demo ===")

	// Создаём простые символьные выражения
	x := symbolic.NewSymbolicVariable("x", symbolic.IntType)
	y := symbolic.NewSymbolicVariable("y", symbolic.IntType)
	five := symbolic.NewIntConstant(5)

	// Создаём выражение: x + y > 5
	sum := symbolic.NewBinaryOperation(x, y, symbolic.ADD)
	condition := symbolic.NewBinaryOperation(sum, five, symbolic.GT)

	fmt.Printf("Выражение: %s\n", condition.String())
	fmt.Printf("Тип выражения: %s\n", condition.Type().String())

	// Создаём Z3 транслятор
	translator := translator.NewZ3Translator()
	defer translator.Close()

	// Транслируем в Z3
	z3Expr, err := translator.TranslateExpression(condition)
	if err != nil {
		log.Fatalf("Ошибка трансляции: %v", err)
	}

	fmt.Printf("Z3 выражение создано: %T\n", z3Expr)
	// Создаём более сложное выражение: (x > 0) && (y < 10)
	zero := symbolic.NewIntConstant(0)
	ten := symbolic.NewIntConstant(10)

	cond1 := symbolic.NewBinaryOperation(x, zero, symbolic.GT)
	cond2 := symbolic.NewBinaryOperation(y, ten, symbolic.LT)

	andExpr := symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{cond1, cond2}, symbolic.AND)

	fmt.Printf("Сложное выражение: %s\n", andExpr.String())

	// Транслируем сложное выражение
	z3AndExpr, err := translator.TranslateExpression(andExpr)
	if err != nil {
		log.Fatalf("Ошибка трансляции сложного выражения: %v", err)
	}

	fmt.Printf("Сложное Z3 выражение создано: %T\n", z3AndExpr)

	// fmt.Println("Реализуйте методы в symbolic и translator пакетах для запуска демо!")
	fmt.Print("================ Translated from examples/ ================\n")

	//	func inRange(x, min, max int) bool {
	//		return x >= min && x <= max
	//	}
	x = symbolic.NewSymbolicVariable("x", symbolic.IntType)
	min := symbolic.NewIntConstant(-38)
	max := symbolic.NewIntConstant(123)
	cond1 = symbolic.NewBinaryOperation(x, min, symbolic.GE)
	cond2 = symbolic.NewBinaryOperation(x, max, symbolic.LE)
	andExpr = symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{cond1, cond2}, symbolic.AND)
	translateAndPrintRes(translator, andExpr, "inRange")

	// func add(a, b int) int {
	// 	return a + b
	// }
	a := symbolic.NewSymbolicVariable("a", symbolic.IntType)
	b := symbolic.NewSymbolicVariable("b", symbolic.IntType)
	add := symbolic.NewBinaryOperation(a, b, symbolic.ADD)
	translateAndPrintRes(translator, add, "add")

	//	func max(a, b int) int {
	//		if a > b {
	//			return a
	//		}
	//		return b
	//	}
	a = symbolic.NewSymbolicVariable("a", symbolic.IntType)
	b = symbolic.NewSymbolicVariable("b", symbolic.IntType)
	cond1 = symbolic.NewBinaryOperation(a, b, symbolic.GT)
	ifStmt := symbolic.NewTernaryOperation(cond1, a, b)
	translateAndPrintRes(translator, ifStmt, "max")

	//	func calculate(x, y int) int {
	//		sum := x + y
	//		diff := x - y
	//		product := sum * diff
	//		return product
	//	}
	x = symbolic.NewSymbolicVariable("x", symbolic.IntType)
	y = symbolic.NewSymbolicVariable("y", symbolic.IntType)
	sum = symbolic.NewBinaryOperation(x, y, symbolic.ADD)
	diff := symbolic.NewBinaryOperation(x, y, symbolic.SUB)
	product := symbolic.NewBinaryOperation(sum, diff, symbolic.MUL)
	translateAndPrintRes(translator, product, "calculate")

	//	func isValid(x, y int) bool {
	//		return x > 0 && y > 0 && x < 100
	//	}
	x = symbolic.NewSymbolicVariable("x", symbolic.IntType)
	y = symbolic.NewSymbolicVariable("y", symbolic.IntType)
	zr := symbolic.NewIntConstant(0)
	xGT0 := symbolic.NewBinaryOperation(x, zr, symbolic.GT)
	yGT0 := symbolic.NewBinaryOperation(y, zr, symbolic.GT)
	hun := symbolic.NewIntConstant(100)
	xLT100 := symbolic.NewBinaryOperation(x, hun, symbolic.LT)
	compound := symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{xGT0, yGT0}, symbolic.AND)
	compound2 := symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{compound, xLT100}, symbolic.AND)
	translateAndPrintRes(translator, compound2, "isValid")

	//	func unaryOps(x int, flag bool) int {
	//		result := -x
	//		if !flag {
	//			result = -result
	//		}
	//		return result
	//	}
	x = symbolic.NewSymbolicVariable("x", symbolic.IntType)
	flag := symbolic.NewSymbolicVariable("flag", symbolic.BoolType)
	result := symbolic.NewUnaryOperation(symbolic.UN_SUB, x)
	cond := symbolic.NewUnaryOperation(symbolic.UN_NOT, flag)
	result2 := symbolic.NewUnaryOperation(symbolic.UN_SUB, result)
	ret := symbolic.NewTernaryOperation(cond, result2, result)
	translateAndPrintRes(translator, ret, "unaryOps")

	//	func arrAcess(arr []int) int {
	//		return result[0]
	//	}
	arr := symbolic.NewSymbolicVariableArray("arr", symbolic.InnerType{ExprTy: symbolic.IntType})
	zr = symbolic.NewIntConstant(0)
	acc := symbolic.NewBinaryOperation(arr, zr, symbolic.SELECT)
	translateAndPrintRes(translator, acc, "arrAccess")

	//	func arrAcess2(arr [][]int) []int {
	//		return result[0]
	//	}
	arr = symbolic.NewSymbolicVariableArray("arr", symbolic.InnerType{ExprTy: symbolic.ArrayType, InnerTy: &symbolic.InnerType{ExprTy: symbolic.IntType}})
	zr = symbolic.NewIntConstant(0)
	acc = symbolic.NewBinaryOperation(arr, zr, symbolic.SELECT)
	translateAndPrintRes(translator, acc, "arrAccess2")

	//	func max(x, y int) int {
	//		if x > y {
	//			return x
	//		}
	//		return y
	//	}
	x = symbolic.NewSymbolicVariable("x", symbolic.IntType)
	y = symbolic.NewSymbolicVariable("y", symbolic.IntType)
	cond1 = symbolic.NewBinaryOperation(x, y, symbolic.GT)
	ifStmt = symbolic.NewTernaryOperation(cond1, x, y)
	//	func functionCall(x, y int) int {
	//		return max(x, y)
	//	}
	var argsTypes []symbolic.InnerType = []symbolic.InnerType{
		{ExprTy: symbolic.IntType},
		{ExprTy: symbolic.IntType},
	}
	foo := symbolic.NewFunction("max", argsTypes, symbolic.InnerType{ExprTy: symbolic.IntType})
	call := symbolic.NewFunctionCall(*foo, []symbolic.SymbolicExpression{x, y})
	binOp := symbolic.NewBinaryOperation(call, ifStmt, symbolic.EQ)
	translateAndPrintRes(translator, binOp, "functionDecl")
}
