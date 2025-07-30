package fuzzy

import (
	"github.com/pkg/errors"
)

type IsExpr struct {
	variable string
	term     string
}

func (e *IsExpr) Variable() string {
	return e.variable
}

func (e *IsExpr) Term() string {
	return e.term
}

func (e *IsExpr) Value(ctx *Context) (float64, error) {
	variable, err := ctx.Variable(e.variable)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	term, err := variable.Term(e.term)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	value, err := ctx.Value(e.variable)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return term.Membership().Value(value), nil
}

func Is(variable string, term string) *IsExpr {
	return &IsExpr{variable, term}
}

var Set = Is
