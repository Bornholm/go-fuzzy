// Package dsl provides a domain-specific language for creating fuzzy logic rules
package dsl

// This file serves as the main entry point for the DSL package.
// The actual implementation is spread across multiple files for better modularity:
//
// - position.go: Position struct for tracking source positions
// - errors.go: Error handling and reporting
// - token.go: Lexical token definitions and tokenization
// - comments.go: Comment handling in source code
// - parser.go: Main parser logic
// - expressions.go: Parsing of logical expressions (IF/THEN/AND/OR/NOT)
// - variables.go: Variable definition handling
// - membership.go: Membership function parsing 
// - api.go: Public API methods

// The package exposes several primary methods:
// - ParseRules: Parse DSL text into Rule objects
// - ParseRulesAndVariables: Parse DSL text into both rules and variables
// - ParseRulesOrPanic: Parse DSL text into Rule objects, panicking on error
// - ParseVariables: Parse DSL text into Variable objects
// - ParseVariablesOrPanic: Parse DSL text into Variable objects, panicking on error