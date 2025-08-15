# go-fuzzy

[![Go Reference](https://pkg.go.dev/badge/github.com/bornholm/go-fuzzy.svg)](https://pkg.go.dev/github.com/bornholm/go-fuzzy)

A Go library for fuzzy logic and fuzzy inference systems. This library provides tools for defining linguistic variables, fuzzy sets, inference rules, and processing fuzzy logic-based decisions.

## Installation

```bash
go get github.com/bornholm/go-fuzzy
```

## What is Fuzzy Logic?

Fuzzy logic is a form of many-valued logic that deals with reasoning that is approximate rather than fixed and exact. Unlike classical logic which requires everything to be either true (1) or false (0), fuzzy logic allows for "degrees of truth" - values between 0 and 1 that represent partial truth.

This makes fuzzy logic particularly useful in control systems and decision-making processes where inputs don't neatly fit into discrete categories.

## Key Components

### Membership Functions

Fuzzy sets are defined by their membership functions, which describe how much an input belongs to a particular set:

- `Linear` - Linear increasing membership from point a to b
- `Triangular` - Triangle-shaped membership peaking at the middle point
- `Trapezoid` - Trapezoidal membership with a flat top
- `Inverted` - Invert any membership function (1 - μ)

### Variables and Terms

- Variables represent linguistic concepts (e.g., "temperature")
- Terms represent linguistic values for those variables (e.g., "cold", "warm", "hot")

### Rules

Fuzzy rules define the relationships between input and output variables using natural language-like syntax:

```go
If(Is("temperature", "hot")).Then("fan_speed", "high")
```

### Inference Engine

The engine processes inputs through the rules to generate output conclusions.

### Defuzzification

Methods to convert fuzzy output back to crisp values:

- `Centroid` - Center of mass of the output distribution
- `MeanOfMaximum` - Average of the points with maximum membership

## Usage Example

Here's a simple temperature control system example:

```go
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
```

The above example:

1. Creates a fuzzy inference engine
2. Defines temperature (input) and AC mode (output) variables
3. Sets up fuzzy rules linking them
4. Processes an input temperature of 30°C
5. Determines both a numeric output and the best matching term

## Logical Operations

You can use logical operations in your rule conditions:

- `And(expr1, expr2, ...)` - All conditions must be true
- `Or(expr1, expr2, ...)` - At least one condition must be true
- `Not(expr)` - Negates the condition

Example:

```go
fuzzy.If(
    fuzzy.Or(
        fuzzy.Is("temperature", "very_hot"),
        fuzzy.And(
            fuzzy.Is("temperature", "hot"),
            fuzzy.Is("humidity", "high"),
        ),
    ),
).Then("ac_mode", "max_cooling")
```

## Advanced Membership Functions

You can compose more complex membership functions:

```go
// Define a plateau-like membership function
plateauFunction := fuzzy.Max(
    fuzzy.Triangular(0, 10, 20),
    fuzzy.Constant(1.0),  // Constant value of 1.0
)

// Define a multi-modal membership function
multiModal := fuzzy.Max(
    fuzzy.Triangular(0, 10, 20),
    fuzzy.Triangular(30, 40, 50),
)
```

## Domain-Specific Language (DSL) for Rules

This library includes a DSL parser that allows you to define fuzzy rules using a simple text-based format instead of programmatic construction. This makes rule creation more intuitive and readable.

You can see some examples in the [`./cmd/fuzzy/examples`](./cmd/fuzzy/examples) directory.

### Basic Syntax

The basic syntax for a rule is:

```
IF [variable] IS [term] THEN [variable] IS [term];
```

Example:

```
IF temperature IS hot THEN ac_mode IS cooling;
```

### Logical Operators

The DSL supports logical operators for complex conditions:

- `AND` - All conditions must be true
- `OR` - At least one condition must be true
- `NOT` - Negates the condition
- Parentheses `(` and `)` - For grouping expressions

Examples:

```
IF temperature IS hot AND humidity IS high THEN ac_mode IS cooling;
IF temperature IS cold OR humidity IS low THEN ac_mode IS heating;
IF NOT temperature IS hot THEN ac_mode IS heating;
IF (temperature IS cold OR humidity IS high) AND NOT pressure IS low THEN ac_mode IS heating;
```

### Usage Example

Here's how to use the DSL parser:

```go
// Define rules using the DSL
script := `
IF temperature IS cold THEN ac_mode IS heating;
IF temperature IS comfortable THEN ac_mode IS off;
IF temperature IS hot THEN ac_mode IS cooling;
`

// Parse the rules
rules, err := dsl.ParseRules(script)
if err != nil {
  panic(err)
}

// Add the parsed rules to the engine
engine.Rules(rules...)
```

Alternatively, you can use `ParseRulesOrPanic` which will panic on parsing errors:

```go
rules := fuzzy.ParseRulesOrPanic(script)
engine.Rules(rules...)
```

# License

This library is distributed under the MIT license.
