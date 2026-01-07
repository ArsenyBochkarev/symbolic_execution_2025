package internal

import (
	"container/heap"
	"fmt"
	"go/types"
	"symbolic-execution-course/internal/memory"
	"symbolic-execution-course/internal/ssa"
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"
)

type Analyser struct {
	Package      *types.Package
	StatesQueue  PriorityQueue
	PathSelector PathSelector
	Results      []Interpreter
	Z3Translator *translator.Z3Translator
	MaxSteps     int
	StepCount    int
}

func Analyse(source string, functionName string) []Interpreter {
	// Analyse for loops first
	source = tryUnrollLoops(source)

	builder := ssa.NewBuilder()
	f := builder.ParseAndBuildSSA(source, functionName)

	translator := translator.NewZ3Translator()
	defer translator.Close()
	a := Analyser{
		Package:      f.Pkg.Pkg,
		PathSelector: &BfsPathSelector{},
		StatesQueue:  make(PriorityQueue, 0),
		Results:      make([]Interpreter, 0),
		Z3Translator: translator,
		MaxSteps:     1000,
		StepCount:    0,
	}

	mem := memory.NewSymbolicMemory()
	initialState := Interpreter{
		CallStack: []CallStackFrame{{
			Function:     f,
			CurrentBlock: f.Blocks[0],
			LocalMemory:  make(map[string]symbolic.SymbolicExpression),
		}},
		Analyser:     &a,
		Heap:         &mem,
		JumpDepth:    0,
		MaxJumpDepth: 10,
		CallDepth:    0,
		MaxCallDepth: 1,
	}
	initialState.CallStack[0].StructsMap = make(map[string][]symbolic.ExpressionType)
	// Add function parameters to LocalMemory
	for _, param := range f.Params {
		var paramType symbolic.ExpressionType
		switch t := param.Type().Underlying().(type) {
		case *types.Basic:
			switch t.Kind() {
			case types.Int:
				paramType = symbolic.IntType
			case types.Int64:
				paramType = symbolic.IntType
			case types.Bool:
				paramType = symbolic.BoolType
			case types.Float32:
				paramType = symbolic.FloatType
			case types.Float64:
				paramType = symbolic.FloatType
			default:
				panic("unsupported param type")
			}
			initialState.CallStack[0].LocalMemory[param.Name()] = symbolic.NewSymbolicVariable(param.Name(), paramType)
		case *types.Slice:
			// Treat this as a reference
			var innerElemTy *symbolic.InnerType = nil
			var elemTy symbolic.ExpressionType
			switch t := param.Type().Underlying().(type) {
			case *types.Slice:
				switch et := t.Elem().Underlying().(type) {
				case *types.Basic:
					switch et.Kind() {
					case types.Int:
						elemTy = symbolic.IntType
					case types.Bool:
						elemTy = symbolic.BoolType
					case types.Float32:
						elemTy = symbolic.FloatType
					case types.Float64:
						elemTy = symbolic.FloatType
					default:
						panic("unsupported function array element type")
					}
				case *types.Slice:
					elemTy = symbolic.ArrayType
					innerElemTy = buildInnerType(et)
				default:
					panic("unsupported function array element type")
				}
			default:
				fmt.Printf("%T\n", t)
				panic("ill-formed slice")
			}
			var arr_proto *symbolic.SymbolicVariable
			if elemTy == symbolic.ArrayType {
				arr_proto = symbolic.NewSymbolicVariableArray(param.Name(), symbolic.InnerType{ExprTy: elemTy, InnerTy: innerElemTy})
			} else {
				arr_proto = symbolic.NewSymbolicVariableArray(param.Name(), symbolic.InnerType{ExprTy: elemTy})
			}
			typeStr := param.Type().String()
			arr := initialState.Heap.Allocate(symbolic.ObjectType, typeStr)
			arr.Expr = arr_proto

			if elemTy != symbolic.ArrayType {
				// Assigning default value
				var arr_i_assign symbolic.SymbolicExpression
				switch elemTy {
				case symbolic.IntType:
					arr_i_assign = initialState.Heap.AssignField(arr, 0, symbolic.NewIntConstant(0))
				case symbolic.BoolType:
					arr_i_assign = initialState.Heap.AssignField(arr, 0, symbolic.NewBoolConstant(false))
				case symbolic.FloatType:
					arr_i_assign = initialState.Heap.AssignField(arr, 0, symbolic.NewFloatConstant(0))
				}
				arr.Expr = arr_i_assign
			}
			initialState.CallStack[0].LocalMemory[param.Name()] = arr
		case *types.Pointer:
			// Treat this as a pointer to struct
			newRef := symbolic.NewSymbolicVariable(param.Name(), symbolic.ObjectType)
			name := "allocated_" + param.Name()
			alloc := initialState.Heap.Allocate(symbolic.ObjectType, name)
			alloc.Expr = newRef

			// Assigning default values to its fields
			switch et := t.Elem().(type) {
			case *types.Named:
				switch s := et.Underlying().(type) {
				case *types.Struct:
					initialState.CallStack[0].StructsMap[name] = make([]symbolic.ExpressionType, s.NumFields())
					for i := range s.NumFields() {
						switch ft := s.Field(i).Type().Underlying().(type) {
						case *types.Basic:
							switch ft.Kind() {
							case types.Int16:
								alloc_assign := initialState.Heap.AssignField(alloc, i, symbolic.NewIntConstant(0))
								alloc.Expr = alloc_assign
								initialState.CallStack[0].StructsMap[name][i] = symbolic.IntType
							case types.Int:
								alloc_assign := initialState.Heap.AssignField(alloc, i, symbolic.NewIntConstant(0))
								alloc.Expr = alloc_assign
								initialState.CallStack[0].StructsMap[name][i] = symbolic.IntType
							case types.Bool:
								alloc_assign := initialState.Heap.AssignField(alloc, i, symbolic.NewBoolConstant(false))
								alloc.Expr = alloc_assign
								initialState.CallStack[0].StructsMap[name][i] = symbolic.BoolType
							case types.Float32:
								alloc_assign := initialState.Heap.AssignField(alloc, i, symbolic.NewFloatConstant(0))
								alloc.Expr = alloc_assign
								initialState.CallStack[0].StructsMap[name][i] = symbolic.FloatType
							case types.Float64:
								alloc_assign := initialState.Heap.AssignField(alloc, i, symbolic.NewFloatConstant(0))
								initialState.CallStack[0].StructsMap[name][i] = symbolic.FloatType
								alloc.Expr = alloc_assign
							default:
								panic("unsupported structure field")
							}
						default:
							panic("ill-formed field")
						}
					}
				default:
					panic("ill-formed struct")
				}
			default:
				fmt.Printf("%T\n", et)
				panic("ill-formed pointer to struct as function parameter")
			}
			initialState.CallStack[0].LocalMemory[param.Name()] = alloc
		}
	}
	// Add 'nil' to memory
	initialState.CallStack[0].LocalMemory["nil"] = symbolic.NewSymbolicVariable("nil", symbolic.RefType)
	// We also need to push init state before running the analysis
	a.StatesQueue.Push(&Item{
		value:    initialState,
		priority: a.PathSelector.CalculatePriority(initialState),
	})

	a.runAnalysis(initialState)
	return a.Results
}

func (a *Analyser) runAnalysis(initialState Interpreter) {
	for a.StatesQueue.Len() > 0 && a.StepCount < a.MaxSteps {
		item := heap.Pop(&a.StatesQueue).(*Item)
		interpreter := item.value
		interpreter.Analyser = a
		currentFrame := interpreter.CallStack[len(interpreter.CallStack)-1]
		currentBlock := currentFrame.CurrentBlock

		var currentInstr int
		if len(interpreter.CallStack) > 0 {
			currentInstr = currentFrame.InstrIdx
			if currentInstr >= len(currentFrame.CurrentBlock.Instrs) {
				a.Results = append(a.Results, interpreter)
				continue
			}
		} else {
			currentInstr = 0
		}
		a.StepCount++
		newStates := interpreter.interpretDynamically(currentBlock.Instrs[currentInstr])

		for _, newState := range newStates {
			heap.Push(&a.StatesQueue, &Item{
				value:    newState,
				priority: a.PathSelector.CalculatePriority(newState),
			})
		}
	}
}
