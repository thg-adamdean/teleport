package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/ast/astutil"
)

type AuditEventEmitCollection struct {
	Calls []*ast.SelectorExpr
	mu    *sync.Mutex
}

func (a *AuditEventEmitCollection) addEmitCall(c *ast.SelectorExpr) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Calls = append(a.Calls, c)
}

// TODO: Change addEmitCall so it records the args as well. Maybe make Calls a slice of some type rather than *ast.SelectorExpr.

func main() {
	col := AuditEventEmitCollection{
		Calls: []*ast.SelectorExpr{},
		mu:    &sync.Mutex{},
	}
	s := token.NewFileSet()
	filepath.Walk(path.Join("..", ".."), func(pth string, i fs.FileInfo, err error) error {
		if strings.HasSuffix(i.Name(), ".go") {
			f, err := parser.ParseFile(s, pth, nil, 0)
			for _, d := range f.Decls {
				astutil.Apply(d, func(c *astutil.Cursor) bool {
					if i, ok := c.Node().(*ast.CallExpr); ok {
						if e, ok := i.Fun.(*ast.SelectorExpr); ok && e.Sel.Name == "EmitAuditEvent" {
							col.addEmitCall(e)
						}

					}
					return true
				}, nil)
			}
			if err != nil {
				// TODO: Replace with proper logger call
				fmt.Fprintf(os.Stderr, "error parsing Go source files: %v", err)
				os.Exit(1)
			}
		}
		return nil
	})
}
