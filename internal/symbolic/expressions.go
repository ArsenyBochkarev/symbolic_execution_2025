// Package symbolic содержит конкретные реализации символьных выражений
package symbolic

import (
	"fmt"
	"strconv"

	"github.com/ebukreev/go-z3/z3"
)

// SymbolicExpression - базовый интерфейс для всех символьных выражений
type SymbolicExpression interface {
	// Type возвращает тип выражения
	Type() ExpressionType

	// String возвращает строковое представление выражения
	String() string

	// Accept принимает visitor для обхода дерева выражений
	Accept(visitor Visitor) interface{}
}

// SymbolicVariable представляет символьную переменную
type SymbolicVariable struct {
	Name       string
	ExprType   ExpressionType
	InnerType  InnerType   // parameter for array
	FieldTypes []InnerType // parameter for objects
}

// NewSymbolicVariable создаёт новую символьную переменную
func NewSymbolicVariable(name string, exprType ExpressionType) *SymbolicVariable {
	return &SymbolicVariable{
		Name:     name,
		ExprType: exprType,
	}
}

// NewSymbolicVariableArray создаёт новую символьную переменную с типом элементов
func NewSymbolicVariableArray(name string, innerTy InnerType) *SymbolicVariable {
	return &SymbolicVariable{
		Name:      name,
		ExprType:  ArrayType,
		InnerType: innerTy,
	}
}

// NewSymbolicVariableObject создаёт новую символьную переменную с массивом типов полей
func NewSymbolicVariableObject(name string) *SymbolicVariable {
	return &SymbolicVariable{
		Name:     name,
		ExprType: ObjectType,
	}
}

// Type возвращает тип переменной
func (sv *SymbolicVariable) Type() ExpressionType {
	return sv.ExprType
}

// String возвращает строковое представление переменной
func (sv *SymbolicVariable) String() string {
	return sv.Name
}

// Accept реализует Visitor pattern
func (sv *SymbolicVariable) Accept(visitor Visitor) interface{} {
	return visitor.VisitVariable(sv)
}

// IntConstant представляет целочисленную константу
type IntConstant struct {
	Value int64
}

// NewIntConstant создаёт новую целочисленную константу
func NewIntConstant(value int64) *IntConstant {
	return &IntConstant{Value: value}
}

// Type возвращает тип константы
func (ic *IntConstant) Type() ExpressionType {
	return IntType
}

// String возвращает строковое представление константы
func (ic *IntConstant) String() string {
	return fmt.Sprintf("%d", ic.Value)
}

// Accept реализует Visitor pattern
func (ic *IntConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitIntConstant(ic)
}

// BoolConstant представляет булеву константу
type BoolConstant struct {
	Value bool
}

// NewBoolConstant создаёт новую булеву константу
func NewBoolConstant(value bool) *BoolConstant {
	return &BoolConstant{Value: value}
}

// Type возвращает тип константы
func (bc *BoolConstant) Type() ExpressionType {
	return BoolType
}

// String возвращает строковое представление константы
func (bc *BoolConstant) String() string {
	return fmt.Sprintf("%t", bc.Value)
}

// Accept реализует Visitor pattern
func (bc *BoolConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitBoolConstant(bc)
}

// FloatConstant представляет целочисленную константу
type FloatConstant struct {
	Value float64
}

// NewFloatConstant создаёт новую целочисленную константу
func NewFloatConstant(value float64) *FloatConstant {
	return &FloatConstant{Value: value}
}

// Type возвращает тип константы
func (fc *FloatConstant) Type() ExpressionType {
	return FloatType
}

// String возвращает строковое представление константы
func (fc *FloatConstant) String() string {
	return fmt.Sprintf("%f", fc.Value)
}

// Accept реализует Visitor pattern
func (fc *FloatConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitFloatConstant(fc)
}

// BinaryOperation представляет бинарную операцию
type BinaryOperation struct {
	Left     SymbolicExpression
	Right    SymbolicExpression
	Operator BinaryOperator
}

// NewBinaryOperation создаёт новую бинарную операцию
func NewBinaryOperation(left, right SymbolicExpression, op BinaryOperator) *BinaryOperation {
	// Создать новую бинарную операцию и проверить совместимость типов
	if left.Type() != ObjectType && left.Type() != ArrayType && left.Type() != right.Type() {
		panic("type error")
	}

	return &BinaryOperation{
		Left:     left,
		Right:    right,
		Operator: op,
	}
}

// Type возвращает результирующий тип операции
func (bo *BinaryOperation) Type() ExpressionType {
	// Определить результирующий тип на основе операции и типов операндов
	// Например: int + int = int, int < int = bool
	// Арифметические операторы
	switch bo.Operator {
	case ADD:
		if bo.Left.Type() == BoolType || bo.Right.Type() == BoolType {
			panic("BoolType in ADD operation")
		}
		if bo.Left.Type() == ArrayType || bo.Right.Type() == ArrayType {
			panic("ArrayType in ADD operation")
		}
		if bo.Left.Type() == FloatType || bo.Right.Type() == FloatType {
			return FloatType
		}
		return IntType
	case SUB:
		if bo.Left.Type() == BoolType || bo.Right.Type() == BoolType {
			panic("BoolType in SUB operation")
		}
		if bo.Left.Type() == ArrayType || bo.Right.Type() == ArrayType {
			panic("ArrayType in SUB operation")
		}
		if bo.Left.Type() == FloatType || bo.Right.Type() == FloatType {
			return FloatType
		}
		return IntType
	case MUL:
		if bo.Left.Type() == BoolType || bo.Right.Type() == BoolType {
			panic("BoolType in MUL operation")
		}
		if bo.Left.Type() == ArrayType || bo.Right.Type() == ArrayType {
			panic("ArrayType in MUL operation")
		}
		if bo.Left.Type() == FloatType || bo.Right.Type() == FloatType {
			return FloatType
		}
		return IntType
	case MOD:
		if bo.Left.Type() == BoolType || bo.Right.Type() == BoolType {
			panic("BoolType in MOD operation")
		}
		if bo.Left.Type() == ArrayType || bo.Right.Type() == ArrayType {
			panic("ArrayType in MOD operation")
		}
		if bo.Left.Type() == FloatType || bo.Right.Type() == FloatType {
			panic("FloatType in MOD operation")
		}
		return IntType
	case DIV:
		if bo.Left.Type() == BoolType || bo.Right.Type() == BoolType {
			panic("BoolType in DIV operation")
		}
		if bo.Left.Type() == ArrayType || bo.Right.Type() == ArrayType {
			panic("ArrayType in DIV operation")
		}
		if bo.Left.Type() == FloatType || bo.Right.Type() == FloatType {
			return FloatType
		}
		return IntType

	// Операторы сравнения
	case NE, EQ:
		return BoolType
	case LE, LT, GE, GT:
		if bo.Left.Type() == ArrayType || bo.Right.Type() == ArrayType {
			panic("ArrayType in LE/LT/GE/GT operation")
		}
		return BoolType

	case SELECT:
		return bo.Left.(*SymbolicVariable).InnerType.ExprTy

	case FIELD_ACCESS:
		return bo.Left.Type()
	case FIELD_ASSIGN:
		return bo.Right.Type()

	case XOR:
		if bo.Left.Type() != IntType || bo.Right.Type() != IntType {
			panic("Non-integer type in bitwise XOR operation")
		}
		return IntType
	case BITOR:
		if bo.Left.Type() != IntType || bo.Right.Type() != IntType {
			panic("Non-integer type in bitwise OR operation")
		}
		return IntType
	case BITAND:
		if bo.Left.Type() != IntType || bo.Right.Type() != IntType {
			panic("Non-integer type in bitwise AND operation")
		}
		return IntType
	case SHL:
		if bo.Left.Type() != IntType || bo.Right.Type() != IntType {
			panic("Non-integer type in SHL operation")
		}
		return IntType
	case SHR:
		if bo.Left.Type() != IntType || bo.Right.Type() != IntType {
			panic("Non-integer type in SHR operation")
		}
		return IntType

	default:
		panic("unknown binary operator")
	}
}

// String возвращает строковое представление операции
func (bo *BinaryOperation) String() string {
	// Формат: "(left operator right)"
	return "(" + bo.Left.String() + bo.Operator.String() + bo.Right.String() + ")"
}

// Accept реализует Visitor pattern
func (bo *BinaryOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitBinaryOperation(bo)
}

// LogicalOperation представляет логическую операцию
type LogicalOperation struct {
	Operands []SymbolicExpression
	Operator LogicalOperator
}

// NewLogicalOperation создаёт новую логическую операцию
func NewLogicalOperation(operands []SymbolicExpression, op LogicalOperator) *LogicalOperation {
	// Создать логическую операцию и проверить типы операндов
	for i := range operands {
		if operands[i].Type() == ArrayType {
			panic("ArrayType in logical expression")
		}
	}
	return &LogicalOperation{
		Operands: operands,
		Operator: op,
	}
}

// Type возвращает тип логической операции (всегда bool)
func (lo *LogicalOperation) Type() ExpressionType {
	return BoolType
}

// String возвращает строковое представление логической операции
func (lo *LogicalOperation) String() string {
	// Для NOT: "!operand"
	// Для AND/OR: "(operand1 && operand2 && ...)"
	// Для IMPLIES: "(operand1 => operand2)"
	switch lo.Operator {
	case AND:
		return "(" + lo.Operands[0].String() + " && " + lo.Operands[1].String() + ")"
	case OR:
		return "(" + lo.Operands[0].String() + " || " + lo.Operands[1].String() + ")"
	case NOT:
		return "!" + lo.Operands[0].String()
	case IMPLIES:
		return "(" + lo.Operands[0].String() + " => " + lo.Operands[1].String() + ")"
	default:
		return "unknown"
	}
}

// Accept реализует Visitor pattern
func (lo *LogicalOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitLogicalOperation(lo)
}

// Операторы для бинарных выражений
type BinaryOperator int

const (
	// Арифметические операторы
	ADD BinaryOperator = iota
	SUB
	MUL
	DIV
	MOD

	// Операторы сравнения
	EQ // равно
	NE // не равно
	LT // меньше
	LE // меньше или равно
	GT // больше
	GE // больше или равно

	// Доступ к элементу массива
	SELECT
	STORE

	// Доступ к объектам
	FIELD_ACCESS
	FIELD_ASSIGN

	XOR
	BITOR  // Bitwise OR operation
	BITAND // Bitwise AND operation
	SHL    // Shift left
	SHR    // Shift right
)

// String возвращает строковое представление оператора
func (op BinaryOperator) String() string {
	switch op {
	case ADD:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case EQ:
		return "=="
	case NE:
		return "!="
	case LT:
		return "<"
	case LE:
		return "<="
	case GT:
		return ">"
	case GE:
		return ">="
	case SELECT:
		return "@"
	case STORE:
		return "="
	case FIELD_ACCESS:
		return "."
	case FIELD_ASSIGN:
		return "="
	case XOR:
		return "^"
	case BITOR:
		return "|"
	case BITAND:
		return "&"
	case SHL:
		return "<<"
	case SHR:
		return ">>"
	default:
		return "unknown"
	}
}

// Логические операторы
type LogicalOperator int

const (
	AND LogicalOperator = iota
	OR
	NOT
	IMPLIES
)

// String возвращает строковое представление логического оператора
func (op LogicalOperator) String() string {
	switch op {
	case AND:
		return "&&"
	case OR:
		return "||"
	case NOT:
		return "!"
	case IMPLIES:
		return "=>"
	default:
		return "unknown"
	}
}

// TernaryOperation представляет тернарный оператор (ConditionalExpression)
type TernaryOperation struct {
	Condition SymbolicExpression
	TrueExpr  SymbolicExpression
	FalseExpr SymbolicExpression
}

// NewTernaryOperation создаёт новый тернарный оператор
func NewTernaryOperation(cond SymbolicExpression, trueExpr, falseExpr SymbolicExpression) *TernaryOperation {
	// Создать тернарный оператор и проверить типы
	if trueExpr.Type() != falseExpr.Type() {
		panic("incompatible types in trueExpr and falseExpr")
	}
	if cond.Type() != BoolType {
		panic("non-bool type in ternary operation condition")
	}

	return &TernaryOperation{
		Condition: cond,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

// Accept реализует Visitor pattern
func (lo *TernaryOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitTernaryOperation(lo)
}

// String возвращает строковое представление тернарного оператора
func (op TernaryOperation) String() string {
	return "(if " + op.Condition.String() + " then " + op.TrueExpr.String() + " else " + op.FalseExpr.String() + ")"
}

// Type возвращает тип тернарного оператора (тип возвращаемого в TrueExpr или FalseExpr значения)
func (to *TernaryOperation) Type() ExpressionType {
	return to.TrueExpr.Type()
}

// Операторы для унарных выражений
type UnaryOperator int

const (
	UN_SUB UnaryOperator = iota
	UN_NOT
)

// String возвращает строковое представление унарного оператора
func (uo *UnaryOperator) String() string {
	switch *uo {
	case UN_NOT:
		return "!"
	case UN_SUB:
		return "-"
	default:
		panic("unknown unary operator")
	}
}

// UnaryOperation представляет унарный оператор
type UnaryOperation struct {
	Operator UnaryOperator
	Expr     SymbolicExpression
}

// NewUnaryOperation создаёт новый тернарный оператор
func NewUnaryOperation(op UnaryOperator, expr SymbolicExpression) *UnaryOperation {
	// Создать унарный оператор и проверить типы
	if op != UN_NOT && op != UN_SUB {
		panic("invalid operation in UnaryOperation")
	}
	if op == UN_NOT && expr.Type() != BoolType {
		panic("incompatible type for UN_NOT in UnaryExpression")
	}
	if op == UN_SUB && expr.Type() != IntType {
		panic("incompatible type for UN_SUB in UnaryExpression")
	}

	return &UnaryOperation{
		Operator: op,
		Expr:     expr,
	}
}

// Accept реализует Visitor pattern
func (uo *UnaryOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitUnaryOperation(uo)
}

// String возвращает строковое представление тернарного оператора
func (op UnaryOperation) String() string {
	return "(" + op.Operator.String() + op.Expr.String() + ")"
}

// Type возвращает тип унарного оператора (тип возвращаемого в Expr значения)
func (uo *UnaryOperation) Type() ExpressionType {
	return uo.Expr.Type()
}

func Type2Sort(ctx *z3.Context, ty *InnerType) z3.Sort {
	switch ty.ExprTy {
	case IntType:
		return ctx.IntSort()
	case BoolType:
		return ctx.BoolSort()
	case FloatType:
		return ctx.FloatSort(11, 53) //<-- standart double precision values
	case ArrayType:
		return ctx.ArraySort(ctx.IntSort(), Type2Sort(ctx, ty.InnerTy))
	case ObjectType:
		panic("ObjectType in Type2Sort")
	case RefType:
		return ctx.ArraySort(ctx.IntSort(), ctx.BoolSort()) // FIXME

	default:
		print(ty.ExprTy.String())
		panic("unknown type")
	}
}

func Type2Sort2(ctx *z3.Context, expr SymbolicExpression) z3.Sort {
	switch expr.Type() {
	case IntType:
		return ctx.IntSort()
	case BoolType:
		return ctx.BoolSort()
	case FloatType:
		return ctx.FloatSort(11, 53) //<-- standart double precision values
	case ArrayType:
		innerT := expr.(*SymbolicVariable).InnerType
		return ctx.ArraySort(ctx.IntSort(), Type2Sort(ctx, &innerT))
	case ObjectType:
		panic("ObjectType in Type2Sort2")
	case RefType:
		return ctx.IntSort()

	default:
		panic("unknown type")
	}
}

// Function представляет функцию
type Function struct {
	Name    string
	Args    []InnerType
	RetType InnerType
}

// NewFunction создаёт новую функцию
func NewFunction(name string, argsTypes []InnerType, retTy InnerType) *Function {
	return &Function{
		Name:    name,
		Args:    argsTypes,
		RetType: retTy,
	}
}

// Type возвращает тип переменной
func (sv *Function) Type() ExpressionType {
	return FunctionType
}

// String возвращает строковое представление переменной
func (sv *Function) String() string {
	return "(" + sv.Name + ": Arg1 x Arg2 x ... x ArgN -> " + sv.RetType.ExprTy.String() + ")"
}

// Accept реализует Visitor pattern
func (sv *Function) Accept(visitor Visitor) interface{} {
	return visitor.VisitFunction(sv)
}

type FunctionCall struct {
	FunctionDecl Function
	Args         []SymbolicExpression
}

// NewFunctionCall создаёт вызов функции
func NewFunctionCall(function Function, args []SymbolicExpression) *FunctionCall {
	return &FunctionCall{
		FunctionDecl: function,
		Args:         args,
	}
}

// Type возвращает тип переменной
func (sv *FunctionCall) Type() ExpressionType {
	return sv.FunctionDecl.RetType.ExprTy
}

// String возвращает строковое представление переменной
func (sv *FunctionCall) String() string {
	return "(" + sv.FunctionDecl.Name + "(Arg1, Arg2, ..., ArgN))"
}

// Accept реализует Visitor pattern
func (sv *FunctionCall) Accept(visitor Visitor) interface{} {
	return visitor.VisitFunctionCall(sv)
}

type MemType int

const (
	Primitive MemType = iota
	Object
	Array
)

type Ref struct {
	Expr       SymbolicExpression
	MemTy      MemType
	ObjectAddr int
	ArrayAddr  int
	StructName string
}

func (ref *Ref) Type() ExpressionType {
	return RefType
}

func (ref *Ref) String() string {
	return "&" + ref.Expr.String()
}

func (ref *Ref) Accept(visitor Visitor) interface{} {
	return visitor.VisitRef(ref)
}

type FieldAccess struct {
	Obj        SymbolicExpression
	ObjAddr    *Ref
	FieldIdx   int
	Key        SymbolicExpression
	StructName string
	InnerTy    InnerType
}

// NewFieldAccess создаёт новое обращение к полю объекта
func NewFieldAccess(obj SymbolicExpression, objAddr *Ref, Idx int, key SymbolicExpression, structName string, innerTy InnerType) *FieldAccess {
	return &FieldAccess{
		Obj:        obj,
		ObjAddr:    objAddr,
		FieldIdx:   Idx,
		Key:        key,
		StructName: structName,
		InnerTy:    innerTy,
	}
}

// Type возвращает тип поля
func (fa *FieldAccess) Type() ExpressionType {
	return fa.Obj.Type()
}

// String возвращает строковое представление доступа к элементу массива
func (fa *FieldAccess) String() string {
	return fa.Obj.String() + "." + strconv.Itoa(fa.FieldIdx)
}

// Accept реализует Visitor pattern
func (fa *FieldAccess) Accept(visitor Visitor) interface{} {
	return visitor.VisitFieldAccess(fa)
}

type FieldAssign struct {
	Obj        SymbolicExpression
	ObjAddr    *Ref
	FieldIdx   int
	Value      SymbolicExpression
	StructName string
}

// NewFieldAccess создаёт новое обращение к полю объекта
func NewFieldAssign(obj SymbolicExpression, objAddr *Ref, Idx int, v SymbolicExpression, structName string) *FieldAssign {
	return &FieldAssign{
		Obj:        obj,
		ObjAddr:    objAddr,
		FieldIdx:   Idx,
		Value:      v,
		StructName: structName,
	}
}

// Type возвращает тип поля
func (fa *FieldAssign) Type() ExpressionType {
	return fa.Value.Type()
}

// String возвращает строковое представление переменной
func (fa *FieldAssign) String() string {
	return "(" + fa.Obj.String() + "." + strconv.Itoa(fa.FieldIdx) + "=" + fa.Value.String() + ")"
}

// Accept реализует Visitor pattern
func (fa *FieldAssign) Accept(visitor Visitor) interface{} {
	return visitor.VisitFieldAssign(fa)
}

// Below are the same as above, but with SymbolicExpression as field indexes

type FieldAccessValueIdx struct {
	Obj        SymbolicExpression
	ObjAddr    *Ref
	FieldIdx   SymbolicExpression
	Key        SymbolicExpression
	StructName string
	InnerTy    InnerType
}

// NewFieldAccess создаёт новое обращение к полю объекта
func NewFieldAccessValueIdx(obj SymbolicExpression, objAddr *Ref, Idx SymbolicExpression, structName string, innerTy InnerType) *FieldAccessValueIdx {
	return &FieldAccessValueIdx{
		Obj:        obj,
		ObjAddr:    objAddr,
		FieldIdx:   Idx,
		StructName: structName,
		InnerTy:    innerTy,
	}
}

// Type возвращает тип поля
func (fa *FieldAccessValueIdx) Type() ExpressionType {
	return fa.Obj.Type()
}

// String возвращает строковое представление доступа к элементу массива
func (fa *FieldAccessValueIdx) String() string {
	return fa.Obj.String() + "." + fa.FieldIdx.String()
}

// Accept реализует Visitor pattern
func (fa *FieldAccessValueIdx) Accept(visitor Visitor) interface{} {
	return visitor.VisitFieldAccessValueIdx(fa)
}

type FieldAssignValueIdx struct {
	Obj        SymbolicExpression
	ObjAddr    *Ref
	FieldIdx   SymbolicExpression
	Value      SymbolicExpression
	StructName string
}

// NewFieldAccess создаёт новое обращение к полю объекта
func NewFieldAssignValueIdx(obj SymbolicExpression, objAddr *Ref, Idx SymbolicExpression, v SymbolicExpression, structName string) *FieldAssignValueIdx {
	return &FieldAssignValueIdx{
		Obj:        obj,
		ObjAddr:    objAddr,
		FieldIdx:   Idx,
		Value:      v,
		StructName: structName,
	}
}

// Type возвращает тип поля
func (fa *FieldAssignValueIdx) Type() ExpressionType {
	return fa.Value.Type()
}

// String возвращает строковое представление переменной
func (fa *FieldAssignValueIdx) String() string {
	return "(" + fa.Obj.String() + "." + fa.FieldIdx.String() + "=" + fa.Value.String() + ")"
}

// Accept реализует Visitor pattern
func (fa *FieldAssignValueIdx) Accept(visitor Visitor) interface{} {
	return visitor.VisitFieldAssignValueIdx(fa)
}
