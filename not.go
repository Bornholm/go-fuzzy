package fuzzy

import (
	"github.com/pkg/errors"
)

type NotExpr struct {
	expr Expr
}

func (e *NotExpr) Value(ctx *Context) (float64, error) {
	v, err := e.expr.Value(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return 1 - v, nil
}

func Not(m Expr) *NotExpr {
	return &NotExpr{m}
}
