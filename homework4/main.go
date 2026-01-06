package main

import (
	"fmt"
	"symbolic-execution-course/internal"
)

func main() {
	source := `
package main

func testFunction(a, b int) int {
	return a + b
}
`
	result := internal.Analyse(source, "testFunction")
	for _, interpreter := range result {
		fmt.Println(interpreter.ToString())
	}

	source2 := `
	package main

	func testFunction2(x int) int {
		if x > 0 {
			return x * 2
		} else {
			return x * -1
		}
	}
	`
	result2 := internal.Analyse(source2, "testFunction2")
	for _, interpreter := range result2 {
		fmt.Println(interpreter.ToString())
	}

	source3 := `
	package main

	func testFunction3() int {
		result := 0
		for i := 1; i < 10; i++ {
			result += 1
		}
		return result
	}
	`
	result3 := internal.Analyse(source3, "testFunction3")
	for _, interpreter := range result3 {
		fmt.Println(interpreter.ToString())
	}

	source4 := `
	package main

		func testFunction4() int {
			result := 0
			for i := 1; i < 10; i++ {
				result += i
				if (result > 30) {
					return 15
				}
			}
			return result
		}
	
	`

	result4 := internal.Analyse(source4, "testFunction4")
	for _, interpreter := range result4 {
		fmt.Println(interpreter.ToString())
	}
}
