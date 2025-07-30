package fuzzy

import (
	"math"

	"github.com/pkg/errors"
)

type OrExpr struct {
	expr []Expr
}

func (e *OrExpr) Value(ctx *Context) (float64, error) {
	max := math.Inf(-1) // Initialize to negative infinity

	for _, m := range e.expr {
		v, err := m.Value(ctx)
		if err != nil {
			return 0, errors.WithStack(err)
		}

		max = math.Max(max, v)
	}

	return max, nil
}

func Or(expr ...Expr) *OrExpr {
	if len(expr) == 0 {
		panic(errors.WithStack(ErrMissingArguments))
	}

	return &OrExpr{expr}
}
