package dsl

import (
	"testing"

	"github.com/bornholm/go-fuzzy"
)

// TestDebugManualSetup is a test that sets up the engine manually
// to help identify what variable might be missing
func TestDebugManualSetup(t *testing.T) {
	// Same DSL as in TestVariableDefinitionsWithComments
	dsl := `
	// Define temperature variable with comments
	DEFINE temperature ( // Comment after opening parenthesis
		// Comment before term definition
		TERM cold LINEAR (-10, 10), /* Multi-line comment after term */
		/* Comment before term */ TERM hot LINEAR (20, 30) // Comment after term
	); // Comment after closing definition

	// Define humidity with multi-line comments
	DEFINE humidity (
		/* Multi-line comment
		   spanning multiple lines */
		TERM low INVERTED(LINEAR(0, 50)),
		TERM high LINEAR /* inline comment */ (50, 100)
	);

	// Now some rules using the defined variables
	IF temperature IS cold THEN ac_mode IS heating; // Use the variables
	IF humidity IS high THEN ac_mode IS cooling; // Another rule
	`

	// Parse the DSL to get the rules
	result, err := ParseRulesAndVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse variables with comments: %v", err)
	}

	// Check that we got the expected number of rules
	if len(result.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(result.Rules))
	}

	// Print out each parsed rule to see what variables it references
	t.Logf("RULES:")
	for i, rule := range result.Rules {
		t.Logf("Rule %d: %v", i+1, rule)
	}

	// Set up the engine completely manually to match setupTestEngine
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))

	// Add all standard variables used in the tests (from setupTestEngine)
	engine.Variables(
		fuzzy.NewVariable(
			"temperature",
			fuzzy.NewTerm("cold", fuzzy.Inverted(fuzzy.Linear(-10, 10))),
			fuzzy.NewTerm("comfortable", fuzzy.Trapezoid(5, 18, 22, 25)),
			fuzzy.NewTerm("hot", fuzzy.Linear(20, 30)),
		),
		fuzzy.NewVariable(
			"humidity",
			fuzzy.NewTerm("low", fuzzy.Inverted(fuzzy.Linear(0, 50))),
			fuzzy.NewTerm("medium", fuzzy.Triangular(30, 50, 70)),
			fuzzy.NewTerm("high", fuzzy.Linear(50, 100)),
		),
		fuzzy.NewVariable(
			"pressure",
			fuzzy.NewTerm("low", fuzzy.Linear(1000, 900)),
			fuzzy.NewTerm("normal", fuzzy.Triangular(950, 1013, 1050)),
			fuzzy.NewTerm("high", fuzzy.Linear(1020, 1100)),
		),
		fuzzy.NewVariable(
			"ac_mode",
			fuzzy.NewTerm("heating", fuzzy.Linear(0, 100)),
			fuzzy.NewTerm("off", fuzzy.Triangular(-50, 0, 50)),
			fuzzy.NewTerm("cooling", fuzzy.Inverted(fuzzy.Linear(-100, 0))),
		),
	)

	// Print all the variables we've added to the engine
	t.Logf("ENGINE VARIABLES ADDED:")
	t.Logf("- temperature")
	t.Logf("- humidity")
	t.Logf("- pressure")
	t.Logf("- ac_mode")

	// Add just the rules from the DSL
	engine.Rules(result.Rules...)

	// Create inputs with all possible variables
	inputs := fuzzy.Values{
		"temperature": 25,  // hot
		"humidity":    80,  // high
		"pressure":    100, // not low (in case it's needed)
	}

	t.Logf("INPUTS: %v", inputs)

	// Try inference
	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// Should get cooling due to humidity IS high rule
	acMode, _ := results.Best("ac_mode")
	t.Logf("Best ac_mode: %s with truth degree %f", acMode.Term(), acMode.TruthDegree())
}
