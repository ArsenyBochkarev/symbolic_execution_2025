package symbolic

// Visitor интерфейс для обхода символьных выражений (Visitor Pattern)
type Visitor interface {
	VisitVariable(expr *SymbolicVariable) interface{}
	VisitIntConstant(expr *IntConstant) interface{}
	VisitBoolConstant(expr *BoolConstant) interface{}
	VisitBinaryOperation(expr *BinaryOperation) interface{}
	VisitLogicalOperation(expr *LogicalOperation) interface{}
	VisitTernaryOperation(expr *TernaryOperation) interface{}
	VisitUnaryOperation(expr *UnaryOperation) interface{}
	VisitFunction(expr *Function) interface{}
	VisitFunctionCall(expr *FunctionCall) interface{}
	VisitRef(expr *Ref) interface{}
	VisitFieldAccess(expr *FieldAccess) interface{}
	VisitFieldAssign(expr *FieldAssign) interface{}
}
