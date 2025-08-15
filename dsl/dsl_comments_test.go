package dsl

import (
	"testing"

	"github.com/bornholm/go-fuzzy"
)

func TestCommentHandling(t *testing.T) {
	// Test single-line comments
	t.Run("SingleLineComments", func(t *testing.T) {
		dsl := `
		// This is a single-line comment
		IF temperature IS cold THEN ac_mode IS heating; // Comment at end of line
		// Another comment
		`

		rules, err := ParseRules(dsl)
		if err != nil {
			t.Fatalf("Failed to parse rules with single-line comments: %v", err)
		}

		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule, got %d", len(rules))
		}
	})

	// Test multi-line comments
	t.Run("MultiLineComments", func(t *testing.T) {
		dsl := `
		/* This is a multi-line comment
		   spanning multiple lines */
		IF temperature IS cold THEN ac_mode IS heating;
		`

		rules, err := ParseRules(dsl)
		if err != nil {
			t.Fatalf("Failed to parse rules with multi-line comments: %v", err)
		}

		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule, got %d", len(rules))
		}
	})

	// Test inline multi-line comments
	t.Run("InlineMultiLineComments", func(t *testing.T) {
		dsl := `
		IF temperature /* comment */ IS /* another */ cold THEN ac_mode IS heating;
		`

		rules, err := ParseRules(dsl)
		if err != nil {
			t.Fatalf("Failed to parse rules with inline multi-line comments: %v", err)
		}

		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule, got %d", len(rules))
		}
	})

	// Test mixed comment types
	t.Run("MixedCommentTypes", func(t *testing.T) {
		dsl := `
		// Single-line comment
		/* Multi-line comment */
		IF /* comment */ temperature IS cold // Comment
		   THEN /* comment */ ac_mode IS heating; // Comment
		`

		rules, err := ParseRules(dsl)
		if err != nil {
			t.Fatalf("Failed to parse rules with mixed comment types: %v", err)
		}

		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule, got %d", len(rules))
		}
	})

	// Test inference with comments
	t.Run("InferenceWithComments", func(t *testing.T) {
		dsl := `
		// This rule defines what happens when temperature is cold
		IF temperature IS cold THEN ac_mode IS heating;
		`

		// Parse the rules
		rules, err := ParseRules(dsl)
		if err != nil {
			t.Fatalf("Failed to parse rules with comments: %v", err)
		}

		if len(rules) != 1 {
			t.Fatalf("Expected 1 rule, got %d", len(rules))
		}

		// Setup the engine
		engine := fuzzy.NewEngine(fuzzy.Centroid(100))

		// Define variables
		engine.Variables(
			fuzzy.NewVariable(
				"temperature",
				fuzzy.NewTerm("cold", fuzzy.Inverted(fuzzy.Linear(-10, 10))),
				fuzzy.NewTerm("hot", fuzzy.Linear(20, 30)),
			),
			fuzzy.NewVariable(
				"ac_mode",
				fuzzy.NewTerm("heating", fuzzy.Linear(0, 100)),
				fuzzy.NewTerm("cooling", fuzzy.Inverted(fuzzy.Linear(-100, 0))),
			),
		)

		// Add rules
		engine.Rules(rules...)

		// Test inference
		inputs := fuzzy.Values{
			"temperature": 0, // cold
		}

		// Log the inputs for debugging
		t.Logf("Inference inputs: %v", inputs)

		// Run inference
		results, err := engine.Infer(inputs)
		if err != nil {
			t.Fatalf("Inference failed: %v", err)
		}

		// Check the results
		acMode, _ := results.Best("ac_mode")
		t.Logf("Best ac_mode: %s with truth degree %f", acMode.Term(), acMode.TruthDegree())

		if acMode.Term() != "heating" {
			t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
		}
	})
}
