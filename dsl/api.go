package dsl

import (
	"github.com/bornholm/go-fuzzy"
	"github.com/pkg/errors"
)

type Options struct {
	Memberships map[string]MembershipParser
}

type OptionFunc func(opts *Options)

func NewOptions(funcs ...OptionFunc) *Options {
	opts := &Options{
		Memberships: DefaultMemberships,
	}
	for _, fn := range funcs {
		fn(opts)
	}
	return opts
}

func WithMembershipParser(funcType string, parser MembershipParser) OptionFunc {
	return func(opts *Options) {
		opts.Memberships[funcType] = parser
	}
}

func WithMembershipParsers(parsers map[string]MembershipParser) OptionFunc {
	return func(opts *Options) {
		opts.Memberships = parsers
	}
}

// ParseRules parses DSL text into a slice of Rule objects
func ParseRules(dsl string, funcs ...OptionFunc) ([]*fuzzy.Rule, error) {
	result, err := ParseRulesAndVariables(dsl, funcs...)
	if err != nil {
		return nil, err
	}
	return result.Rules, nil
}

// ParseRulesAndVariables parses DSL text into both rules and variables
func ParseRulesAndVariables(dsl string, funcs ...OptionFunc) (*ParseResult, error) {
	opts := NewOptions(funcs...)
	tokens, err := tokenize(dsl)
	if err != nil {
		return nil, errors.Wrap(err, "tokenization error")
	}

	parser := &Parser{
		tokens:      tokens,
		current:     0,
		memberships: opts.Memberships,
	}

	result, err := parser.parse()
	if err != nil {
		return nil, errors.Wrap(err, "parsing error")
	}

	return result, nil
}

// ParseRulesOrPanic parses DSL text into a slice of Rule objects or panics on error
func ParseRulesOrPanic(dsl string, funcs ...OptionFunc) []*fuzzy.Rule {
	rules, err := ParseRules(dsl, funcs...)
	if err != nil {
		panic(errors.Errorf("failed to parse rules: %v", err))
	}
	return rules
}

// ParseVariables parses DSL text into a slice of Variable objects
func ParseVariables(dsl string, funcs ...OptionFunc) ([]*fuzzy.Variable, error) {
	result, err := ParseRulesAndVariables(dsl, funcs...)
	if err != nil {
		return nil, err
	}
	return result.Variables, nil
}

// ParseVariablesOrPanic parses DSL text into a slice of Variable objects or panics on error
func ParseVariablesOrPanic(dsl string, funcs ...OptionFunc) []*fuzzy.Variable {
	variables, err := ParseVariables(dsl, funcs...)
	if err != nil {
		panic(errors.Errorf("failed to parse variables: %v", err))
	}
	return variables
}
