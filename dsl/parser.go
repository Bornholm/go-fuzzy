package dsl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bornholm/go-fuzzy"
	"github.com/pkg/errors"
)

// ParseResult contains both rules and variables parsed from the DSL
type ParseResult struct {
	Rules     []*fuzzy.Rule
	Variables []*fuzzy.Variable
}

// Parser holds the state during parsing
type Parser struct {
	tokens      []Token
	current     int
	memberships map[string]MembershipParser
}

// parse processes the tokens and produces rules and variables
func (p *Parser) parse() (*ParseResult, error) {
	var rules []*fuzzy.Rule
	var variables []*fuzzy.Variable
	var errs []string

	for p.current < len(p.tokens) {
		if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenDEFINE {
			// Parse variable definition
			variable, err := p.parseVariableDefinition()
			if err != nil {
				errs = append(errs, err.Error())
			}
			if variable != nil {
				variables = append(variables, variable)
			}
		} else {
			// Parse rule
			rule, err := p.parseRule()
			if err != nil {
				errs = append(errs, err.Error())
			}
			if rule != nil {
				rules = append(rules, rule)
			}
		}

		// If we've reached the end of tokens, break
		if p.current >= len(p.tokens) {
			break
		}
	}

	// If we encountered any errors, return them all together
	if len(errs) > 0 {
		return nil, errors.Errorf("parsing errors: %s", strings.Join(errs, "; "))
	}

	return &ParseResult{
		Rules:     rules,
		Variables: variables,
	}, nil
}

// parseFloat parses a string to a float64
func parseFloat(s string, pos Position) (float64, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, newParseError(fmt.Sprintf("invalid number: %s", s), pos, err)
	}
	return val, nil
}
