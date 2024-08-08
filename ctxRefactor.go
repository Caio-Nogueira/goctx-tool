package ctx

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

var contextTypes = []string{
	"context.Context",
	"go.mongodb.org/mongo-driver/mongo.SessionContext",
	"github.com/gin-gonic/gin.Context",
}

func LoadPackages(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedFiles,
		Tests: true,
		Dir:   dir,
	}
	return packages.Load(cfg, "./...")
}

// TraverseAST traverses an AST starting on any Node type. Ideally this should start in ast.File anr return a bool indicating whether that file was modified or not.
func TraverseAST(file ast.Node, info *types.Info) bool {
	res := false
	ast.Inspect(file, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			idx := findContextIdx(fn.Type.Params.List, info)
			if idx > 0 {
				refactorContextPos(fn.Type.Params.List, idx)
				res = true
			}

		case *ast.FuncLit:
			idx := findContextIdx(fn.Type.Params.List, info)
			if idx > 0 {
				refactorContextPos(fn.Type.Params.List, idx)
				res = true
			}

		case *ast.CallExpr:
			idx := findContextFnCall(fn.Args, info)
			if idx > 0 {
				fn.Args = refactorContextArgPos(fn.Args, idx)
				res = true
			}

		case *ast.InterfaceType:
			for _, method := range fn.Methods.List {
				if fn, ok := method.Type.(*ast.FuncType); ok {
					idx := findContextIdx(fn.Params.List, info)
					if idx > 0 {
						refactorContextPos(fn.Params.List, idx)
						res = true
					}
				}
			}
		}

		return true // traverse all nodes
	})

	return res
}

func isContext(expr ast.Expr, info *types.Info) bool {
	typ := info.TypeOf(expr)
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}

	for _, contextType := range contextTypes {
		if typ.String() == contextType {
			return true
		}
	}
	return false
}

func findContextIdx(params []*ast.Field, info *types.Info) int {
	for idx, field := range params {
		if isContext(field.Type, info) {
			return idx
		}
	}
	return 0
}

func findContextFnCall(args []ast.Expr, info *types.Info) int {
	for idx, field := range args {
		if isContext(field, info) {
			return idx
		}
	}
	return 0
}

func refactorContextPos(list []*ast.Field, i int) {
	if i == 0 {
		return
	}
	list[0], list[i] = list[i], list[0]
}

func refactorContextArgPos(args []ast.Expr, idx int) []ast.Expr {
	if idx == 0 {
		return args
	}
	args[0], args[idx] = args[idx], args[0]
	return args
}

