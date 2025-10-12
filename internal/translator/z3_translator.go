// Package translator содержит реализацию транслятора в Z3
package translator

import (
	"math/big"
	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// Z3Translator транслирует символьные выражения в Z3 формулы
type Z3Translator struct {
	ctx    *z3.Context
	config *z3.Config
	vars   map[string]z3.Value // Кэш переменных
}

// NewZ3Translator создаёт новый экземпляр Z3 транслятора
func NewZ3Translator() *Z3Translator {
	config := &z3.Config{}
	ctx := z3.NewContext(config)

	return &Z3Translator{
		ctx:    ctx,
		config: config,
		vars:   make(map[string]z3.Value),
	}
}

// GetContext возвращает Z3 контекст
func (zt *Z3Translator) GetContext() interface{} {
	return zt.ctx
}

// Reset сбрасывает состояние транслятора
func (zt *Z3Translator) Reset() {
	zt.vars = make(map[string]z3.Value)
}

// Close освобождает ресурсы
func (zt *Z3Translator) Close() {
	// Z3 контекст закрывается автоматически
}

// TranslateExpression транслирует символьное выражение в Z3
func (zt *Z3Translator) TranslateExpression(expr symbolic.SymbolicExpression) (interface{}, error) {
	return expr.Accept(zt), nil
}

// VisitVariable транслирует символьную переменную в Z3
func (zt *Z3Translator) VisitVariable(expr *symbolic.SymbolicVariable) interface{} {
	// Проверить, есть ли переменная в кэше
	// Если нет - создать новую Z3 переменную соответствующего типа
	// Добавить в кэш и вернуть

	// Подсказки:
	// - Используйте zt.ctx.IntConst(name) для int переменных
	// - Используйте zt.ctx.BoolConst(name) для bool переменных
	// - Храните переменные в zt.vars для повторного использования

	if v, ok := zt.vars[expr.Name]; ok {
		return v
	}

	var z z3.Value
	switch expr.Type() {
	case symbolic.IntType:
		z = zt.ctx.IntConst(expr.Name)
	case symbolic.BoolType:
		z = zt.ctx.BoolConst(expr.Name)
	case symbolic.ArrayType:
		as := zt.ctx.ArraySort(zt.ctx.IntSort(), symbolic.Type2Sort(zt.ctx, &expr.InnerType))
		z = zt.ctx.Const(expr.Name, as)

	default:
		panic("unsupported variable type")
	}

	zt.vars[expr.Name] = z
	return z
}

// VisitIntConstant транслирует целочисленную константу в Z3
func (zt *Z3Translator) VisitIntConstant(expr *symbolic.IntConstant) interface{} {
	// Создать Z3 константу с помощью zt.ctx.FromBigInt или аналогичного метода
	bigint := big.NewInt(expr.Value)
	return zt.ctx.FromBigInt(bigint, zt.ctx.IntSort())
}

// VisitBoolConstant транслирует булеву константу в Z3
func (zt *Z3Translator) VisitBoolConstant(expr *symbolic.BoolConstant) interface{} {
	// Использовать zt.ctx.FromBool для создания Z3 булевой константы
	return zt.ctx.FromBool(expr.Value)
}

// VisitBinaryOperation транслирует бинарную операцию в Z3
func (zt *Z3Translator) VisitBinaryOperation(expr *symbolic.BinaryOperation) interface{} {
	// 1. Транслировать левый и правый операнды
	// 2. В зависимости от оператора создать соответствующую Z3 операцию
	// Подсказки по операциям в Z3:
	// - Арифметические: left.Add(right), left.Sub(right), left.Mul(right), left.Div(right)
	// - Сравнения: left.Eq(right), left.LT(right), left.LE(right), etc.
	// - Приводите типы: left.(z3.Int), right.(z3.Int) для int операций

	leftOp := expr.Left.Accept(zt)
	rightOp := expr.Right.Accept(zt)

	switch expr.Operator {
	// Arithmetic binary operations
	case symbolic.MUL:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).Mul(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.ADD:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).Add(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.SUB:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).Sub(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.BinaryOperator(symbolic.DIV):
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).Div(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.MOD:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).Mod(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}

	// Comparison binary operations
	case symbolic.EQ:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).Eq(rightOp.(z3.Int))
		case symbolic.BoolType:
			return leftOp.(z3.Bool).Eq(rightOp.(z3.Bool))
		case symbolic.ArrayType:
			return leftOp.(z3.Array).Eq(rightOp.(z3.Array))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.NE:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).NE(rightOp.(z3.Int))
		case symbolic.BoolType:
			return leftOp.(z3.Bool).NE(rightOp.(z3.Bool))
		case symbolic.ArrayType:
			return leftOp.(z3.Array).NE(rightOp.(z3.Array))
		}
	case symbolic.GE:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).GE(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.GT:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).GT(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.LE:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).LE(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}
	case symbolic.LT:
		switch expr.Left.Type() {
		case symbolic.IntType:
			return leftOp.(z3.Int).LT(rightOp.(z3.Int))
		default:
			panic("unknown type in VisitBinaryOperation")
		}

	case symbolic.SELECT:
		switch expr.Left.Type() {
		case symbolic.ArrayType:
			return leftOp.(z3.Array).Select(rightOp.(z3.Value))
		default:
			panic("unknown type in VisitBinaryOperation")
		}

	default:
		panic("unknown binary operation")
	}
	panic("unreachable")
}

// VisitLogicalOperation транслирует логическую операцию в Z3
func (zt *Z3Translator) VisitLogicalOperation(expr *symbolic.LogicalOperation) interface{} {
	// 1. Транслировать все операнды
	// 2. Применить соответствующую логическую операцию

	// Подсказки:
	// - AND: zt.ctx.And(operands...)
	// - OR: zt.ctx.Or(operands...)
	// - NOT: operand.Not() (для единственного операнда)
	// - IMPLIES: antecedent.Implies(consequent)
	var translatedOperands []z3.Value
	for _, op := range expr.Operands {
		z := op.Accept(zt)
		translatedOperands = append(translatedOperands, z.(z3.Value))
	}

	switch expr.Operator {
	case symbolic.AND:
		res := translatedOperands[0]
		for i := 1; i < len(translatedOperands); i++ {
			res = res.(z3.Bool).And(translatedOperands[i].(z3.Bool))
		}
		return res
	case symbolic.OR:
		res := translatedOperands[0]
		for i := 1; i < len(translatedOperands); i++ {
			res = res.(z3.Bool).Or(translatedOperands[i].(z3.Bool))
		}
		return res
	case symbolic.NOT:
		return translatedOperands[0].(z3.Bool).Not()
	case symbolic.IMPLIES:
		return translatedOperands[0].(z3.Bool).Implies(translatedOperands[1].(z3.Bool))

	default:
		panic("unknown logical operator")
	}
}

func (zt *Z3Translator) VisitTernaryOperation(expr *symbolic.TernaryOperation) interface{} {
	translatedOp := expr.Condition.Accept(zt)
	trueOp := expr.TrueExpr.Accept(zt)
	falseOp := expr.FalseExpr.Accept(zt)

	return translatedOp.(z3.Bool).IfThenElse(trueOp.(z3.Value), falseOp.(z3.Value))
}

func (zt *Z3Translator) VisitUnaryOperation(expr *symbolic.UnaryOperation) interface{} {
	switch expr.Operator {
	case symbolic.UN_NOT:
		translatedExpr := expr.Expr.Accept(zt)
		return translatedExpr.(z3.Bool).Not()
	case symbolic.UN_SUB:
		translatedExpr := expr.Expr.Accept(zt)
		return translatedExpr.(z3.Int).Neg()
	default:
		panic("unknown unary operator")
	}
}

func (zt *Z3Translator) VisitFunction(expr *symbolic.Function) interface{} {
	var argsSorts []z3.Sort
	for i := range expr.Args {
		argTy := expr.Args[i]
		argsSorts = append(argsSorts, symbolic.Type2Sort(zt.ctx, &argTy))
	}
	return zt.ctx.FuncDecl(expr.Name, argsSorts, symbolic.Type2Sort(zt.ctx, &expr.RetType))
}

func (zt *Z3Translator) VisitFunctionCall(expr *symbolic.FunctionCall) interface{} {
	decl := zt.VisitFunction(&expr.FunctionDecl)
	var args []z3.Value
	for i := range expr.Args {
		arg := expr.Args[i]
		translatedArg := arg.Accept(zt)
		args = append(args, translatedArg.(z3.Value))
	}
	return decl.(z3.FuncDecl).Apply(args...)
}

// Вспомогательные методы

// createZ3Variable создаёт Z3 переменную соответствующего типа
func (zt *Z3Translator) createZ3Variable(name string, exprType symbolic.ExpressionType) z3.Value {
	// Создать Z3 переменную на основе типа
	panic("не реализовано")
}

// castToZ3Type приводит значение к нужному Z3 типу
func (zt *Z3Translator) castToZ3Type(value interface{}, targetType symbolic.ExpressionType) (z3.Value, error) {
	// Безопасно привести interface{} к конкретному Z3 типу
	panic("не реализовано")
}
