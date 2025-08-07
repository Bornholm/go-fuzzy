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

func ExampleParseRules() {
	// Define a set of rules using the DSL
	dsl := `
	IF temperature IS cold THEN ac_mode IS heating;
	IF temperature IS comfortable THEN ac_mode IS off;
	IF temperature IS hot THEN ac_mode IS cooling;
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
