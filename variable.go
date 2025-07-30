package fuzzy

import (
	"math"

	"github.com/pkg/errors"
)

type Variable struct {
	name  string
	terms map[string]*Term

	universeMin float64
	universeMax float64
}

func (v *Variable) Name() string {
	return v.name
}

func (v *Variable) Term(name string) (*Term, error) {
	t, exists := v.terms[name]
	if !exists {
		return nil, errors.WithStack(ErrUndefinedTerm)
	}

	return t, nil
}

func (v *Variable) UniverseMin() float64 {
	return v.universeMin
}

func (v *Variable) UniverseMax() float64 {
	return v.universeMax
}

func NewVariable(name string, terms ...*Term) *Variable {
	indexedTerms := make(map[string]*Term, len(terms))
	universeMin := math.Inf(1)
	universeMax := math.Inf(-1)

	for _, t := range terms {
		if _, exists := indexedTerms[t.Name()]; exists {
			panic(errors.WithStack(ErrTermAlreadyExists))
		}

		indexedTerms[t.Name()] = t
		min, max := t.Domain()
		universeMin = math.Min(universeMin, min)
		universeMax = math.Max(universeMax, max)
	}

	return &Variable{
		name:        name,
		terms:       indexedTerms,
		universeMin: universeMin,
		universeMax: universeMax,
	}
}

type Term struct {
	name       string
	membership Membership
}

func (t *Term) Name() string {
	return t.name
}

func (t *Term) Membership() Membership {
	return t.membership
}

func (t *Term) Domain() (float64, float64) {
	return t.membership.Domain()
}

func NewTerm(name string, membership Membership) *Term {
	return &Term{
		name:       name,
		membership: membership,
	}
}
