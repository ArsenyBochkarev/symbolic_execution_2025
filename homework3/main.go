package main

import (
	"fmt"
	"symbolic-execution-course/internal/memory"
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"

	"github.com/ebukreev/go-z3/z3"
)

func translateAndPrintRes(translator *translator.Z3Translator, expr symbolic.SymbolicExpression, name string) {
	z3Expr, _ := translator.TranslateExpression(expr)
	fmt.Printf("func %s: %s with type %T\n --------------------------------- \n", name, expr.String(), z3Expr)
}

func main() {
	var mem = memory.NewSymbolicMemory()
	translator := translator.NewZ3Translator()
	defer translator.Close()
	array_proto := symbolic.NewSymbolicVariable("array", symbolic.ObjectType)
	var array = mem.Allocate(symbolic.ArrayType, "[]int")
	array.Expr = array_proto

	mem.AssignToArray(array, 5, symbolic.NewIntConstant(10))

	var fromArray = mem.GetFromArray(array, 5, symbolic.InnerType{ExprTy: symbolic.IntType})
	println(fromArray)

	var anotherFromArray = mem.GetFromArray(array, 10, symbolic.InnerType{ExprTy: symbolic.IntType})
	println(anotherFromArray)

	println("\n======================================================\n")

	translateAndPrintRes(translator, fromArray, "fromArray")

	// type Person struct {
	// 	Age  int
	// 	ID   int
	// }

	// func testStructBasic() Person {
	// 	var p Person
	// 	p.Age = 25
	// 	p.ID = 1001
	// 	return p
	// }

	// var p Person
	// p.Age
	p_Age := symbolic.NewSymbolicVariableArray("p_Age", symbolic.InnerType{ExprTy: symbolic.IntType})
	// p.Age = 25
	ageAssign := symbolic.NewBinaryOperation(p_Age, symbolic.NewIntConstant(25), symbolic.FIELD_ASSIGN)
	translateAndPrintRes(translator, ageAssign, "testStructBasic")

	// p.ID
	p_ID := symbolic.NewSymbolicVariableArray("p_ID", symbolic.InnerType{ExprTy: symbolic.IntType})
	// p.ID = 1001
	idAssign := symbolic.NewBinaryOperation(p_ID, symbolic.NewIntConstant(1001), symbolic.FIELD_ASSIGN)
	translateAndPrintRes(translator, idAssign, "testStructBasic")

	//	func testStructModification(p Person) Person {
	//		p.Age = p.Age + 1
	//		p.ID = p.ID * 2
	//		return p
	//	}
	Addition := symbolic.NewBinaryOperation(p_Age, symbolic.NewIntConstant(1), symbolic.ADD)
	ageAssign = symbolic.NewBinaryOperation(p_Age, Addition, symbolic.FIELD_ASSIGN)
	translateAndPrintRes(translator, ageAssign, "testStructModification")
	mul := symbolic.NewBinaryOperation(p_ID, symbolic.NewIntConstant(2), symbolic.MUL)
	idAssign = symbolic.NewBinaryOperation(p_ID, mul, symbolic.FIELD_ASSIGN)
	translateAndPrintRes(translator, idAssign, "testStructModification")

	// func testStructPointer() *Person {
	// 	p := &Person{Name: "Bob", Age: 30, ID: 2002}
	// 	p.Age = p.Age + 5
	// 	return p
	// }
	mem = memory.NewSymbolicMemory()

	// We should allocate as follows:
	p_proto := symbolic.NewSymbolicVariable("p", symbolic.ObjectType)
	p := mem.Allocate(symbolic.ObjectType, "Person")
	p.Expr = p_proto

	mem.AssignField(p, 0, symbolic.NewIntConstant(30))   // Initialize Age
	mem.AssignField(p, 1, symbolic.NewIntConstant(2002)) // Initialize ID

	// We also must set the field type (as we cannot deduce it)
	extractField := mem.GetFieldValue(p, 0, symbolic.InnerType{ExprTy: symbolic.IntType})
	Addition = symbolic.NewBinaryOperation(extractField, symbolic.NewIntConstant(5), symbolic.ADD)
	mem.AssignField(p, 0, Addition)

	extractField = mem.GetFieldValue(p, 0, symbolic.InnerType{ExprTy: symbolic.IntType})
	translateAndPrintRes(translator, extractField, "testStructPointer")

	// type Foo struct {
	// 	a int
	// }
	//
	// func Aliasing(foo1 *Foo, foo2 *Foo) int {
	// 	foo2.a = 5
	// 	foo1.a = 2
	// 	if foo2.a == 2 {
	// 		return 4
	// 	}
	// 	return 5
	// }

	foo1_proto := symbolic.NewSymbolicVariable("foo1", symbolic.ObjectType)
	foo1 := mem.Allocate(symbolic.ObjectType, "Foo")
	foo1.Expr = foo1_proto

	foo2_proto := symbolic.NewSymbolicVariable("foo2", symbolic.ObjectType)
	foo2 := mem.Allocate(symbolic.ObjectType, "Foo")
	foo2.Expr = foo2_proto

	// Our function:
	foo2_assign := mem.AssignField(foo2, 0, symbolic.NewIntConstant(5))
	translator.TranslateExpression(foo2_assign) // We have to translate it, or the assignment will not be performed by the solver
	foo1_assign := mem.AssignField(foo1, 0, symbolic.NewIntConstant(2))
	translator.TranslateExpression(foo1_assign)

	foo2_a := mem.GetFieldValue(foo2, 0, symbolic.InnerType{ExprTy: symbolic.IntType})
	cond := symbolic.NewBinaryOperation(foo2_a, symbolic.NewIntConstant(2), symbolic.EQ)
	ifStmt := symbolic.NewTernaryOperation(cond, symbolic.NewIntConstant(4), symbolic.NewIntConstant(5))
	z3Expr, _ := translator.TranslateExpression(ifStmt)

	foo2_a2 := mem.GetFieldValue(foo2, 0, symbolic.InnerType{ExprTy: symbolic.IntType})
	foo2_a2_z3Expr, _ := translator.TranslateExpression(foo2_a2)
	foo1_a2 := mem.GetFieldValue(foo1, 0, symbolic.InnerType{ExprTy: symbolic.IntType})
	foo1_a2_z3Expr, _ := translator.TranslateExpression(foo1_a2)

	println("ALIASING CHECKS:")
	ints := translator.GetContext().IntSort()
	s := z3.NewSolver(translator.GetContext())
	six := translator.GetContext().FromInt(6, ints)
	assertion := z3Expr.(z3.Int).Eq(six.(z3.Int))
	s.Assert(assertion)
	sat, _ := s.Check()
	println("test for 6")
	println(sat) // Should not be possible
	if sat {
		println(s.Model().String())
	}
	fmt.Printf("======================================\n")

	s = z3.NewSolver(translator.GetContext())
	five := translator.GetContext().FromInt(5, ints)
	assertion = z3Expr.(z3.Int).Eq(five.(z3.Int))
	s.Assert(assertion)
	sat, _ = s.Check()
	println("test for 5")
	println(sat)
	if sat {
		println(s.Model().String()) // Should be possible
	}
	fmt.Printf("======================================\n")

	s = z3.NewSolver(translator.GetContext())
	assertion = z3Expr.(z3.Int).Eq(five.(z3.Int))
	assertion2 := foo2_a2_z3Expr.(z3.Int).Eq(foo1_a2_z3Expr.(z3.Int))
	s.Assert(assertion)
	s.Assert(assertion2)
	sat, _ = s.Check()
	println("test for 5 with equal fields")
	println(sat)
	if sat {
		println(s.String()) // Should not be possible
	}
	fmt.Printf("======================================\n")

	s = z3.NewSolver(translator.GetContext())
	four := translator.GetContext().FromInt(4, ints)
	assertion = z3Expr.(z3.Int).Eq(four.(z3.Int))
	s.Assert(assertion)
	s.Assert(assertion2)
	sat, _ = s.Check()
	println("test for 4")
	println(sat)
	if sat {
		println(s.Model().String()) // Should be possible
	}
	fmt.Printf("======================================\n")

	// func testArrayFixed() [5]int {
	// 	var arr [5]int
	// 	for i := 0; i < 5; i++ {
	// 		arr[i] = i * i
	// 	}
	// 	return arr
	// }

	arr_proto := symbolic.NewSymbolicVariableArray("arr", symbolic.InnerType{ExprTy: symbolic.IntType})
	arr := mem.Allocate(symbolic.ObjectType, "[5]int")
	arr.Expr = arr_proto

	var assignments [5]symbolic.SymbolicExpression
	for i := int64(0); i < 5; i++ {
		mul_i := symbolic.NewBinaryOperation(symbolic.NewIntConstant(i), symbolic.NewIntConstant(i), symbolic.MUL)
		arr_i_assign := mem.AssignField(arr, int(i), mul_i)
		assignments[i] = arr_i_assign
	}
	arr_4 := mem.GetFieldValue(arr, 4, symbolic.InnerType{ExprTy: symbolic.IntType})
	translateAndPrintRes(translator, arr_4, "testArrayFixed")
}
