package internal

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"symbolic-execution-course/internal/memory"
	"symbolic-execution-course/internal/symbolic"

	"golang.org/x/tools/go/ssa"
)

type Interpreter struct {
	CallStack     []CallStackFrame
	Analyser      *Analyser
	PathCondition symbolic.SymbolicExpression
	Heap          memory.Memory
	JumpDepth     int
	MaxJumpDepth  int
	CallDepth     int
	MaxCallDepth  int
	DeclaredFuncs []symbolic.Function
}

type CallStackFrame struct {
	Function       *ssa.Function
	CurrentBlock   *ssa.BasicBlock
	PreviousBlock  *ssa.BasicBlock
	LocalMemory    map[string]symbolic.SymbolicExpression
	StructsMap     map[string][]symbolic.ExpressionType
	ReturnValue    symbolic.SymbolicExpression
	InstrIdx       int
	ReturnBlock    *ssa.BasicBlock
	ReturnInstr    int
	ReturnVariable string
}

func (interpreter *Interpreter) interpretDynamically(element ssa.Instruction) []Interpreter {
	if len(interpreter.CallStack) > 0 {
		currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
		currentFrame.InstrIdx++
	}
	switch instr := element.(type) {
	case *ssa.UnOp:
		return interpreter.interpretUnOp(instr)
	case *ssa.BinOp:
		return interpreter.interpretBinOp(instr)
	case *ssa.Return:
		return interpreter.interpretReturn(instr)
	case *ssa.If:
		return interpreter.interpretIf(instr)
	case *ssa.Call:
		return interpreter.interpretCall(instr)
	case *ssa.FieldAddr:
		return interpreter.interpretFieldAddr(instr)
	case *ssa.Store:
		return interpreter.interpretStore(instr)
	case *ssa.Alloc:
		return interpreter.interpretAlloc(instr)
	case *ssa.Slice:
		return interpreter.interpretSlice(instr)
	case *ssa.IndexAddr:
		return interpreter.interpretIndexAddr(instr)
	case *ssa.MakeSlice:
		return interpreter.interpretMakeSlice(instr)
	case *ssa.Jump:
		return interpreter.interpretJump(instr)
	case *ssa.Phi:
		return interpreter.interpretPhi(instr)

	default:
		fmt.Printf("%T\n", element)
		panic("unknown ssa op")
	}
}

func (interpreter *Interpreter) interpretUnOp(instr *ssa.UnOp) []Interpreter {
	result := interpreter.resolveExpression(instr)
	interpreter.CallStack[len(interpreter.CallStack)-1].LocalMemory[instr.Name()] = result

	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretBinOp(instr *ssa.BinOp) []Interpreter {
	result := interpreter.resolveExpression(instr)
	interpreter.CallStack[len(interpreter.CallStack)-1].LocalMemory[instr.Name()] = result

	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretReturn(instr *ssa.Return) []Interpreter {
	if len(interpreter.CallStack) == 0 {
		return []Interpreter{*interpreter}
	}

	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	if len(instr.Results) > 0 {
		currentFrame.ReturnValue = interpreter.resolveExpression(instr.Results[0])
	} else {
		// void funcs (empty `return`)
		currentFrame.ReturnValue = nil
	}

	if len(interpreter.CallStack) > 1 {
		returningValue := currentFrame.ReturnValue
		currentBlock := currentFrame.CurrentBlock
		interpreter.CallStack = interpreter.CallStack[:len(interpreter.CallStack)-1]
		interpreter.CallDepth -= 1

		prevFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
		if len(interpreter.CallStack) > 0 {
			if currentFrame.ReturnVariable != "" && returningValue != nil {
				prevFrame.LocalMemory[currentFrame.ReturnVariable] = returningValue
			}
		}
		prevFrame.CurrentBlock = currentFrame.ReturnBlock
		prevFrame.InstrIdx = currentFrame.ReturnInstr
		prevFrame.PreviousBlock = currentBlock
	}

	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretIf(instr *ssa.If) []Interpreter {
	currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
	condition := interpreter.resolveExpression(instr.Cond)

	trueState := *interpreter
	trueState.CallStack = make([]CallStackFrame, len(interpreter.CallStack))
	copy(trueState.CallStack, interpreter.CallStack)

	falseState := *interpreter
	falseState.CallStack = make([]CallStackFrame, len(interpreter.CallStack))
	copy(falseState.CallStack, interpreter.CallStack)

	if trueState.PathCondition == nil {
		trueState.PathCondition = condition
	} else {
		trueState.PathCondition = symbolic.NewLogicalOperation(
			[]symbolic.SymbolicExpression{trueState.PathCondition, condition},
			symbolic.AND,
		)
	}

	negatedCond := symbolic.NewUnaryOperation(symbolic.UN_NOT, condition)
	if falseState.PathCondition == nil {
		falseState.PathCondition = negatedCond
	} else {
		falseState.PathCondition = symbolic.NewLogicalOperation(
			[]symbolic.SymbolicExpression{falseState.PathCondition, negatedCond},
			symbolic.AND,
		)
	}

	// We also need to update blocks for both branches
	trueState.CallStack[len(trueState.CallStack)-1].CurrentBlock = instr.Block().Succs[0]
	trueState.CallStack[len(trueState.CallStack)-1].InstrIdx = 0
	trueState.CallStack[len(trueState.CallStack)-1].PreviousBlock = currentFrame.CurrentBlock

	falseState.CallStack[len(falseState.CallStack)-1].CurrentBlock = instr.Block().Succs[1]
	falseState.CallStack[len(falseState.CallStack)-1].InstrIdx = 0
	falseState.CallStack[len(falseState.CallStack)-1].PreviousBlock = currentFrame.CurrentBlock

	return []Interpreter{trueState, falseState}
}

var builtinCnt int = 0

func (interpreter *Interpreter) interpretCall(instr *ssa.Call) []Interpreter {
	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	call := instr.Call.Value
	switch callee := call.(type) {
	case *ssa.Function:
		if callee.Pkg == nil || callee.Pkg.Pkg.Path() != currentFrame.Function.Pkg.Pkg.Path() {
			// We could do NewFunction and NewFunctionCall here
			panic("external call")
		}

		// We don't do NewFunction and NewFunctionCall here
		// Instead:
		// - `call` will push on the call stack
		// - traverse the function
		// - `return` will pop the stack
		newFrame := CallStackFrame{
			Function:       callee,
			CurrentBlock:   callee.Blocks[0], // Starting from first block of callee
			InstrIdx:       0,
			LocalMemory:    make(map[string]symbolic.SymbolicExpression),
			ReturnValue:    nil,
			ReturnBlock:    currentFrame.CurrentBlock,
			ReturnInstr:    currentFrame.InstrIdx,
			ReturnVariable: instr.Name(),
			PreviousBlock:  currentFrame.CurrentBlock,
		}

		params := callee.Params
		args := instr.Call.Args
		var argsTypes []symbolic.InnerType
		var symbolicArgs []symbolic.SymbolicExpression
		for i, param := range params {
			if i < len(args) {
				argValue := interpreter.resolveExpression(args[i])
				symbolicArgs = append(symbolicArgs, argValue)
				newFrame.LocalMemory[param.Name()] = argValue
				if _, ok := argValue.(*symbolic.Ref); ok && interpreter.CallDepth >= interpreter.MaxCallDepth {
					// Passing address of ref
					argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: symbolic.IntType})
				} else if fa, ok := argValue.(*symbolic.FieldAccess); ok {
					argsTypes = append(argsTypes, fa.InnerTy)
				} else if fa, ok := argValue.(*symbolic.FieldAccessValueIdx); ok {
					argsTypes = append(argsTypes, fa.InnerTy)
				} else {
					argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: argValue.Type()})
				}
			}
		}
		if interpreter.CallDepth >= interpreter.MaxCallDepth {
			var retType symbolic.ExpressionType
			switch t := callee.Signature.Results().At(0).Type().Underlying().(type) {
			case *types.Basic:
				switch t.Kind() {
				case types.Int:
					retType = symbolic.IntType
				case types.Bool:
					retType = symbolic.BoolType
				default:
					panic("unsupported return type for recursive call")
				}
			default:
				panic("unsupported return type for recursive call")
			}
			declareFunc := symbolic.NewFunction(callee.Name(), argsTypes, symbolic.InnerType{ExprTy: retType})
			interpreter.DeclaredFuncs = append(interpreter.DeclaredFuncs, *declareFunc)
			currentFrame.LocalMemory[instr.Name()] = symbolic.NewFunctionCall(*declareFunc, symbolicArgs)
			return []Interpreter{*interpreter}
		}
		interpreter.CallStack = append(interpreter.CallStack, newFrame)
		interpreter.CallDepth++

		return []Interpreter{*interpreter}

	case *ssa.Builtin:
		// Default strategy -- same as in regular calls
		args := instr.Call.Args
		var argsTypes []symbolic.InnerType
		var symbolicArgs []symbolic.SymbolicExpression
		for i := range len(args) {
			argValue := interpreter.resolveExpression(args[i])
			symbolicArgs = append(symbolicArgs, argValue)
			if _, ok := argValue.(*symbolic.Ref); ok {
				// Passing address of ref
				argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: symbolic.IntType})
			} else if fa, ok := argValue.(*symbolic.FieldAccess); ok {
				argsTypes = append(argsTypes, fa.InnerTy)
			} else if fa, ok := argValue.(*symbolic.FieldAccessValueIdx); ok {
				argsTypes = append(argsTypes, fa.InnerTy)
			} else {
				argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: argValue.Type()})
			}
		}

		var foundDeclaredFunc bool
		foundDeclaredFunc = false
		var declaredFunc *symbolic.Function
		for i := range interpreter.DeclaredFuncs {
			if interpreter.DeclaredFuncs[i].Name == callee.Name() {
				declaredFunc = &interpreter.DeclaredFuncs[i]
				foundDeclaredFunc = true
				break
			}
		}
		if !foundDeclaredFunc {
			// Declare it
			retType := symbolic.IntType
			builtinCnt += 1
			declaredFunc = symbolic.NewFunction(callee.Name()+strconv.Itoa(builtinCnt), argsTypes, symbolic.InnerType{ExprTy: retType})
			interpreter.DeclaredFuncs = append(interpreter.DeclaredFuncs, *declaredFunc)
		}

		currentFrame.LocalMemory[instr.Name()] = symbolic.NewFunctionCall(*declaredFunc, symbolicArgs)
		return []Interpreter{*interpreter}

	default:
		panic("unknown call")
	}
}

var allocCnt int = 0

func (interpreter *Interpreter) interpretFieldAddr(instr *ssa.FieldAddr) []Interpreter {
	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	proto := interpreter.resolveExpression(instr.X)
	fieldIndex := instr.Field

	var fieldTy symbolic.ExpressionType
	switch et := instr.Type().Underlying().(type) {
	case *types.Pointer:
		switch pt := et.Elem().(type) {
		case *types.Basic:
			switch pt.Kind() {
			case types.Int:
				fieldTy = symbolic.IntType
			case types.Bool:
				fieldTy = symbolic.BoolType
			case types.Float32:
				fieldTy = symbolic.FloatType
			case types.Float64:
				fieldTy = symbolic.FloatType
			}
		default:
			panic("ill-formed pointer inside struct")
		}
	default:
		panic("ill-formed struct")
	}
	var result symbolic.SymbolicExpression
	if alloc, ok := proto.(*symbolic.Ref); ok {
		result = interpreter.Heap.GetFieldValue(alloc, fieldIndex, symbolic.InnerType{ExprTy: fieldTy})
	} else {
		// Allocate it
		newVar := symbolic.NewSymbolicVariable(proto.String(), symbolic.ObjectType)
		alloc := interpreter.Heap.Allocate(symbolic.ObjectType, "allocated_"+proto.String()+"_"+strconv.Itoa(allocCnt))
		allocCnt += 1
		alloc.Expr = newVar
		var alloc_assign symbolic.SymbolicExpression
		switch fieldTy {
		case symbolic.IntType:
			alloc_assign = interpreter.Heap.AssignField(alloc, fieldIndex, symbolic.NewIntConstant(0)) // Assigning default 0 value to this field
		case symbolic.BoolType:
			alloc_assign = interpreter.Heap.AssignField(alloc, fieldIndex, symbolic.NewBoolConstant(false)) // Assigning default false value to this field
		case symbolic.FloatType:
			alloc_assign = interpreter.Heap.AssignField(alloc, fieldIndex, symbolic.NewFloatConstant(0)) // Assigning default 0 value to this field
		}
		alloc.Expr = alloc_assign
		result = interpreter.Heap.GetFieldValue(alloc, fieldIndex, symbolic.InnerType{ExprTy: fieldTy})
	}

	currentFrame.LocalMemory[instr.Name()] = result
	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretStore(instr *ssa.Store) []Interpreter {
	proto := interpreter.resolveExpression(instr.Addr)
	value := interpreter.resolveExpression(instr.Val)

	// We also letting the solver know about this store
	// As it may cause some side-effects, such as aliasing
	switch v := proto.(type) {
	case *symbolic.FieldAccess:
		assign := interpreter.Heap.AssignField(v.ObjAddr, v.FieldIdx, value)
		interpreter.Analyser.Z3Translator.TranslateExpression(assign)
	case *symbolic.FieldAccessValueIdx:
		assign := interpreter.Heap.AssignFieldFromValue(v.ObjAddr, v.FieldIdx, value)
		interpreter.Analyser.Z3Translator.TranslateExpression(assign)
	case *symbolic.FieldAssign:
		assign := interpreter.Heap.AssignField(v.ObjAddr, v.FieldIdx, value)
		interpreter.Analyser.Z3Translator.TranslateExpression(assign)
	case *symbolic.FieldAssignValueIdx:
		assign := interpreter.Heap.AssignFieldFromValue(v.ObjAddr, v.FieldIdx, value)
		interpreter.Analyser.Z3Translator.TranslateExpression(assign)
	case *symbolic.Ref:
		assign := interpreter.Heap.AssignField(v, 0, value) // Assigning to default index
		interpreter.Analyser.Z3Translator.TranslateExpression(assign)

	default:
		panic("unknown store proto")
	}

	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretAlloc(instr *ssa.Alloc) []Interpreter {
	typeStr := instr.Type().String()

	var ref *symbolic.Ref
	switch t := instr.Type().Underlying().(type) {
	case *types.Pointer:
		var paramType symbolic.ExpressionType
		switch et := t.Elem().(type) {
		case *types.Array:
			switch arr_et := et.Elem().Underlying().(type) {
			case *types.Basic:
				switch arr_et.Kind() {
				case types.Int:
					paramType = symbolic.IntType
				case types.Bool:
					paramType = symbolic.BoolType
				}
				arr_proto := symbolic.NewSymbolicVariableArray(instr.Name(), symbolic.InnerType{ExprTy: paramType})
				ref = interpreter.Heap.Allocate(symbolic.ObjectType, typeStr) // FIXME: ArrayType?
				ref.Expr = arr_proto
			case *types.Slice:
				arr_proto := symbolic.NewSymbolicVariableArray(instr.Name(), symbolic.InnerType{ExprTy: paramType, InnerTy: buildInnerType(et)})
				ref = interpreter.Heap.Allocate(symbolic.ObjectType, typeStr)
				ref.Expr = arr_proto
			default:
				fmt.Printf("%T\n", arr_et)
				panic("unsupported array element type")
			}
		case *types.Named: // A structure...
			struct_proto := symbolic.NewSymbolicVariable(instr.Name(), symbolic.ObjectType)
			ref = interpreter.Heap.Allocate(symbolic.ObjectType, typeStr)
			ref.Expr = struct_proto
		default:
			panic("unsupported array element type")
		}
	default:
		panic("unsupported instr type")
	}

	currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
	currentFrame.LocalMemory[instr.Name()] = ref

	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretSlice(instr *ssa.Slice) []Interpreter {
	typeStr := instr.Type().String()
	proto := interpreter.resolveExpression(instr.X)
	index := interpreter.resolveExpression(instr.High)

	var innerElemTy *symbolic.InnerType = nil
	var elemTy symbolic.ExpressionType
	switch t := instr.Type().Underlying().(type) {
	case *types.Slice:
		switch et := t.Elem().Underlying().(type) {
		case *types.Basic:
			switch et.Kind() {
			case types.Int:
				elemTy = symbolic.IntType
			case types.Bool:
				elemTy = symbolic.BoolType
			}
		case *types.Slice:
			elemTy = symbolic.ArrayType
			innerElemTy = buildInnerType(et)
		}
	default:
		panic("non-slice type")
	}

	var result symbolic.SymbolicExpression
	if ref, ok := proto.(*symbolic.Ref); ok {
		result = interpreter.Heap.GetFieldValueFromValue(ref, index, symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
	} else {
		// Allocate it
		arr_proto := symbolic.NewSymbolicVariableArray(instr.Name(), symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
		arr := interpreter.Heap.Allocate(symbolic.ObjectType, typeStr)
		arr.Expr = arr_proto

		// Assigning default value
		if elemTy != symbolic.ArrayType {
			var arr_i_assign symbolic.SymbolicExpression
			switch elemTy {
			case symbolic.IntType:
				arr_i_assign = interpreter.Heap.AssignFieldFromValue(arr, index, symbolic.NewIntConstant(0))
			case symbolic.BoolType:
				arr_i_assign = interpreter.Heap.AssignFieldFromValue(arr, index, symbolic.NewBoolConstant(false))
			default:
				panic("cannot assign default value")
			}
			arr.Expr = arr_i_assign
		}
		result = interpreter.Heap.GetFieldValueFromValue(arr, index, symbolic.InnerType{ExprTy: elemTy})
	}

	currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
	currentFrame.LocalMemory[instr.Name()] = result
	return []Interpreter{*interpreter}
}

func buildInnerType(t types.Type) *symbolic.InnerType {
	switch et := t.Underlying().(type) {
	case *types.Basic:
		switch et.Kind() {
		case types.Int:
			return &symbolic.InnerType{ExprTy: symbolic.IntType}
		case types.Bool:
			return &symbolic.InnerType{ExprTy: symbolic.BoolType}
		}
	case *types.Slice:
		return &symbolic.InnerType{ExprTy: symbolic.ArrayType, InnerTy: buildInnerType(et.Elem())}
	case *types.Array:
		return &symbolic.InnerType{ExprTy: symbolic.ArrayType, InnerTy: buildInnerType(et.Elem())}
	}
	panic("unreachable")
}

func (interpreter *Interpreter) interpretIndexAddr(instr *ssa.IndexAddr) []Interpreter {
	typeStr := instr.Type().String()
	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	proto := interpreter.resolveExpression(instr.X)
	index := interpreter.resolveExpression(instr.Index)

	var innerElemTy *symbolic.InnerType = nil
	var elemTy symbolic.ExpressionType
	switch t := instr.Type().Underlying().(type) {
	case *types.Pointer:
		switch et := t.Elem().(type) {
		case *types.Basic:
			switch et.Kind() {
			case types.Int:
				elemTy = symbolic.IntType
			case types.Bool:
				elemTy = symbolic.BoolType
			}
		case *types.Slice:
			elemTy = symbolic.ArrayType
			innerElemTy = buildInnerType(et)
		default:
			panic("unsupported array element type")
		}
	default:
		panic("unsupported array elements type")
	}

	var result symbolic.SymbolicExpression
	if alloc, ok := proto.(*symbolic.Ref); ok {
		result = interpreter.Heap.GetFieldValueFromValue(alloc, index, symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
	} else {
		// Allocate it
		arr_proto := symbolic.NewSymbolicVariableArray(instr.Name(), symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
		arr := interpreter.Heap.Allocate(symbolic.ObjectType, typeStr)
		arr.Expr = arr_proto

		if elemTy != symbolic.ArrayType {
			// Assigning default value
			var arr_i_assign symbolic.SymbolicExpression
			switch elemTy {
			case symbolic.IntType:
				arr_i_assign = interpreter.Heap.AssignFieldFromValue(arr, index, symbolic.NewIntConstant(0))
			case symbolic.BoolType:
				arr_i_assign = interpreter.Heap.AssignFieldFromValue(arr, index, symbolic.NewBoolConstant(false))
			default:
				panic("cannot assign default value")
			}
			arr.Expr = arr_i_assign
		}
		result = interpreter.Heap.GetFieldValueFromValue(arr, index, symbolic.InnerType{ExprTy: elemTy})
	}

	currentFrame.LocalMemory[instr.Name()] = result
	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretMakeSlice(instr *ssa.MakeSlice) []Interpreter {
	// Translating both of them since they might have some side-effects
	_ = interpreter.resolveExpression(instr.Len)
	_ = interpreter.resolveExpression(instr.Cap)

	var elemTy symbolic.ExpressionType
	switch t := instr.Type().(type) {
	case *types.Slice:
		switch et := t.Elem().Underlying().(type) {
		case *types.Basic:
			switch et.Kind() {
			case types.Int:
				elemTy = symbolic.IntType
			case types.Bool:
				elemTy = symbolic.BoolType
			}
		default:
			// panic("unsupported inner type for slice")
		}
	default:
		panic("ill-formed MakeSlice")
	}
	arr_proto := symbolic.NewSymbolicVariableArray(instr.Name(), symbolic.InnerType{ExprTy: elemTy})
	typeStr := instr.Type().String()
	arr := interpreter.Heap.Allocate(symbolic.ObjectType, typeStr)
	arr.Expr = arr_proto

	arr_assign := interpreter.Heap.AssignField(arr, 0, symbolic.NewIntConstant(0)) // Assigning default 0 value to 0
	arr.Expr = arr_assign

	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	currentFrame.LocalMemory[instr.Name()] = arr
	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretJump(instr *ssa.Jump) []Interpreter {
	nextBlock := instr.Block().Succs[0]
	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	currentFrame.CurrentBlock = nextBlock
	if interpreter.JumpDepth < interpreter.MaxJumpDepth {
		currentFrame.InstrIdx = 0
		interpreter.JumpDepth++
	}
	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretPhi(instr *ssa.Phi) []Interpreter {
	currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
	var result symbolic.SymbolicExpression
	// We should get results only from previous block
	for i, pred := range instr.Block().Preds {
		if pred == currentFrame.PreviousBlock {
			result = interpreter.resolveExpression(instr.Edges[i])
			break
		}
	}
	currentFrame.LocalMemory[instr.Name()] = result
	return []Interpreter{*interpreter}
}

var varCnt int = 0

func (interpreter *Interpreter) resolveExpression(value ssa.Value) symbolic.SymbolicExpression {
	switch v := value.(type) {
	case *ssa.Const:
		return interpreter.resolveConst(v)
	case *ssa.UnOp:
		return interpreter.resolveUnOp(v)
	case *ssa.BinOp:
		return interpreter.resolveBinOp(v)
	case *ssa.Call:
		return interpreter.resolveCall(v)
	case *ssa.IndexAddr:
		return interpreter.resolveIndexAddr(v)
	case *ssa.Parameter:
		return interpreter.resolveParameter(v)

	default:
		// This *should* be variable
		currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
		name := v.Name()
		if name == "" {
			name = "var" + strconv.Itoa(varCnt)
			varCnt += 1
		} else if currentFrame.LocalMemory[name] != nil {
			if fa, ok := currentFrame.LocalMemory[name].(*symbolic.FieldAccess); ok {
				if len(currentFrame.StructsMap[fa.StructName]) > fa.FieldIdx && fa.Type() != currentFrame.StructsMap[fa.StructName][fa.FieldIdx] {
					// We *could* do the following sequence of actions:
					// - modify/access one field
					// - do a binop with different field
					// - but the field access/assign would still hold type of the first field!
					// So I'll just hack it here with an assign of 0 value of proper type to some distant field
					fieldToAssign := 50
					var rightTy symbolic.ExpressionType
					switch currentFrame.StructsMap[fa.StructName][fa.FieldIdx] {
					case symbolic.FloatType:
						rightTy = symbolic.FloatType
						fa.ObjAddr.Expr = interpreter.Heap.AssignField(fa.ObjAddr, fieldToAssign, symbolic.NewFloatConstant(0))
					case symbolic.IntType:
						rightTy = symbolic.IntType
						fa.ObjAddr.Expr = interpreter.Heap.AssignField(fa.ObjAddr, fieldToAssign, symbolic.NewIntConstant(0))
					}
					currentFrame.LocalMemory[name] = interpreter.Heap.GetFieldValue(fa.ObjAddr, fa.FieldIdx, symbolic.InnerType{ExprTy: rightTy})
				}
			}
			return currentFrame.LocalMemory[name]
		}
		return symbolic.NewSymbolicVariable(name, symbolic.IntType)
	}
}

func (interpreter *Interpreter) resolveConst(v *ssa.Const) symbolic.SymbolicExpression {
	if v.IsNil() {
		currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
		return currentFrame.LocalMemory["nil"]
	}

	constVal := v.Value
	switch constVal.Kind() {
	case constant.Bool:
		return symbolic.NewBoolConstant(constant.BoolVal(constVal))
	case constant.Int:
		return symbolic.NewIntConstant(v.Int64())
	case constant.Float:
		return symbolic.NewFloatConstant(v.Float64())

	default:
		panic("unknown type for const")
	}
}

func (interpreter *Interpreter) resolveUnOp(v *ssa.UnOp) symbolic.SymbolicExpression {
	expr := interpreter.resolveExpression(v.X)

	switch v.Op {
	case token.NOT:
		return symbolic.NewUnaryOperation(symbolic.UN_NOT, expr)
	case token.SUB:
		return symbolic.NewUnaryOperation(symbolic.UN_SUB, expr)

	default:
		// This *should* be deref
		return expr
	}
}

func (interpreter *Interpreter) resolveBinOp(v *ssa.BinOp) symbolic.SymbolicExpression {
	left := interpreter.resolveExpression(v.X)
	right := interpreter.resolveExpression(v.Y)
	var op symbolic.BinaryOperator

	switch v.Op {
	case token.ADD:
		op = symbolic.ADD
	case token.SUB:
		op = symbolic.SUB
	case token.MUL:
		op = symbolic.MUL
	case token.QUO:
		op = symbolic.DIV
	case token.REM:
		op = symbolic.MOD

	case token.EQL:
		op = symbolic.EQ
	case token.NEQ:
		op = symbolic.NE
	case token.LSS:
		op = symbolic.LT
	case token.LEQ:
		op = symbolic.LE
	case token.GTR:
		op = symbolic.GT
	case token.GEQ:
		op = symbolic.GE

	case token.XOR:
		op = symbolic.XOR
	case token.OR:
		op = symbolic.BITOR
	case token.AND:
		op = symbolic.BITAND
	case token.SHL:
		op = symbolic.SHL
	case token.SHR:
		op = symbolic.SHR

	default:
		panic("unknown binop")
	}

	return symbolic.NewBinaryOperation(left, right, op)
}

func (interpreter *Interpreter) resolveCall(c *ssa.Call) symbolic.SymbolicExpression {
	call := c.Call.Value
	switch callee := call.(type) {
	case *ssa.Function:
		params := callee.Params
		args := c.Call.Args
		var argsTypes []symbolic.InnerType
		var symbolicArgs []symbolic.SymbolicExpression
		for i := range params {
			if i < len(args) {
				argValue := interpreter.resolveExpression(args[i])
				symbolicArgs = append(symbolicArgs, argValue)
				argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: argValue.Type()})
			}
		}

		currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
		if interpreter.CallDepth < interpreter.MaxCallDepth && currentFrame.LocalMemory[c.Name()] != nil {
			// We should have been there
			return currentFrame.LocalMemory[c.Name()]
		}

		var foundDeclaredFunc bool
		foundDeclaredFunc = false
		var declaredFunc *symbolic.Function
		for i := range interpreter.DeclaredFuncs {
			if interpreter.DeclaredFuncs[i].Name == callee.Name() {
				declaredFunc = &interpreter.DeclaredFuncs[i]
				foundDeclaredFunc = true
				break
			}
		}
		if !foundDeclaredFunc {
			// Declare it
			var retType symbolic.ExpressionType
			switch t := callee.Signature.Results().At(0).Type().Underlying().(type) {
			case *types.Basic:
				switch t.Kind() {
				case types.Int:
					retType = symbolic.IntType
				case types.Bool:
					retType = symbolic.BoolType
				default:
					panic("unsupported return type for recursive call")
				}
			default:
				panic("unsupported return type for recursive call")
			}
			declaredFunc = symbolic.NewFunction(callee.Name(), argsTypes, symbolic.InnerType{ExprTy: retType})
			interpreter.DeclaredFuncs = append(interpreter.DeclaredFuncs, *declaredFunc)
		}
		return symbolic.NewFunctionCall(*declaredFunc, symbolicArgs)

	case *ssa.Builtin:
		args := c.Call.Args
		var argsTypes []symbolic.InnerType
		var symbolicArgs []symbolic.SymbolicExpression
		for i := range len(args) {
			argValue := interpreter.resolveExpression(args[i])
			symbolicArgs = append(symbolicArgs, argValue)
			if _, ok := argValue.(*symbolic.Ref); ok {
				// Passing address of ref
				argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: symbolic.IntType})
			} else if fa, ok := argValue.(*symbolic.FieldAccess); ok {
				argsTypes = append(argsTypes, fa.InnerTy)
			} else if fa, ok := argValue.(*symbolic.FieldAccessValueIdx); ok {
				argsTypes = append(argsTypes, fa.InnerTy)
			} else {
				argsTypes = append(argsTypes, symbolic.InnerType{ExprTy: argValue.Type()})
			}
		}

		var foundDeclaredFunc bool
		foundDeclaredFunc = false
		var declaredFunc *symbolic.Function
		for i := range interpreter.DeclaredFuncs {
			if interpreter.DeclaredFuncs[i].Name == callee.Name() {
				declaredFunc = &interpreter.DeclaredFuncs[i]
				foundDeclaredFunc = true
				break
			}
		}
		if !foundDeclaredFunc {
			// Declare it
			retType := symbolic.IntType
			builtinCnt += 1
			declaredFunc = symbolic.NewFunction(callee.Name()+strconv.Itoa(builtinCnt), argsTypes, symbolic.InnerType{ExprTy: retType})
			interpreter.DeclaredFuncs = append(interpreter.DeclaredFuncs, *declaredFunc)
		}
		return symbolic.NewFunctionCall(*declaredFunc, symbolicArgs)

	default:
		panic("unsupported call in resolveCall")
	}
}

func (interpreter *Interpreter) resolveIndexAddr(instr *ssa.IndexAddr) symbolic.SymbolicExpression {
	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	if val, ok := currentFrame.LocalMemory[instr.Name()]; ok {
		return val
	}

	typeStr := instr.Type().String()
	proto := interpreter.resolveExpression(instr.X)
	index := interpreter.resolveExpression(instr.Index)

	var innerElemTy *symbolic.InnerType = nil
	var elemTy symbolic.ExpressionType
	switch t := instr.Type().Underlying().(type) {
	case *types.Pointer:
		switch et := t.Elem().(type) {
		case *types.Basic:
			switch et.Kind() {
			case types.Int:
				elemTy = symbolic.IntType
			case types.Bool:
				elemTy = symbolic.BoolType
			}
		case *types.Slice:
			elemTy = symbolic.ArrayType
			innerElemTy = buildInnerType(et)
		default:
			panic("unsupported array element type")
		}
	default:
		panic("unsupported array elements type")
	}

	var result symbolic.SymbolicExpression
	if alloc, ok := proto.(*symbolic.Ref); ok {
		result = interpreter.Heap.GetFieldValueFromValue(alloc, index, symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
	} else {
		// Allocate it
		arr_proto := symbolic.NewSymbolicVariableArray(instr.Name(), symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
		arr := interpreter.Heap.Allocate(symbolic.ObjectType, typeStr)
		arr.Expr = arr_proto

		if elemTy != symbolic.ArrayType {
			// Assigning default value
			var arr_i_assign symbolic.SymbolicExpression
			switch elemTy {
			case symbolic.IntType:
				arr_i_assign = interpreter.Heap.AssignFieldFromValue(arr, index, symbolic.NewIntConstant(0))
			case symbolic.BoolType:
				arr_i_assign = interpreter.Heap.AssignFieldFromValue(arr, index, symbolic.NewBoolConstant(false))
			default:
				panic("cannot assign default value")
			}
			arr.Expr = arr_i_assign
		}
		result = interpreter.Heap.GetFieldValueFromValue(arr, index, symbolic.InnerType{ExprTy: elemTy})
	}

	currentFrame.LocalMemory[instr.Name()] = result
	return result
}

func (interpreter *Interpreter) resolveParameter(instr *ssa.Parameter) symbolic.SymbolicExpression {
	currentFrame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	if val, ok := currentFrame.LocalMemory[instr.Name()]; ok {
		return val
	}
	println(instr.Name())
	panic("we should have initialized this before")
}

func (interpreter *Interpreter) ToString() string {
	if interpreter == nil {
		return "nil"
	}

	var builder strings.Builder
	currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
	builder.WriteString("Interpreter state:\n")

	builder.WriteString("- path condition: ")
	if interpreter.PathCondition == nil {
		builder.WriteString("true\n")
	} else {
		conditionStr := interpreter.formatExpression(interpreter.PathCondition)
		builder.WriteString(fmt.Sprintf("%s\n", conditionStr))
	}

	if currentFrame.ReturnValue != nil {
		returnStr := interpreter.formatExpression(currentFrame.ReturnValue)
		builder.WriteString(fmt.Sprintf("- return value: %s\n", returnStr))
	}

	builder.WriteString("================================================")
	return builder.String()
}

func (interpreter *Interpreter) formatExpression(expr symbolic.SymbolicExpression) string {
	z3Expr, _ := interpreter.Analyser.Z3Translator.TranslateExpression(expr)
	return fmt.Sprintf("%s: %T", expr.String(), z3Expr)
}
