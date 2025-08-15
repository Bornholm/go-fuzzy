package dsl

import (
	"testing"

	"github.com/bornholm/go-fuzzy"
)

// TestDebugParseWithComments is a debug version of TestParseWithComments that
// will help us identify which rule is causing the inference failure
func TestDebugParseWithComments(t *testing.T) {
	// Same DSL as in TestParseWithComments
	dsl := `
	// This is a single-line comment at the beginning of a line
	IF temperature IS cold THEN ac_mode IS heating; // Single-line comment at the end of a rule

	/* Multi-line comment spanning
	   multiple lines with indentation
	   should be completely ignored */
	IF temperature IS comfortable /* inline multi-line comment */ THEN ac_mode IS off;

	// The following rule has comments embedded within it
	IF temperature // Comment after IF
	   IS // Comment after IS
	   hot // Comment after variable
	   THEN // Comment after THEN
	   ac_mode // Comment after variable
	   IS // Comment after IS
	   cooling; // Comment after term

	// Rule with more complex comments
	IF /* comment */ (temperature IS cold OR /* nested comment */ humidity IS high) 
	   AND /* another comment */ NOT pressure IS low 
	   THEN /* final comment */ ac_mode IS heating;
	`

	// Parse the rules
	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rules with comments: %v", err)
	}

	// Should have 4 rules despite all the comments
	if len(rules) != 4 {
		t.Fatalf("Expected 4 rules, got %d", len(rules))
	}

	// Print each rule to see what was parsed
	for i, rule := range rules {
		t.Logf("Rule %d: %v", i+1, rule)
	}

	// Verify the engine can process these rules individually
	for i, rule := range rules {
		engine := fuzzy.NewEngine(fuzzy.Centroid(100))
		setupTestEngine(engine)
		engine.Rules(rule)

		// Test inference with appropriate input for each rule
		var inputs fuzzy.Values
		if i == 0 || i == 3 {
			// For first and fourth rule (cold temperature)
			inputs = fuzzy.Values{
				"temperature": 0, // cold
			}
			if i == 3 {
				// For fourth rule, also need humidity and pressure
				inputs["humidity"] = 80  // high
				inputs["pressure"] = 100 // not low
			}
		} else if i == 1 {
			// For second rule (comfortable temperature)
			inputs = fuzzy.Values{
				"temperature": 20, // comfortable
			}
		} else if i == 2 {
			// For third rule (hot temperature)
			inputs = fuzzy.Values{
				"temperature": 25, // hot
			}
		}

		// Log the inputs we're using
		t.Logf("Rule %d testing with inputs: %v", i+1, inputs)

		// Run inference
		results, err := engine.Infer(inputs)
		if err != nil {
			t.Fatalf("Rule %d: Inference failed: %v", i+1, err)
		}

		// The ac_mode should have a value
		acMode, _ := results.Best("ac_mode")
		t.Logf("Rule %d: Best ac_mode: %s with truth degree %f",
			i+1, acMode.Term(), acMode.TruthDegree())
	}

	// Now test all rules together (this should pass if individual rules work)
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test inference with cold temperature AND providing all required variables
	inputs := fuzzy.Values{
		"temperature": 0,   // cold
		"humidity":    80,  // high
		"pressure":    100, // not low
	}
	t.Logf("Testing all rules with inputs: %v", inputs)

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("All rules: Inference failed: %v", err)
	}

	// The ac_mode should be heating
	acMode, _ := results.Best("ac_mode")
	t.Logf("All rules: Best ac_mode: %s with truth degree %f",
		acMode.Term(), acMode.TruthDegree())
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
}
