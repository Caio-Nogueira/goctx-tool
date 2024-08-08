package main

import (
	ctx "github.com/Caio-Nogueira/goctx-tool"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

func main() {
	// Run this tool to consume the AST of the Go source code in the current directory and refactor the Context parameter to the first parameter of the function.
	pkgs, err := ctx.LoadPackages(".")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			modified := ctx.TraverseAST(file, pkg.TypesInfo)
			if !modified {
				continue
			}

			if err := writeFile(pkg.Fset, file); err != nil {
				panic(err)
			}
		}
	}
}

func writeFile(fset *token.FileSet, file *ast.File) error {
	path := fset.Position(file.Pos()).Filename
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return printer.Fprint(f, fset, file)
}
