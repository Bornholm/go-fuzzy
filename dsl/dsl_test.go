package dsl

import (
	"fmt"
	"testing"

	"github.com/bornholm/go-fuzzy"
)

func TestParseSingleRule(t *testing.T) {
	dsl := "IF temperature IS cold THEN ac_mode IS heating;"

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rule: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}
}

func TestParseMultipleRules(t *testing.T) {
	dsl := `
	IF temperature IS cold THEN ac_mode IS heating;
	IF temperature IS comfortable THEN ac_mode IS off;
	IF temperature IS hot THEN ac_mode IS cooling;
	`

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rules: %v", err)
	}

	if len(rules) != 3 {
		t.Fatalf("Expected 3 rules, got %d", len(rules))
	}
}

func TestParseRuleWithAnd(t *testing.T) {
	dsl := "IF temperature IS cold AND humidity IS high THEN ac_mode IS heating;"

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rule: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	// Verify that the engine can process this rule
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test with input values
	inputs := fuzzy.Values{
		"temperature": 0,  // cold
		"humidity":    80, // high
	}

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// The ac_mode should be heating with high truth degree
	acMode := results.Best("ac_mode")
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for heating, got %f", acMode.TruthDegree())
	}
}

func TestParseRuleWithOr(t *testing.T) {
	dsl := "IF temperature IS cold OR humidity IS low THEN ac_mode IS heating;"

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rule: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	// Verify that the engine can process this rule
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test with input values
	inputs := fuzzy.Values{
		"temperature": 20, // not cold
		"humidity":    20, // low
	}

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// The ac_mode should still be heating due to low humidity
	acMode := results.Best("ac_mode")
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for heating, got %f", acMode.TruthDegree())
	}
}

func TestParseRuleWithNot(t *testing.T) {
	dsl := "IF NOT temperature IS hot THEN ac_mode IS heating;"

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rule: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	// Verify that the engine can process this rule
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test with input values
	inputs := fuzzy.Values{
		"temperature": 15, // not hot
	}

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// The ac_mode should be heating with high truth degree
	acMode := results.Best("ac_mode")
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for heating, got %f", acMode.TruthDegree())
	}
}

// TestParseRuleWithSimpleNot tests a simple NOT expression to help debug
func TestParseRuleWithSimpleNot(t *testing.T) {
	dsl := "IF NOT pressure IS low THEN ac_mode IS heating;"

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rule: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	// Verify that the engine can process this rule
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test with input values
	inputs := fuzzy.Values{
		"pressure": 100, // not low
	}

	// Debug: Log input values
	t.Logf("Inputs: pressure=%v", inputs["pressure"])

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// Debug: Print results
	debugAcMode := results.Best("ac_mode")
	t.Logf("Best ac_mode: %s with truth degree %f",
		debugAcMode.Term(), debugAcMode.TruthDegree())

	// The ac_mode should be heating with high truth degree
	acMode := results.Best("ac_mode")
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for heating, got %f", acMode.TruthDegree())
	}
}

func TestParseRuleWithParentheses(t *testing.T) {
	dsl := "IF (temperature IS cold OR humidity IS high) AND NOT pressure IS low THEN ac_mode IS heating;"

	rules, err := ParseRules(dsl)
	if err != nil {
		t.Fatalf("Failed to parse rule: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	// Verify that the engine can process this rule
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test with input values
	inputs := fuzzy.Values{
		"temperature": 0,   // cold
		"humidity":    50,  // medium
		"pressure":    100, // not low
	}

	// Debug: Log input values
	t.Logf("Inputs: temperature=%v, humidity=%v, pressure=%v",
		inputs["temperature"], inputs["humidity"], inputs["pressure"])

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// Debug: Print results
	debugAcMode := results.Best("ac_mode")
	t.Logf("Best ac_mode: %s with truth degree %f",
		debugAcMode.Term(), debugAcMode.TruthDegree())

	// The ac_mode should be heating with high truth degree
	acMode := results.Best("ac_mode")
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for heating, got %f", acMode.TruthDegree())
	}
}

func TestParseRulesOrPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ParseRulesOrPanic panicked: %v", r)
		}
	}()

	dsl := "IF temperature IS cold THEN ac_mode IS heating;"
	rules := ParseRulesOrPanic(dsl)

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}
}

func TestParseInvalidRule(t *testing.T) {
	testCases := []struct {
		name    string
		dsl     string
		wantErr bool
	}{
		{
			name:    "Missing IF",
			dsl:     "temperature IS cold THEN ac_mode IS heating;",
			wantErr: true,
		},
		{
			name:    "Missing THEN",
			dsl:     "IF temperature IS cold ac_mode IS heating;",
			wantErr: true,
		},
		{
			name:    "Missing semicolon",
			dsl:     "IF temperature IS cold THEN ac_mode IS heating",
			wantErr: true,
		},
		{
			name:    "Missing variable after IF",
			dsl:     "IF THEN ac_mode IS heating;",
			wantErr: true,
		},
		{
			name:    "Missing term after IS",
			dsl:     "IF temperature IS THEN ac_mode IS heating;",
			wantErr: true,
		},
		{
			name:    "Unbalanced parentheses",
			dsl:     "IF (temperature IS cold AND humidity IS high THEN ac_mode IS heating;",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseRules(tc.dsl)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseRules() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func setupTestEngine(engine *fuzzy.Engine) {
	// Define variables and terms for testing
	// Create variables with more explicit membership functions
	// Especially for pressure.low to make NOT pressure IS low evaluate correctly
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
			// Make "low" clearly defined: 1.0 at 900, linear down to 0.0 at 1000
			// This means pressure=100 should definitely NOT be "low"
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
}

func TestParseWithComments(t *testing.T) {
	// DSL text with various types of comments
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

	// Verify the engine can process these rules
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)
	engine.Rules(rules...)

	// Test inference with all required inputs for all rules
	inputs := fuzzy.Values{
		"temperature": 0,   // cold
		"humidity":    80,  // high
		"pressure":    100, // not low
	}
	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// The ac_mode should be heating
	acMode := results.Best("ac_mode")
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
}

func TestVariableDefinitionsWithComments(t *testing.T) {
	// DSL text with variable definitions and comments
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

	// Verify variable names
	if result.Variables[0].Name() != "temperature" {
		t.Errorf("Expected first variable to be 'temperature', got '%s'", result.Variables[0].Name())
	}
	if result.Variables[1].Name() != "humidity" {
		t.Errorf("Expected second variable to be 'humidity', got '%s'", result.Variables[1].Name())
	}

	// Verify terms in the temperature variable
	temperature := result.Variables[0]
	if term, err := temperature.Term("cold"); err != nil {
		t.Errorf("Term 'cold' not found in temperature variable: %v", err)
	} else {
		// Verify the term has a membership function
		if term.Membership() == nil {
			t.Errorf("Term 'cold' has no membership function")
		}
	}
	if term, err := temperature.Term("hot"); err != nil {
		t.Errorf("Term 'hot' not found in temperature variable: %v", err)
	} else {
		// Verify the term has a membership function
		if term.Membership() == nil {
			t.Errorf("Term 'hot' has no membership function")
		}
	}

	// Create engine and set up variables using setupTestEngine
	// This ensures consistent variable definitions across all tests
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	setupTestEngine(engine)

	// Add the rules
	engine.Rules(result.Rules...)

	// Test inference with hot temperature and humidity
	inputs := fuzzy.Values{
		"temperature": 25, // hot
		"humidity":    80, // high
	}
	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// The ac_mode should be cooling due to hot temperature
	acMode := results.Best("ac_mode")
	if acMode.Term() != "cooling" {
		t.Errorf("Expected ac_mode to be cooling, got %s", acMode.Term())
	}
}

func TestEdgeCaseComments(t *testing.T) {
	// Test edge cases with comments
	dslCases := []struct {
		name     string
		dsl      string
		expected int // Expected number of rules
	}{
		{
			name: "Comments at start and end of file",
			dsl: `// Starting comment
			// Another starting comment
			IF temperature IS cold THEN ac_mode IS heating;
			// Ending comment`,
			expected: 1,
		},
		{
			name: "Multi-line comment at start and end",
			dsl: `/* 
			Starting multi-line comment
			with multiple lines
			*/
			IF temperature IS cold THEN ac_mode IS heating;
			/* Ending multi-line comment */`,
			expected: 1,
		},
		{
			name: "Only comments, no rules",
			dsl: `// Just a comment
			/* Another comment
			   spanning multiple lines */
			// Final comment`,
			expected: 0,
		},
		{
			name: "Comments between rules",
			dsl: `IF temperature IS cold THEN ac_mode IS heating;
			// Comment between rules
			/* Multi-line comment
			   between rules */
			IF temperature IS hot THEN ac_mode IS cooling;`,
			expected: 2,
		},
		{
			name: "Comment markers within string literals",
			// This tests that comment markers inside variable/term names are not treated as comments
			dsl:      `IF temperature IS cold THEN comment_marker IS not_a_comment;`,
			expected: 1,
		},
		{
			name: "Nested comment-like markers",
			dsl: `// This comment contains /* which should not start a nested comment
			IF temperature IS cold THEN ac_mode IS heating;`,
			expected: 1,
		},
		{
			name: "Comment with 'escaped' markers",
			dsl: `/* Comment with // inside it */
			IF temperature IS cold THEN ac_mode IS heating;`,
			expected: 1,
		},
	}

	for _, tc := range dslCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := ParseRules(tc.dsl)
			if err != nil {
				t.Fatalf("Failed to parse rules: %v", err)
			}

			if len(rules) != tc.expected {
				t.Errorf("Expected %d rules, got %d", tc.expected, len(rules))
			}
		})
	}
}

func ExampleParseRules() {
	// Define a set of rules using the DSL
	dsl := `
	// Define temperature rules
	IF temperature IS cold THEN ac_mode IS heating;
	IF temperature IS comfortable THEN ac_mode IS off;
	IF temperature IS hot THEN ac_mode IS cooling; // Hot temperature needs cooling
	`

	// Parse the rules
	rules, err := ParseRules(dsl)
	if err != nil {
		panic(err)
	}

	// Create a new inference engine
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))

	// Define variables and terms
	engine.Variables(
		fuzzy.NewVariable(
			"temperature",
			fuzzy.NewTerm("cold", fuzzy.Inverted(fuzzy.Linear(-10, 10))),
			fuzzy.NewTerm("comfortable", fuzzy.Trapezoid(5, 18, 22, 25)),
			fuzzy.NewTerm("hot", fuzzy.Linear(20, 30)),
		),
		fuzzy.NewVariable(
			"ac_mode",
			fuzzy.NewTerm("heating", fuzzy.Linear(0, 100)),
			fuzzy.NewTerm("off", fuzzy.Triangular(-50, 0, 50)),
			fuzzy.NewTerm("cooling", fuzzy.Inverted(fuzzy.Linear(-100, 0))),
		),
	)

	// Add the parsed rules to the engine
	engine.Rules(rules...)

	// Process input values
	inputs := fuzzy.Values{
		"temperature": 30,
	}

	// Run inference
	results, err := engine.Infer(inputs)
	if err != nil {
		panic(err)
	}

	// Get the best matching term
	bestMatch := results.Best("ac_mode")
	fmt.Printf("AC Mode: %s (truth degree: %.2f)\n", bestMatch.Term(), bestMatch.TruthDegree())
	// Output: AC Mode: cooling (truth degree: 1.00)
}
