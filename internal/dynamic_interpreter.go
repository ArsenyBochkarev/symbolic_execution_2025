package internal

import (
	"fmt"
	"go/constant"
	"go/token"
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
}

type CallStackFrame struct {
	Function     *ssa.Function
	CurrentBlock *ssa.BasicBlock
	LocalMemory  map[string]symbolic.SymbolicExpression
	ReturnValue  symbolic.SymbolicExpression
}

func (interpreter *Interpreter) interpretDynamically(element ssa.Instruction) []Interpreter {
	switch instr := element.(type) {
	case *ssa.UnOp:
		return interpreter.interpretUnOp(instr)
	case *ssa.BinOp:
		return interpreter.interpretBinOp(instr)
	case *ssa.Return:
		return interpreter.interpretReturn(instr)
	case *ssa.If:
		return interpreter.interpretIf(instr)

	default:
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
	if len(instr.Results) > 0 {
		interpreter.CallStack[len(interpreter.CallStack)-1].ReturnValue = interpreter.resolveExpression(instr.Results[0])
	} else {
		// void funcs (empty `return`)
		interpreter.CallStack[len(interpreter.CallStack)-1].ReturnValue = nil
	}

	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) interpretIf(instr *ssa.If) []Interpreter {
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
	falseState.CallStack[len(falseState.CallStack)-1].CurrentBlock = instr.Block().Succs[1]

	return []Interpreter{trueState, falseState}
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

	default:
		// This *should* be variable
		name := v.Name()
		if name == "" {
			name = "var" + strconv.Itoa(varCnt)
			varCnt += 1
		}
		return symbolic.NewSymbolicVariable(name, symbolic.IntType)
	}
}

func (interpreter *Interpreter) resolveConst(v *ssa.Const) symbolic.SymbolicExpression {
	constVal := v.Value
	switch constVal.Kind() {
	case constant.Bool:
		return symbolic.NewBoolConstant(constant.BoolVal(constVal))
	case constant.Int:
		return symbolic.NewIntConstant(v.Int64())

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
		panic("unknown unop")
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

	default:
		panic("unknown binop")
	}

	return symbolic.NewBinaryOperation(left, right, op)
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
