package main

import (
	"fmt"
	"symbolic-execution-course/internal"
)

func main() {
	source1 := `
	package main

	func factorial(n int) int {
		if n <= 1 {
			return 1
		}
		return n * factorial(n-1)
	}

	func testRecursive(x int) int {
		if x < 5 {
			return factorial(x)
		}
		return -1
	}
	`
	result := internal.Analyse(source1, "testRecursive")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source2 := `
	package main

	func mult(a, b int) int {
		return a * b
	}

	func half(a int) int {
		return a / 2
	}

	func SimpleFormula(fst, snd int) int {
		if fst < 100 {
			return 0
		} else if snd < 100 {
			return 0
		}

		x := fst + 5
		y := half(snd)

		return mult(x, y)
	}
	`
	result = internal.Analyse(source2, "SimpleFormula")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source3 := `
package main

type InvokeClass struct {
	Value int
}

func (i *InvokeClass) DivBy(den int) int {
	return i.Value / den
}

func (i *InvokeClass) UpdateValue(newValue int) {
	i.Value = newValue
}

func changeValue(objectValue *InvokeClass, value int) {
	objectValue.Value = value
}

func initialize(value int) *InvokeClass {
	objectValue := &InvokeClass{
		Value: value,
	}
	changeValue(objectValue, 4)
	return objectValue // We'll return address of object, of type int
}

func getFive() int {
	return 5
}

func getTwo() int {
	return 2
}

func ParticularValue(invokeObject *InvokeClass) *InvokeClass {
	objectValue := &InvokeClass{
		Value: invokeObject.Value,
	}

	if objectValue.Value < 0 {
		return nil
	}
	x := getFive() * getTwo()
	y := getFive() / getTwo()

	objectValue.Value = x + y
	return objectValue
}
`
	result = internal.Analyse(source3, "initialize")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	result = internal.Analyse(source3, "ParticularValue")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source4 := `
package main

func DefaultBooleanValues() []bool {
	array := make([]bool, 3)
	if !array[1] {
		return array
	}
	return array
}

func BooleanArray(arr []bool) int {
	if len(arr) == 0 {
		return 1
	}
	if arr[0] {
		return 2
	}
	arr[0] = true
	return 3
}

func CharSizeAndIndex(a []int, x int) byte {
	if a == nil || len(a) <= x || x < 1 {
		return 255
	}
	b := make([]int, x)
	b[0] = 5
	a[x] = x
	if b[0]+a[x] > 7 {
		return 1
	}
	return 0
}

func IsIdentityMatrix(matrix [][]int) bool {
	if len(matrix) < 3 {
		return false
	}
	for i := 0; i < len(matrix); i++ {
		if len(matrix[i]) != len(matrix) {
			return false
		}
		for j := 0; j < len(matrix[i]); j++ {
			if i == j && matrix[i][j] != 1 {
				return false
			}
			if i != j && matrix[i][j] != 0 {
				return false
			}
		}
	}
	return true
}

func ReallyMultiDimensionalArray(array [][][]int) [][][]int {
	if array[1][2][3] != 12345 {
		array[1][2][3] = 12345
	} else {
		array[1][2][3] -= 12345 * 2
	}
	return array
}

func FillMultiArrayWithArray(value []int) [][]int {
	if len(value) < 2 {
		return make([][]int, 0)
	}
	for i := range value {
		value[i] += i
	}
	length := 3
	array := make([][]int, length)
	for i := 0; i < length; i++ {
		array[i] = value
	}
	return array
}
`
	result = internal.Analyse(source4, "DefaultBooleanValues")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source4, "BooleanArray")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source4, "CharSizeAndIndex")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	// Commented two tests below due to too large output
	// result = internal.Analyse(source4, "IsIdentityMatrix")
	// for _, interpreter := range result {
	// 	fmt.Println(interpreter.ToString())
	// }
	// result = internal.Analyse(source4, "FillMultiArrayWithArray")
	// for _, interpreter := range result {
	// 	fmt.Println(interpreter.ToString())
	// }
	result = internal.Analyse(source4, "ReallyMultiDimensionalArray")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source5 := `
package main
func Complement(x int) bool {
	return ^x == 1
}

func Xor(x, y int) bool {
	return (x ^ y) == 0
}

func Or(val int) bool {
	return (val | 7) == 15
}

func And(value int) bool {
	return (value & (value - 1)) == 0
}

func BooleanNot(boolA, boolB bool) int {
	d := boolA && boolB
	e := (!boolA) || boolB
	if d && e {
		return 100
	}
	return 200
}

func BooleanXorCompare(aBool, bBool bool) int {
	if aBool != bBool {
		return 1
	}
	return 0
}

func ShlWithBigLongShift(shift int64) int {
	if shift < 40 {
		return 1
	}
	if (0x77777777 << shift) == 0x77777770 {
		return 2
	}
	return 3
}
`
	result = internal.Analyse(source5, "Complement")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source5, "Xor")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source5, "Or")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source5, "And")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source5, "BooleanNot")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source5, "BooleanXorCompare")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source5, "ShlWithBigLongShift")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source6 := `
package main

type Foo struct {
	a int
}

func Aliasing(foo1 *Foo, foo2 *Foo) int {
	foo2.a = 5
	foo1.a = 2
	if foo2.a == 2 {
		return 4
	}
	return 5
}

func ArrayAliasing(arr1 []int, arr2 []int) int {
	arr2[0] = 5
	arr1[0] = 2
	if arr2[0] == 2 {
		return 4
	}
	return 5
}
`
	result = internal.Analyse(source6, "Aliasing") // Both paths are possible if we're aliasing, see solver queries from hw3
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source6, "ArrayAliasing") // Both paths are possible if we're aliasing, see solver queries from hw3
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source7 := `
package main

func CompareWithDiv(a, b float64) float64 {
	z := a + 0.5
	if (a / z) > b {
		return 1.0
	} else {
		return 0.0
	}
}

func Mul(a, b float64) float64 {
	if a*b > 33.32 && a*b < 33.333 {
		return 1.1
	} else if a*b > 33.333 && a*b < 33.7592 {
		return 1.2
	} else {
		return 1.3
	}
}`
	result = internal.Analyse(source7, "CompareWithDiv")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source7, "Mul")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source8 := `
package main

func LoopWithConcreteBound(n int) int {
	result := 0
	for i := 0; i < 10; i++ {
		result += i
	}
	return result
}

func LoopWithSymbolicBound(n int) int {
	if n > 10 {
		return -1
	}

	result := 0
	for i := 0; i < n; i++ {
		result += i
	}
	return result
}

func LoopWithSymbolicBoundAndComplexControlFlow(n int, condition bool) int {
	if n > 10 {
		return -1
	}

	result := 0
	for i := 0; i < n; i++ {
		if condition {
			if i == 3 {
				return result
			}
		}
		if i%2 != 0 {
			// 'continue' was here, but I removed it for unrolling purposes
			// Just not to overcomplicate the loop_analyser with Phi's insertion, etc
		} else {
			result += i
		}
	}
	return result
}

func LoopInsideLoop(x int) int {
	for i := x - 5; i < x; i++ {
		if i < 0 {
			return 2
		} else {
			for j := i; j < x+i; j++ {
				if j == 7 {
					return 1
				}
			}
		}
	}
	return -1
}
`
	result = internal.Analyse(source8, "LoopWithConcreteBound")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source8, "LoopWithSymbolicBound")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	// Commented two tests below due to too large output
	// `false` is default value for `condition`
	// result = internal.Analyse(source8, "LoopWithSymbolicBoundAndComplexControlFlow")
	// for _, interpreter := range result {
	// 	fmt.Println(interpreter.ToString())
	// }
	// result := internal.Analyse(source8, "LoopInsideLoop")
	// for _, interpreter := range result {
	// 	fmt.Println(interpreter.ToString())
	// }

	source9 := `
package main

type ObjectWithPrimitivesClass struct {
	ValueByDefault int
	x, y           int
	ShortValue     int16
	Weight         float64
}

func NewObjectWithPrimitivesClass() *ObjectWithPrimitivesClass {
	return &ObjectWithPrimitivesClass{ValueByDefault: 5}
}

func Max(fst, snd *ObjectWithPrimitivesClass) *ObjectWithPrimitivesClass {
	if fst.x > snd.x && fst.y > snd.y {
		return fst
	} else if fst.x < snd.x && fst.y < snd.y {
		return snd
	}
	return fst
}

func Example(value *ObjectWithPrimitivesClass) *ObjectWithPrimitivesClass {
	if value.x == 1 {
		return value
	}
	value.x = 1
	return value
}

func CreateObject(a, b int, objectExample *ObjectWithPrimitivesClass) *ObjectWithPrimitivesClass {
	object := NewObjectWithPrimitivesClass()
	object.x = a + 5
	object.y = b + 6
	object.Weight = objectExample.Weight
	if object.Weight < 0.0 {
		return nil
	}
	return object
}

func Memory(objectExample *ObjectWithPrimitivesClass, value int) *ObjectWithPrimitivesClass {
	if value > 0 {
		objectExample.x = 1
		objectExample.y = 2
		objectExample.Weight = 1.2
	} else {
		objectExample.x = -1
		objectExample.y = -2
		objectExample.Weight = -1.2
	}
	return objectExample
}

func CompareTwoNullObjects(value int) int {
	fst := NewObjectWithPrimitivesClass()
	snd := NewObjectWithPrimitivesClass()

	fst.x = value + 1
	snd.x = value + 2

	if fst.x == value+1 {
		fst = nil
	}
	if snd.x == value+2 {
		snd = nil
	}

	if fst == snd {
		return 1
	}
	return 0
}
`
	result = internal.Analyse(source9, "Max")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source9, "Example")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source9, "Memory")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source9, "CreateObject")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
	result = internal.Analyse(source9, "CompareTwoNullObjects")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}
}
