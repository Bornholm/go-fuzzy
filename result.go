package fuzzy

import "sort"

type Results map[string]map[string]Result

func (r Results) Best(variable string) (*Result, bool) {
	var best *Result

	for _, res := range r[variable] {
		if best == nil || res.TruthDegree() > best.TruthDegree() {
			best = &res
		}
	}

	if best == nil || best.TruthDegree() == 0 {
		return nil, false
	}

	return best, true
}

func (r Results) Variables() []string {
	variables := make([]string, 0, len(r))
	for name := range r {
		variables = append(variables, name)
	}
	sort.Strings(variables)
	return variables
}

type Result struct {
	term        string
	truthDegree float64
	membership  Membership
}

func (r Result) Term() string {
	return r.term
}

func (r Result) TruthDegree() float64 {
	return r.truthDegree
}

func (r Result) Membership() Membership {
	return r.membership
}

func NewResult(term string, thruthDegree float64, membership Membership) Result {
	return Result{
		term:        term,
		truthDegree: thruthDegree,
		membership:  membership,
	}
}
