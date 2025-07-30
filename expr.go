package fuzzy

type Expr interface {
	Value(ctx *Context) (float64, error)
}
