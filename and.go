package fuzzy

import (
	"math"

	"github.com/pkg/errors"
)

type AndExpr struct {
	exprs []Expr
}

func (e *AndExpr) Value(ctx *Context) (float64, error) {
	min := math.Inf(1) // Initialize to positive infinity

	for _, m := range e.exprs {
		v, err := m.Value(ctx)
		if err != nil {
			return 0, errors.WithStack(err)
		}

		min = math.Min(min, v)
	}

	return min, nil
}

func (e *AndExpr) Exprs() []Expr {
	return e.exprs
}

func And(expr ...Expr) *AndExpr {
	if len(expr) == 0 {
		panic(errors.WithStack(ErrMissingArguments))
	}

	return &AndExpr{expr}
}
