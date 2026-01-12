// Package ssa предоставляет функции для построения SSA представления
package ssa

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// Builder отвечает за построение SSA из исходного кода Go
type Builder struct {
	fset *token.FileSet
}

// NewBuilder создаёт новый экземпляр Builder
func NewBuilder() *Builder {
	return &Builder{
		fset: token.NewFileSet(),
	}
}

// ParseAndBuildSSA парсит исходный код Go и создаёт SSA представление
// Возвращает SSA программу и функцию по имени
func (b *Builder) ParseAndBuildSSA(source string, funcName string) *ssa.Function {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "main.go", source, parser.ParseComments)
	if err != nil {
		panic("parser error")
	}
	files := []*ast.File{file}

	pkg := types.NewPackage("main", "main")

	lprog, _, err := ssautil.BuildPackage(
		&types.Config{Importer: importer.Default()}, fset, pkg, files, ssa.SanityCheckFunctions)
	if err != nil {
		panic("type error in package")
	}

	cfg := &packages.Config{Mode: packages.LoadSyntax}
	initial, err := packages.Load(cfg, "main")
	if err != nil {
		panic("error in package load")
	}
	prog, _ := ssautil.Packages(initial, ssa.PrintPackages)
	_ = prog
	lprog.Build()

	fullFuncName := funcName
	if member, ok := lprog.Members["InvokeClass"]; ok {
		if typeName, ok := member.(*ssa.Type); ok {
			mset := lprog.Prog.MethodSets.MethodSet(types.NewPointer(typeName.Type()))
			selection := mset.Lookup(pkg, funcName)
			if selection != nil {
				fnObj := lprog.Prog.MethodValue(selection)
				// Debug output of SSA built:
				fmt.Printf("\nFunc %s:\n", fnObj.Name())
				fnObj.WriteTo(os.Stdout)
				fmt.Println()
				return fnObj
			}
		}
	}

	if fnObj := lprog.Func(funcName); fnObj != nil {
		// Debug output of SSA built:
		fmt.Printf("\nFunc %s:\n", fnObj.Name())
		fnObj.WriteTo(os.Stdout)
		fmt.Println()
		return fnObj
	}

	panicStr := fmt.Sprintf("function %s not found", fullFuncName)
	panic(panicStr)
}
