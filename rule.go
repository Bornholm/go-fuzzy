package fuzzy

type Rule struct {
	premise    Expr
	conclusion *IsExpr
}

func (r *Rule) Premise() Expr {
	return r.premise
}

func (r *Rule) Conclusion() *IsExpr {
	return r.conclusion
}

func (r *Rule) Then(variable string, term string) *Rule {
	r.conclusion = Set(variable, term)

	return r
}

func NewRule(premise Expr, conclusion *IsExpr) *Rule {
	return &Rule{premise, conclusion}
}

func If(expr Expr) *Rule {
	return &Rule{
		premise: expr,
	}
}
