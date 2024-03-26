package value

import (
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/literal"
	"cuelang.org/go/cue/token"
)

// Fix fixes the CUE code represented by x so that expressions
// of the form:
//
//	{ for k in parameter.foo { ... }}
//
// are rewritten to:
//
//	{ for k in *parameter.foo | {} { ... }}
//
// The node should have been parsed with the [parser.ParseComments]
// option.
func Fix(x ast.Node) ast.Node {
	return astutil.Apply(x, visit, nil)
}

func visit(c astutil.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.StructLit:
		visitClauses(n.Elts)
		return false
	case *ast.File:
		visitClauses(n.Decls)
		return false
	default:
		return true
	}
}

func visitClauses[T ast.Node](elts []T) {
	for i, e := range elts {
		c, ok := any(e).(*ast.Comprehension)
		if !ok {
			elts[i] = astutil.Apply(e, visit, nil).(T)
			continue
		}
		var guarded map[string]bool
		for i, clause := range c.Clauses {
			var field string
			if isGuard(clause, &field) {
				if guarded == nil {
					guarded = make(map[string]bool)
				}
				guarded[field] = true
			}
			forClause, ok := clause.(*ast.ForClause)
			if !ok || !isParameterDot(forClause.Source, &field) || guarded[field] {
				c.Clauses[i] = astutil.Apply(clause, visit, nil).(ast.Clause)
				continue
			}
			forClause.Source = &ast.BinaryExpr{
				X: &ast.UnaryExpr{
					Op: token.MUL,
					X:  forClause.Source,
				},
				Op: token.OR,
				Y:  &ast.StructLit{},
			}
		}
	}
}

// isGuard reports whether x is a clause of the form
//
//	if parameter.foo != _|_
//
// If field is non-nil and it returns true, it fills in *field
// with the name of the parameter field (foo above).
func isGuard(x0 ast.Clause, field *string) bool {
	x, ok := x0.(*ast.IfClause)
	if !ok {
		return false
	}
	e, ok := x.Condition.(*ast.BinaryExpr)
	if !ok || e.Op != token.NEQ || !is[*ast.BottomLit](e.Y) || !isParameterDot(e.X, field) {
		return false
	}
	return true
}

func is[T any](x any) bool {
	_, ok := x.(T)
	return ok
}

// isParameterDot reports whether x represents an expression
// such as:
//
//	parameter.foo
//	parameter.foo.bar
//	parameter["foo"]
//
// If field is non-nil and it returns true, it fills in *field
// with the name of the parameter field (foo above).
func isParameterDot(x ast.Expr, field *string) bool {
	var lhs ast.Expr
	var rhs ast.Node
	switch x := x.(type) {
	case *ast.SelectorExpr:
		lhs = x.X
		rhs = x.Sel
	case *ast.IndexExpr:
		lhs = x.X
		rhs = x.Index
	default:
		return false
	}
	switch lhs := lhs.(type) {
	case *ast.Ident:
		if lhs.Name == "parameter" {
			if field != nil {
				switch rhs := rhs.(type) {
				case *ast.Ident:
					*field = rhs.Name
				case *ast.BasicLit:
					if rhs.Kind == token.STRING {
						*field, _ = literal.Unquote(rhs.Value)
					}
				}
			}
			return true
		}
	case *ast.SelectorExpr, *ast.IndexExpr:
		return isParameterDot(lhs, field)
	}
	return false
}
