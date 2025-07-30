package fuzzy

import (
	"math"

	"github.com/pkg/errors"
)

type Context struct {
	variables map[string]*Variable
	inputs    map[string]float64
	results   map[string]map[string]Result
}

func (c *Context) Variable(name string) (*Variable, error) {
	v, exists := c.variables[name]
	if !exists {
		return nil, errors.WithStack(ErrUndefinedVariable)
	}

	return v, nil
}

func (c *Context) Value(variable string) (float64, error) {
	v, exists := c.inputs[variable]
	if !exists {
		return 0, errors.WithStack(ErrValueNotFound)
	}

	return v, nil
}

func (c *Context) AddResult(variable string, term *Term, truthDegree float64) {
	terms, exists := c.results[variable]
	if !exists {
		terms = make(map[string]Result)
	}

	result, exists := terms[term.Name()]
	if !exists {
		result = Result{
			term:        term.Name(),
			truthDegree: truthDegree,
		}
	}

	clippedMembership := Min(Constant(truthDegree), term.Membership())

	if result.membership != nil {
		result.membership = Max(result.Membership(), clippedMembership)
	} else {
		result.membership = clippedMembership
	}

	result.truthDegree = math.Max(result.truthDegree, truthDegree)
	terms[term.Name()] = result
	c.results[variable] = terms
}

func (c *Context) Result(variable string) map[string]Result {
	terms, exists := c.results[variable]
	if !exists {
		terms = make(map[string]Result, 1)
	}

	return terms
}

func (c *Context) Results() Results {
	return c.results
}

func NewContext(variables []*Variable, inputs map[string]float64) *Context {
	vars := make(map[string]*Variable, len(inputs))

	for _, v := range variables {
		if _, exists := vars[v.Name()]; exists {
			panic(errors.WithStack(ErrVariableAlreadyExists))
		}

		vars[v.Name()] = v
	}

	return &Context{
		variables: vars,
		inputs:    inputs,
		results:   make(map[string]map[string]Result),
	}
}
