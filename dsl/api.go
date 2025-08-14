package dsl

import (
	"fmt"

	"github.com/bornholm/go-fuzzy"
	"github.com/pkg/errors"
)

// ParseRules parses DSL text into a slice of Rule objects
func ParseRules(dsl string) ([]*fuzzy.Rule, error) {
	result, err := ParseRulesAndVariables(dsl)
	if err != nil {
		return nil, err
	}
	return result.Rules, nil
}

// ParseRulesAndVariables parses DSL text into both rules and variables
func ParseRulesAndVariables(dsl string) (*ParseResult, error) {
	tokens, err := tokenize(dsl)
	if err != nil {
		return nil, errors.Wrap(err, "tokenization error")
	}

	parser := &Parser{
		tokens:  tokens,
		current: 0,
	}

	result, err := parser.parse()
	if err != nil {
		return nil, errors.Wrap(err, "parsing error")
	}

	return result, nil
}

// ParseRulesOrPanic parses DSL text into a slice of Rule objects or panics on error
func ParseRulesOrPanic(dsl string) []*fuzzy.Rule {
	rules, err := ParseRules(dsl)
	if err != nil {
		panic(fmt.Sprintf("failed to parse rules: %v", err))
	}
	return rules
}

// ParseVariables parses DSL text into a slice of Variable objects
func ParseVariables(dsl string) ([]*fuzzy.Variable, error) {
	result, err := ParseRulesAndVariables(dsl)
	if err != nil {
		return nil, err
	}
	return result.Variables, nil
}

// ParseVariablesOrPanic parses DSL text into a slice of Variable objects or panics on error
func ParseVariablesOrPanic(dsl string) []*fuzzy.Variable {
	variables, err := ParseVariables(dsl)
	if err != nil {
		panic(fmt.Sprintf("failed to parse variables: %v", err))
	}
	return variables
}