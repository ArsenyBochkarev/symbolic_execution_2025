package internal

import (
	"container/heap"
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
	}

	mem := memory.NewSymbolicMemory()
	initialState := Interpreter{
		CallStack: []CallStackFrame{{
			Function:     f,
			CurrentBlock: f.Blocks[0],
			LocalMemory:  make(map[string]symbolic.SymbolicExpression),
		}},
		Analyser: &a,
		Heap:     &mem,
	}
	// Add function parameters to LocalMemory
	for _, param := range f.Params {
		initialState.CallStack[0].LocalMemory[param.Name()] = symbolic.NewSymbolicVariable(param.Name(), symbolic.IntType)
	}
	// We also need to push init state before running the analysis
	a.StatesQueue.Push(&Item{
		value:    initialState,
		priority: a.PathSelector.CalculatePriority(initialState),
	})

	a.runAnalysis(initialState)
	return a.Results
}

func (a *Analyser) runAnalysis(initialState Interpreter) {
	instrsCount := 0
	var prevState Interpreter = initialState
	for a.StatesQueue.Len() > 0 {
		item := a.StatesQueue.Pop().(*Item)
		currentState := item.value
		currentBlock := currentState.CallStack[len(currentState.CallStack)-1].CurrentBlock

		// We should check if it is next BB
		prevBB := prevState.CallStack[len(prevState.CallStack)-1].CurrentBlock
		prevBBIndex := prevBB.Index

		if prevBBIndex != currentBlock.Index {
			instrsCount = 0
		}

		if instrsCount >= len(currentBlock.Instrs) {
			// Current stopping strategy: end of BB & nothing in the queue
			a.Results = append(a.Results, currentState)
			instrsCount = 0
			// All its successors should already be in the StatesQueue (after interpretDynamically)
			continue
		}
		nextInstr := currentBlock.Instrs[instrsCount]
		newStates := currentState.interpretDynamically(nextInstr)

		for _, newState := range newStates {
			heap.Push(&a.StatesQueue, &Item{
				value:    newState,
				priority: a.PathSelector.CalculatePriority(newState),
			})
		}
		instrsCount += 1
		prevState = currentState
	}
}
