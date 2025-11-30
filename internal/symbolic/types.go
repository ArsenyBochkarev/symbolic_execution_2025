// Package symbolic определяет базовые типы символьных выражений
package symbolic

// ExpressionType представляет тип символьного выражения
type ExpressionType int

const (
	IntType ExpressionType = iota
	BoolType
	ArrayType
	FunctionType
	ObjectType
	RefType
	// Добавьте другие типы по необходимости
)

type InnerType struct {
	ExprTy  ExpressionType
	InnerTy *InnerType
}

// String возвращает строковое представление типа
func (et ExpressionType) String() string {
	switch et {
	case IntType:
		return "int"
	case BoolType:
		return "bool"
	case ArrayType:
		return "array"
	case FunctionType:
		return "function"
	case RefType:
		return "reference"
	default:
		return "unknown"
	}
}
