package dsl

import (
	"math"
	"testing"

	"github.com/bornholm/go-fuzzy"
)

func TestParseVariableDefinition(t *testing.T) {
	dsl := `DEFINE temperature (
		TERM hot LINEAR (20, 30),
		TERM cold LINEAR (10, 0),
		TERM pleasant TRIANGULAR (10, 20, 25)
	);`

	variables, err := ParseVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse variable definition: %v", err)
	}

	if len(variables) != 1 {
		t.Fatalf("Expected 1 variable, got %d", len(variables))
	}

	temp := variables[0]
	if temp.Name() != "temperature" {
		t.Errorf("Expected variable name to be temperature, got %s", temp.Name())
	}

	// Check that the terms were parsed correctly
	hotTerm, err := temp.Term("hot")
	if err != nil {
		t.Fatalf("Term 'hot' not found: %v", err)
	}
	coldTerm, err := temp.Term("cold")
	if err != nil {
		t.Fatalf("Term 'cold' not found: %v", err)
	}
	pleasantTerm, err := temp.Term("pleasant")
	if err != nil {
		t.Fatalf("Term 'pleasant' not found: %v", err)
	}

	// Verify membership functions
	// Hot: LINEAR (20, 30)
	checkLinearMembership(t, hotTerm.Membership(), 20, 30)

	// Cold: LINEAR (10, 0) - For descending functions, this is implemented as Inverted(Linear(0, 10))
	checkDescendingLinearMembership(t, coldTerm.Membership(), 10, 0)

	// Pleasant: TRIANGULAR (10, 20, 25)
	checkTriangularMembership(t, pleasantTerm.Membership(), 10, 20, 25)
}

func TestParseMultipleVariableDefinitions(t *testing.T) {
	dsl := `
	DEFINE temperature (
		TERM hot LINEAR (20, 30),
		TERM cold LINEAR (10, 0)
	);
	
	DEFINE humidity (
		TERM dry LINEAR (0, 30),
		TERM comfortable TRIANGULAR (30, 50, 70),
		TERM wet LINEAR (70, 100)
	);`

	variables, err := ParseVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse variable definitions: %v", err)
	}

	if len(variables) != 2 {
		t.Fatalf("Expected 2 variables, got %d", len(variables))
	}

	// Check first variable (temperature)
	temp := variables[0]
	if temp.Name() != "temperature" {
		t.Errorf("Expected first variable name to be temperature, got %s", temp.Name())
	}
	if _, err := temp.Term("hot"); err != nil {
		t.Errorf("Term 'hot' not found in temperature: %v", err)
	}
	if _, err := temp.Term("cold"); err != nil {
		t.Errorf("Term 'cold' not found in temperature: %v", err)
	}

	// Check second variable (humidity)
	humidity := variables[1]
	if humidity.Name() != "humidity" {
		t.Errorf("Expected second variable name to be humidity, got %s", humidity.Name())
	}
	if _, err := humidity.Term("dry"); err != nil {
		t.Errorf("Term 'dry' not found in humidity: %v", err)
	}
	if _, err := humidity.Term("comfortable"); err != nil {
		t.Errorf("Term 'comfortable' not found in humidity: %v", err)
	}
	if _, err := humidity.Term("wet"); err != nil {
		t.Errorf("Term 'wet' not found in humidity: %v", err)
	}
}

func TestParseTrapezoidMembershipFunction(t *testing.T) {
	dsl := `DEFINE temperature (
		TERM comfortable TRAPEZOID (15, 20, 25, 30)
	);`

	variables, err := ParseVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse variable definition: %v", err)
	}

	temp := variables[0]
	comfortableTerm, err := temp.Term("comfortable")
	if err != nil {
		t.Fatalf("Term 'comfortable' not found: %v", err)
	}

	// Check trapezoid membership function
	checkTrapezoidMembership(t, comfortableTerm.Membership(), 15, 20, 25, 30)
}

func TestParseInvertedMembershipFunction(t *testing.T) {
	dsl := `DEFINE temperature (
		TERM not_hot INVERTED (LINEAR (20, 30))
	);`

	variables, err := ParseVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse variable definition: %v", err)
	}

	temp := variables[0]
	notHotTerm, err := temp.Term("not_hot")
	if err != nil {
		t.Fatalf("Term 'not_hot' not found: %v", err)
	}

	// Check that it behaves like an inverted linear function
	membership := notHotTerm.Membership()
	_, ok := membership.(*fuzzy.InvertedMembership)
	if !ok {
		t.Errorf("Expected InvertedMembership, got %T", membership)
	}

	// Check the behavior of the inverted function
	// For a LINEAR(20, 30), the value at 20 should be 0 and at 30 should be 1
	// When inverted, the value at 20 should be 1 and at 30 should be 0
	if !almostEqual(membership.Value(20), 1.0) {
		t.Errorf("Expected inverted value at 20 to be 1.0, got %f", membership.Value(20))
	}
	if !almostEqual(membership.Value(30), 0.0) {
		t.Errorf("Expected inverted value at 30 to be 0.0, got %f", membership.Value(30))
	}
	if !almostEqual(membership.Value(25), 0.5) {
		t.Errorf("Expected inverted value at 25 to be 0.5, got %f", membership.Value(25))
	}
}

func TestVariablesAndRulesCombined(t *testing.T) {
	dsl := `
	DEFINE temperature (
		TERM hot LINEAR (20, 30),
		TERM cold LINEAR (10, 0)
	);
	
	DEFINE ac_mode (
		TERM heating LINEAR (0, 100),
		TERM cooling LINEAR (100, 0)
	);
	
	IF temperature IS hot THEN ac_mode IS cooling;
	IF temperature IS cold THEN ac_mode IS heating;
	`

	// Parse both variables and rules
	result, err := ParseRulesAndVariables(dsl)
	if err != nil {
		t.Fatalf("Failed to parse DSL: %v", err)
	}

	// Check variables
	if len(result.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result.Variables))
	}

	// Check rules
	if len(result.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(result.Rules))
	}

	// Print debug info about variables and rules
	t.Logf("Variables: %d", len(result.Variables))
	for i, v := range result.Variables {
		t.Logf("Variable %d: %s", i, v.Name())
		// We don't have access to all terms, just check the expected ones
		if v.Name() == "temperature" {
			if term, err := v.Term("hot"); err == nil {
				t.Logf("  Has term: hot")
				t.Logf("  hot value at 25: %f", term.Membership().Value(25))
			}
			if term, err := v.Term("cold"); err == nil {
				t.Logf("  Has term: cold")
				t.Logf("  cold value at 5: %f", term.Membership().Value(5))
			}
		}
		if v.Name() == "ac_mode" {
			if _, err := v.Term("heating"); err == nil {
				t.Logf("  Has term: heating")
			}
			if _, err := v.Term("cooling"); err == nil {
				t.Logf("  Has term: cooling")
			}
		}
	}

	t.Logf("Rules: %d", len(result.Rules))
	for i, r := range result.Rules {
		t.Logf("Rule %d: %v", i, r)
	}

	// Create an engine with the parsed variables and rules
	engine := fuzzy.NewEngine(fuzzy.Centroid(100))
	engine.Variables(result.Variables...)
	engine.Rules(result.Rules...)

	// Test with hot temperature
	inputs := fuzzy.Values{
		"temperature": 25, // hot
	}

	// Debug the input values
	t.Logf("Testing with temperature: %v", inputs["temperature"])

	results, err := engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// Debug all results
	t.Logf("Results for ac_mode:")
	acModeResults := results["ac_mode"]
	for term, result := range acModeResults {
		t.Logf("  %s: %f", term, result.TruthDegree())
	}

	// The ac_mode should be cooling with high truth degree
	acMode, _ := results.Best("ac_mode")
	t.Logf("Best ac_mode: %s with truth degree %f",
		acMode.Term(), acMode.TruthDegree())
	if acMode.Term() != "cooling" {
		t.Errorf("Expected ac_mode to be cooling, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for cooling, got %f", acMode.TruthDegree())
	}

	// Test with cold temperature
	inputs = fuzzy.Values{
		"temperature": 5, // cold
	}

	// Debug the input values
	t.Logf("Testing with temperature: %v", inputs["temperature"])

	results, err = engine.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	// Debug all results
	t.Logf("Results for ac_mode:")
	acModeResults = results["ac_mode"]
	for term, result := range acModeResults {
		t.Logf("  %s: %f", term, result.TruthDegree())
	}

	// The ac_mode should be heating with high truth degree
	acMode, _ = results.Best("ac_mode")
	t.Logf("Best ac_mode: %s with truth degree %f",
		acMode.Term(), acMode.TruthDegree())
	if acMode.Term() != "heating" {
		t.Errorf("Expected ac_mode to be heating, got %s", acMode.Term())
	}
	if acMode.TruthDegree() < 0.5 {
		t.Errorf("Expected high truth degree for heating, got %f", acMode.TruthDegree())
	}
}

// Helper functions to check membership functions
func checkLinearMembership(t *testing.T, membership fuzzy.Membership, x1, x2 float64) {
	t.Helper()

	linearMembership, ok := membership.(*fuzzy.LinearMembership)
	if !ok {
		t.Fatalf("Expected LinearMembership, got %T", membership)
	}

	min, max := linearMembership.Domain()
	// The domain might be stored as provided (x1, x2) rather than min/max order
	if !((almostEqual(min, x1) && almostEqual(max, x2)) ||
		(almostEqual(min, math.Min(x1, x2)) && almostEqual(max, math.Max(x1, x2)))) {
		t.Errorf("Expected domain to be either (%f, %f) or (%f, %f), got (%f, %f)",
			x1, x2, math.Min(x1, x2), math.Max(x1, x2), min, max)
	}

	// Check linear function behavior
	// If x1 < x2, then value at x1 should be 0 and value at x2 should be 1
	// If x1 > x2, then value at x1 should be 1 and value at x2 should be 0
	if x1 < x2 {
		if !almostEqual(linearMembership.Value(x1), 0) {
			t.Errorf("Expected value at %f to be 0, got %f", x1, linearMembership.Value(x1))
		}
		if !almostEqual(linearMembership.Value(x2), 1) {
			t.Errorf("Expected value at %f to be 1, got %f", x2, linearMembership.Value(x2))
		}
	} else {
		if !almostEqual(linearMembership.Value(x1), 1) {
			t.Errorf("Expected value at %f to be 1, got %f", x1, linearMembership.Value(x1))
		}
		if !almostEqual(linearMembership.Value(x2), 0) {
			t.Errorf("Expected value at %f to be 0, got %f", x2, linearMembership.Value(x2))
		}
	}
}

func checkTriangularMembership(t *testing.T, membership fuzzy.Membership, x1, x2, x3 float64) {
	t.Helper()

	triangularMembership, ok := membership.(*fuzzy.TriangularMembership)
	if !ok {
		t.Fatalf("Expected TriangularMembership, got %T", membership)
	}

	min, max := triangularMembership.Domain()
	if !almostEqual(min, x1) || !almostEqual(max, x3) {
		t.Errorf("Expected domain (%f, %f), got (%f, %f)", x1, x3, min, max)
	}

	// Check membership values at key points
	if !almostEqual(triangularMembership.Value(x1), 0) {
		t.Errorf("Expected value at %f to be 0, got %f", x1, triangularMembership.Value(x1))
	}
	if !almostEqual(triangularMembership.Value(x2), 1) {
		t.Errorf("Expected value at %f to be 1, got %f", x2, triangularMembership.Value(x2))
	}
	if !almostEqual(triangularMembership.Value(x3), 0) {
		t.Errorf("Expected value at %f to be 0, got %f", x3, triangularMembership.Value(x3))
	}
}

func checkTrapezoidMembership(t *testing.T, membership fuzzy.Membership, x1, x2, x3, x4 float64) {
	t.Helper()

	trapezoidMembership, ok := membership.(*fuzzy.TrapezoidalMembership)
	if !ok {
		t.Fatalf("Expected TrapezoidalMembership, got %T", membership)
	}

	min, max := trapezoidMembership.Domain()
	if !almostEqual(min, x1) || !almostEqual(max, x4) {
		t.Errorf("Expected domain (%f, %f), got (%f, %f)", x1, x4, min, max)
	}

	// Check membership values at key points
	if !almostEqual(trapezoidMembership.Value(x1), 0) {
		t.Errorf("Expected value at %f to be 0, got %f", x1, trapezoidMembership.Value(x1))
	}
	if !almostEqual(trapezoidMembership.Value(x2), 1) {
		t.Errorf("Expected value at %f to be 1, got %f", x2, trapezoidMembership.Value(x2))
	}
	if !almostEqual(trapezoidMembership.Value(x3), 1) {
		t.Errorf("Expected value at %f to be 1, got %f", x3, trapezoidMembership.Value(x3))
	}
	if !almostEqual(trapezoidMembership.Value(x4), 0) {
		t.Errorf("Expected value at %f to be 0, got %f", x4, trapezoidMembership.Value(x4))
	}
}

// Helper function to check descending linear membership function (implemented as InvertedMembership)
func checkDescendingLinearMembership(t *testing.T, membership fuzzy.Membership, x1, x2 float64) {
	t.Helper()

	// For descending functions (x1 > x2), we implement as Inverted(Linear(x2, x1))
	invertedMembership, ok := membership.(*fuzzy.InvertedMembership)
	if !ok {
		t.Fatalf("Expected InvertedMembership for descending function, got %T", membership)
	}

	// Check domain - it should reflect the original x1, x2 range
	min, max := invertedMembership.Domain()
	if !(almostEqual(min, math.Min(x1, x2)) && almostEqual(max, math.Max(x1, x2))) {
		t.Errorf("Expected domain (%f, %f), got (%f, %f)",
			math.Min(x1, x2), math.Max(x1, x2), min, max)
	}

	// In the actual implementation with InvertedMembership(Linear(x2, x1)):
	// Value at x1 should be 0 (lowest)
	// Value at x2 should be 1 (highest)
	// Value at midpoint should be 0.5
	if !almostEqual(membership.Value(x1), 0.0) {
		t.Errorf("Expected value at %f to be 0.0, got %f", x1, membership.Value(x1))
	}
	if !almostEqual(membership.Value(x2), 1.0) {
		t.Errorf("Expected value at %f to be 1.0, got %f", x2, membership.Value(x2))
	}

	// Test midpoint
	midpoint := (x1 + x2) / 2
	if !almostEqual(membership.Value(midpoint), 0.5) {
		t.Errorf("Expected value at midpoint %f to be 0.5, got %f",
			midpoint, membership.Value(midpoint))
	}
}

// Helper function for float comparison
func almostEqual(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) < epsilon
}
