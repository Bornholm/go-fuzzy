package fuzzy_test

import (
	"fmt"

	"github.com/bornholm/go-fuzzy"
)

func ExampleEngine() {
	// Create a new inference engine using Centroid defuzzification
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))

	// Define our variables and their fuzzy terms
	engine.Variables(
		// Temperature variable with 3 terms
		fuzzy.NewVariable(
			"temperature",
			fuzzy.NewTerm("cold", fuzzy.Inverted(fuzzy.Linear(-10, 10))),
			fuzzy.NewTerm("comfortable", fuzzy.Trapezoid(5, 18, 22, 25)),
			fuzzy.NewTerm("hot", fuzzy.Linear(20, 30)),
		),

		// Output variable for AC mode
		fuzzy.NewVariable(
			"ac_mode",
			fuzzy.NewTerm("heating", fuzzy.Linear(0, 100)),
			fuzzy.NewTerm("off", fuzzy.Triangular(-50, 0, 50)),
			fuzzy.NewTerm("cooling", fuzzy.Inverted(fuzzy.Linear(-100, 0))),
		),
	)

	// Define our fuzzy rules
	engine.Rules(
		// If temperature is cold, then AC should be heating
		fuzzy.If(
			fuzzy.Is("temperature", "cold"),
		).Then("ac_mode", "heating"),

		// If temperature is comfortable, then AC should be off
		fuzzy.If(
			fuzzy.Is("temperature", "comfortable"),
		).Then("ac_mode", "off"),

		// If temperature is hot, then AC should be cooling
		fuzzy.If(
			fuzzy.Is("temperature", "hot"),
		).Then("ac_mode", "cooling"),
	)

	// Process input values (e.g., current temperature is 30°C)
	inputs := fuzzy.Values{
		"temperature": 30,
	}

	// Run inference
	results, err := engine.Infer(inputs)
	if err != nil {
		panic(err)
	}

	// Get defuzzified output value
	acMode, err := engine.Defuzzify("ac_mode", results)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Temperature: %.1f°C\n", inputs["temperature"])

	fmt.Printf("AC Mode value: %.2f\n", acMode)

	// Get the best matching term
	bestMatch := results.Best("ac_mode")
	fmt.Printf("AC Mode: %s (truth degree: %.2f)\n", bestMatch.Term(), bestMatch.TruthDegree())

	// Output: Temperature: 30.0°C
	// AC Mode value: -67.33
	// AC Mode: cooling (truth degree: 1.00)
}
