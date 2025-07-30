package fuzzy

import (
	"slices"
	"sort"
	"testing"

	"github.com/pkg/errors"
)

type testCase struct {
	Temperature                      float64
	ExpectedAirConditionning         string
	ExpectedMinValueThreshold        float64
	ExpectedMinThruthDegreeThreshold float64
}

var testCases = []testCase{
	{
		Temperature:                      20,
		ExpectedAirConditionning:         "stopped",
		ExpectedMinThruthDegreeThreshold: 1,
	},
	{
		Temperature:                      15,
		ExpectedAirConditionning:         "stopped",
		ExpectedMinThruthDegreeThreshold: 1,
	},
	{
		Temperature:                      25,
		ExpectedAirConditionning:         "stopped",
		ExpectedMinThruthDegreeThreshold: 1,
	},
	{
		Temperature:                      -10,
		ExpectedAirConditionning:         "heating",
		ExpectedMinThruthDegreeThreshold: 1,
	},
	{
		Temperature:                      30,
		ExpectedAirConditionning:         "cooling",
		ExpectedMinThruthDegreeThreshold: 1,
	},
	{
		Temperature:                      50,
		ExpectedAirConditionning:         "cooling",
		ExpectedMinThruthDegreeThreshold: 1,
	},
}

func TestEngine(t *testing.T) {
	engine := NewEngine(MeanOfMaximum(1000))

	engine.Variables(
		NewVariable(
			"temperature",
			// La température est "cold" (froide) en dessous de 10°
			// On considère "cold" particulièrement vrai
			// si la température descend en dessous de -10°
			NewTerm("cold", Inverted(Linear(-10, 10))),

			// La température est "cool" (fraiche) entre 0° et 20°
			// On considère "cool" particulièrement vrai
			// lorsque la temperature == 10°
			NewTerm("cool", Triangular(0, 10, 20)),

			// La température est "ok" (acceptable) entre 15° et 25°
			// mais elle est considérée parfaitement "ok" (valeur de 1.0)
			// dans la plage de 18° à 22°.
			NewTerm("ok", Trapezoid(15, 18, 22, 25)),

			// La température est "warm" (chaude) entre 20° et 30°
			// On considère "warm" particulièrement vrai
			// lorsque la temperature == 25°
			NewTerm("warm", Triangular(20, 25, 30)),

			// La température est "hot" (étouffante) à partir de 25°
			// On considère "hot" particulièrement vrai
			// si la température dépasse 30°
			NewTerm("hot", Linear(25, 30)),
		),
		NewVariable(
			"air-conditioning",
			// Soit une climatisation avec 3 états: "cooling", "stopped", "heating"
			// La puissance du "cooling" évolue de -100 (puissance max) à 0 (puissance nulle)
			// "stopped" n'est pas variable: il est actif ou pas
			// La puissance du "heating" évolue de 0 (puissance nulle) à 100 (puissance maximum)
			NewTerm("cooling", Inverted(Linear(-100, 0))),
			NewTerm("stopped", Triangular(-100, 0, 100)),
			NewTerm("heating", Linear(0, 100)),
		),
	)

	engine.Rules(
		If(
			Or(
				Is("temperature", "cold"),
				Is("temperature", "cool"),
			),
		).Then("air-conditioning", "heating"),
		If(
			Or(
				Is("temperature", "ok"),
				Is("temperature", "warm"),
			),
		).Then("air-conditioning", "stopped"),
		If(
			Is("temperature", "hot"),
		).Then("air-conditioning", "cooling"),
	)

	const outputVariable string = "air-conditioning"

	for _, tc := range testCases {
		inputs := Values{
			"temperature": tc.Temperature,
		}

		results, err := engine.Infer(inputs)
		if err != nil {
			t.Error(errors.WithStack(err))
		}

		dumpValues(t, inputs)
		t.Log("-----------------")

		dumpResult(t, engine, results, outputVariable)

		// outputResult := results[outputVariable]

		// value, err := engine.Defuzzify("air-conditioning", results)
		// if err != nil {
		// 	t.Errorf("%+v", errors.WithStack(err))
		// 	continue
		// }
	}

}

func dumpValues(t *testing.T, values Values) {
	t.Log("Values:")
	t.Log("|")

	keys := slices.Collect(func(yield func(v string) bool) {
		for k := range values {
			if !yield(k) {
				return
			}
		}
	})

	sort.Strings(keys)

	for _, k := range keys {
		t.Logf("|--> %s: %v", k, values[k])
		t.Log("|")
	}
}

func dumpResult(t *testing.T, engine *Engine, results Results, variable string) {
	t.Logf("Result: %s", variable)

	value, err := engine.Defuzzify(variable, results)
	if err != nil {
		panic(errors.WithStack(err))
	}

	t.Log("|")
	t.Logf("|-> Value: %f", value)

	variableResults := results[variable]

	keys := slices.Collect(func(yield func(v string) bool) {
		for k := range variableResults {
			if !yield(k) {
				return
			}
		}
	})

	sort.Strings(keys)

	for _, term := range keys {
		res := variableResults[term]
		t.Logf("|--> %s", term)
		t.Log("|    |")
		t.Logf("|    |--> TruthDegree: %f", res.TruthDegree())
		t.Log("|")
	}
}
