package ctx

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestContextIdxFDecl(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		expectedIdx int
		hasContext  bool
	}{
		{
			name: "ContextFirstParam",
			src: `package main
                  import "context"
                  func foo(ctx context.Context, a int) {}`,
			expectedIdx: 0,
			hasContext:  true,
		},
		{
			name: "ContextSecondParam",
			src: `package main
                  import "context"
                  func foo(a int, ctx context.Context) {}`,
			expectedIdx: 1,
			hasContext:  true,
		},
		{
			name: "NoContextParam",
			src: `package main
                  func foo(a int, b string) {}`,
			expectedIdx: 0,
			hasContext:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.src, parser.AllErrors)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			conf := types.Config{
				Importer: importer.Default(),
			}

			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Uses:  make(map[*ast.Ident]types.Object),
			}

			var confErr error
			_, confErr = conf.Check("main", fset, []*ast.File{file}, info)
			if confErr != nil {
				t.Fatalf("failed to type check: %v", confErr)
			}

			var funcDecl *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if fdecl, ok := n.(*ast.FuncDecl); ok {
					funcDecl = fdecl
					return false
				}
				return true
			})

			if funcDecl == nil {
				t.Fatalf("failed to find function declaration")
			}

			idx := findContextIdx(funcDecl.Type.Params.List, info)
			if idx != tt.expectedIdx {
				t.Errorf("expected index %d, got %d", tt.expectedIdx, idx)
			}
			if tt.hasContext {
				refactorContextPos(funcDecl.Type.Params.List, idx)
				assert.Equal(t, "ctx", funcDecl.Type.Params.List[0].Names[0].Name)
			}
		})
	}
}

func TestContextIdxFnCall(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		expectedIdx int
		hasContext  bool
	}{
		{
			name: "ContextFirstParam",
			src: `package main
                  import "context"
                  func foo(ctx context.Context, a int) {
					foo(ctx, 0)
				  } `,
			expectedIdx: 0,
			hasContext:  true,
		},
		{
			name: "ContextFirstParam",
			src: `package main
                  import "context"
				func bar(a any, ctx context.Context) {}
                  func foo(ctx context.Context, a int) {
					bar(a, ctx)
				  } `,
			expectedIdx: 1,
			hasContext:  true,
		},
		{
			name: "ContextFirstParam",
			src: `	package main
                  	import "context"
                	func foo(a, b int) {
						ctx := context.Background()	
						defer ctx.Done()
						foo(0, 0)	
				  	}`,
			expectedIdx: 0,
			hasContext:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.src, parser.AllErrors)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			conf := types.Config{
				Importer: importer.Default(),
			}

			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Uses:  make(map[*ast.Ident]types.Object),
			}
			var confErr error
			_, confErr = conf.Check("main", fset, []*ast.File{file}, info)
			if confErr != nil {
				t.Fatalf("failed to type check: %v", confErr)
			}

			var call *ast.CallExpr
			ast.Inspect(file, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					call = c
					return false
				}
				return true
			})

			if call == nil {
				t.Fatalf("failed to find function declaration")
			}

			idx := findContextFnCall(call.Args, info)
			if idx != tt.expectedIdx {
				t.Errorf("expected index %d, got %d", tt.expectedIdx, idx)
			}
			if tt.hasContext {
				call.Args = refactorContextArgPos(call.Args, idx)
				assert.Equal(t, "ctx", call.Args[0].(*ast.Ident).Name)
			}

		})
	}
}
