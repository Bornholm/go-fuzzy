package fuzzy

type Results map[string]map[string]Result

func (r Results) Best(variable string) Result {
	var best Result

	for _, res := range r[variable] {
		if res.TruthDegree() > best.TruthDegree() {
			best = res
		}
	}

	return best
}

type Result struct {
	term        string
	truthDegree float64
	membership  Membership
}

func (r *Result) Term() string {
	return r.term
}

func (r *Result) TruthDegree() float64 {
	return r.truthDegree
}

func (r *Result) Membership() Membership {
	return r.membership
}
