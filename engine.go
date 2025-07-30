package fuzzy

import "github.com/pkg/errors"

type Values map[string]float64

type DefuzzifyFunc func(m Membership, min, max float64) float64

type Engine struct {
	rules     []*Rule
	variables []*Variable
	defuzzify DefuzzifyFunc
}

func (e *Engine) Infer(values Values) (Results, error) {
	ctx := NewContext(e.variables, values)

	for _, r := range e.rules {
		outputVariableName := r.conclusion.Variable()
		outputTermName := r.conclusion.Term()

		outputVariable, err := ctx.Variable(outputVariableName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		outputTerm, err := outputVariable.Term(outputTermName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		truthDegree, err := r.premise.Value(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		ctx.AddResult(outputVariableName, outputTerm, truthDegree)
	}

	return ctx.Results(), nil
}

func (e *Engine) Defuzzify(variableName string, results Results) (float64, error) {
	var targetVariable *Variable
	for _, v := range e.variables {
		if v.Name() == variableName {
			targetVariable = v
			break
		}
	}

	if targetVariable == nil {
		return 0, errors.WithStack(ErrUndefinedVariable)
	}

	variableResults, ok := results[variableName]
	if !ok || len(variableResults) == 0 {
		return (targetVariable.UniverseMin() + targetVariable.UniverseMax()) / 2, nil
	}

	finalMembership := Max()
	for _, res := range variableResults {
		finalMembership.memberships = append(finalMembership.memberships, res.Membership())
	}

	return e.defuzzify(finalMembership, targetVariable.UniverseMin(), targetVariable.UniverseMax()), nil
}

func (e *Engine) Variables(variables ...*Variable) *Engine {
	e.variables = variables
	return e
}

func (e *Engine) Rules(rules ...*Rule) *Engine {
	e.rules = rules
	return e
}

func NewEngine(defuzzify DefuzzifyFunc) *Engine {
	if defuzzify == nil {
		defuzzify = Centroid(1000)
	}
	return &Engine{
		defuzzify: defuzzify,
	}
}
