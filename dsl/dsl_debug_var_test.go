package dsl

import (
	"testing"

	"github.com/bornholm/go-fuzzy"
)

// TestDebugVariableDefinitionsWithComments is a debug version of TestVariableDefinitionsWithComments
// that will help us identify which variable is undefined
func TestDebugVariableDefinitionsWithComments(t *testing.T) {
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

	// Parse the DSL
	result, err := ParseRulesAndVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse variables with comments: %v", err)
	}

	// Check that we got the expected number of variables and rules
	if len(result.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result.Variables))
	}
	if len(result.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(result.Rules))
	}

	// Print out each parsed rule to see what variables it references
	t.Logf("RULES:")
	for i, rule := range result.Rules {
		t.Logf("Rule %d: %v", i+1, rule)
	}

	// Verify variable names
	t.Logf("VARIABLES FROM DSL:")
	for i, v := range result.Variables {
		t.Logf("Variable %d: %s", i+1, v.Name())
		// Print the terms in each variable
		for j, term := range v.Terms() {
			t.Logf("  Term %d: %s", j+1, term.Name())
		}
	}

	// Create engine
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))

	// Add the variables from the DSL
	engine.Variables(result.Variables...)

	// Add the missing ac_mode variable that's not defined in the DSL but used in rules
	engine.Variables(
		fuzzy.NewVariable(
			"ac_mode",
			fuzzy.NewTerm("heating", fuzzy.Linear(0, 100)),
			fuzzy.NewTerm("off", fuzzy.Triangular(-50, 0, 50)),
			fuzzy.NewTerm("cooling", fuzzy.Inverted(fuzzy.Linear(-100, 0))),
		),
	)

	// Print out what variables we're adding to the engine
	t.Logf("ENGINE VARIABLES ADDED:")
	t.Logf("- temperature (from DSL)")
	t.Logf("- humidity (from DSL)")
	t.Logf("- ac_mode (added manually)")

	// Add the rules
	engine.Rules(result.Rules...)

	// Test inference with hot temperature and humidity
	inputs := fuzzy.Values{
		"temperature": 25, // hot
		"humidity":    80, // high
	}

	t.Logf("INPUTS: %v", inputs)

	// Try inference and catch the error
	results, err := engine.Infer(inputs)
	if err != nil {
		t.Logf("INFERENCE ERROR: %v", err)
		// Continue so we can see debug output
	} else {
		acMode, _ := results.Best("ac_mode")
		t.Logf("Best ac_mode: %s with truth degree %f", acMode.Term(), acMode.TruthDegree())
	}
}
