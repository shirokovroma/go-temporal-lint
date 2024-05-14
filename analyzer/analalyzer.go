package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "temporallint",
		Doc:  "checks temporal activities and workflows calls for match types in declarations",
		Run:  run,
	}
}

type CompositeVisitor struct {
	visitors []ast.Visitor
}

func (c *CompositeVisitor) Visit(node ast.Node) ast.Visitor {
	for _, visitor := range c.visitors {
		visitor.Visit(node)
	}
	return c
}

func run(pass *analysis.Pass) (interface{}, error) {
	visitor := &CompositeVisitor{
		visitors: []ast.Visitor{
			&ExecuteActivityVisitor{pass: pass},
			&ExecuteWorkflowVisitor{pass: pass},
		},
	}
	for _, f := range pass.Files {
		ast.Walk(visitor, f)
	}
	return nil, nil
}

type ExecuteActivityVisitor struct {
	pass *analysis.Pass
}

func (v *ExecuteActivityVisitor) Visit(node ast.Node) ast.Visitor {
	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}

	fun, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return v
	}

	// Check if it's a call to workflow.ExecuteActivity
	if fun.Sel.Name == "ExecuteActivity" {
		xIdent, ok := fun.X.(*ast.Ident)
		if ok && xIdent.Name == "workflow" {
			if len(callExpr.Args) > 0 {
				// Get the type of the second argument of ExecuteActivity
				funcType := v.pass.TypesInfo.TypeOf(callExpr.Args[1])
				if sig, ok := funcType.(*types.Signature); ok {
					// check if args count match
					if sig.Params().Len() != len(callExpr.Args)-1 {
						v.pass.Reportf(node.Pos(), "The activity function %s accepts %d arguments, but %d were passed\n", callExpr.Args[1], sig.Params().Len(), len(callExpr.Args)-1)
						return v
					}
					// Get the type of the last argument of activity function
					if sig.Params().Len() > 0 {
						// check the arguments
						for i := 1; i < sig.Params().Len(); i++ {
							param := sig.Params().At(i)
							paramType := param.Type()
							arg := callExpr.Args[i+1]
							argType := v.pass.TypesInfo.TypeOf(arg)
							if !types.Identical(paramType, argType) {
								v.pass.Reportf(node.Pos(), "In the function %s, the type of argument %d is %s, but %s was passed\n", callExpr.Args[1], i+1, paramType, argType)
							}
						}
					}
				}
			}
		}
	}

	return v
}

type ExecuteWorkflowVisitor struct {
	pass *analysis.Pass
}

func (v *ExecuteWorkflowVisitor) Visit(node ast.Node) ast.Visitor {
	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}

	fun, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return v
	}

	// Check if it's a call to workflow.ExecuteWorkflow
	if fun.Sel.Name == "ExecuteWorkflow" || fun.Sel.Name == "ExecuteChildWorkflow" {
		xIdent, ok := fun.X.(*ast.Ident)
		if ok && xIdent.Name == "workflow" {
			if len(callExpr.Args) > 0 {
				// Get the type of the second argument of ExecuteWorkflow
				funcType := v.pass.TypesInfo.TypeOf(callExpr.Args[1])
				if sig, ok := funcType.(*types.Signature); ok {
					// check if args count match
					if sig.Params().Len() != len(callExpr.Args)-1 {
						v.pass.Reportf(node.Pos(), "The workflow function %s accepts %d arguments, but %d were passed\n", callExpr.Args[1], sig.Params().Len(), len(callExpr.Args)-1)
						return v
					}
					// Get the type of the last argument of workflow function
					if sig.Params().Len() > 0 {
						// check the arguments
						for i := 1; i < sig.Params().Len(); i++ {
							param := sig.Params().At(i)
							paramType := param.Type()
							arg := callExpr.Args[i+1]
							argType := v.pass.TypesInfo.TypeOf(arg)
							if !types.Identical(paramType, argType) {
								v.pass.Reportf(node.Pos(), "In the function %s, the type of argument %d is %s, but %s was passed\n", callExpr.Args[1], i+1, paramType, argType)
							}
						}
					}
				}
			}
		}
	}

	return v
}
