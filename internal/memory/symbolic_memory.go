package memory

import (
	"symbolic-execution-course/internal/symbolic"
)

type Memory interface {
	Allocate(tpe symbolic.ExpressionType) *symbolic.Ref

	AssignPrimitive(ref *symbolic.Ref, value symbolic.SymbolicExpression)
	GetPrimitive(ref *symbolic.Ref) symbolic.SymbolicExpression

	AssignField(ref *symbolic.Ref, fieldIdx int, value symbolic.SymbolicExpression) symbolic.SymbolicExpression
	GetFieldValue(ref *symbolic.Ref, fieldIdx int, fieldTy symbolic.InnerType) symbolic.SymbolicExpression

	// We can reuse AssignField and GetFieldValue for arrays
	AssignToArray(ref *symbolic.Ref, fieldIdx int, value symbolic.SymbolicExpression)
	GetFromArray(ref *symbolic.Ref, fieldIdx int, fieldTy symbolic.InnerType) symbolic.SymbolicExpression
}

type SymbolicMemory struct {
	// Note: надо использовать SMT массивы для хранения адресов
	// Note: для организации полей у объектов также используем массивы:
	//       заводим по массиву на каждое поле класса, индексация в
	//       массиве будет производиться по expression этого поля
	// Note: чтобы получать соответствующий массив для поля, делаем map
	PrimitivesMap map[symbolic.Ref]symbolic.SymbolicExpression
	ObjectsMap    map[int]map[int]symbolic.SymbolicExpression
	ObjectsCnt    int
	ArraysMap     map[int]map[int]symbolic.SymbolicExpression
	ArraysCnt     int
}

func NewSymbolicMemory() SymbolicMemory {
	return SymbolicMemory{
		PrimitivesMap: make(map[symbolic.Ref]symbolic.SymbolicExpression),
		ObjectsMap:    make(map[int]map[int]symbolic.SymbolicExpression),
	}
}

func (mem *SymbolicMemory) Allocate(tpe symbolic.ExpressionType, structName string) *symbolic.Ref {
	switch tpe {
	case symbolic.ObjectType:
		mem.ObjectsCnt += 1
		return &symbolic.Ref{MemTy: symbolic.Object, ObjectAddr: mem.ObjectsCnt, ArrayAddr: -1, StructName: structName}
	default:
		// Primitives
		return &symbolic.Ref{MemTy: symbolic.Primitive, ArrayAddr: -1, ObjectAddr: -1}
	}
}

func (mem *SymbolicMemory) AssignPrimitive(ref *symbolic.Ref, value symbolic.SymbolicExpression) {
	mem.PrimitivesMap[*ref] = value
}

func (mem *SymbolicMemory) GetPrimitive(ref *symbolic.Ref) symbolic.SymbolicExpression {
	return mem.PrimitivesMap[*ref]
}

func (mem *SymbolicMemory) AssignField(ref *symbolic.Ref, fieldIdx int, value symbolic.SymbolicExpression) symbolic.SymbolicExpression {
	if mem.ObjectsMap[ref.ObjectAddr] == nil {
		mem.ObjectsMap[ref.ObjectAddr] = make(map[int]symbolic.SymbolicExpression)
	}
	res := symbolic.NewFieldAssign(ref.Expr, fieldIdx, value, ref.StructName)
	mem.ObjectsMap[ref.ObjectAddr][fieldIdx] = ref.Expr // Remembering expression with assign
	ref.Expr = res
	return res
}

func (mem *SymbolicMemory) GetFieldValue(ref *symbolic.Ref, fieldIdx int, fieldTy symbolic.InnerType) symbolic.SymbolicExpression {
	key := mem.ObjectsMap[ref.ObjectAddr][fieldIdx] // Get expression for assign (its String() will be used as key)
	return symbolic.NewFieldAccess(ref.Expr, fieldIdx, key, ref.StructName, fieldTy)
}

func (mem *SymbolicMemory) AssignToArray(ref *symbolic.Ref, index int, value symbolic.SymbolicExpression) symbolic.SymbolicExpression {
	return mem.AssignField(ref, index, value)
}

func (mem *SymbolicMemory) GetFromArray(ref *symbolic.Ref, fieldIdx int, fieldTy symbolic.InnerType) symbolic.SymbolicExpression {
	return mem.GetFieldValue(ref, fieldIdx, fieldTy)
}
