package dsl

import (
	"github.com/bornholm/go-fuzzy"
)

// parseVariableDefinition parses a variable definition (DEFINE variable (...);)
func (p *Parser) parseVariableDefinition() (*fuzzy.Variable, error) {
	// Skip DEFINE token
	defineToken := p.tokens[p.current]
	p.current++

	// Get variable name
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected variable name after DEFINE",
			defineToken.Position, nil)
	}
	variableName := p.tokens[p.current].Value
	p.current++

	// Expect open parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenLPAREN {
		return nil, newParseError("expected ( after variable name",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse term definitions
	var terms []*fuzzy.Term
	for p.current < len(p.tokens) && p.tokens[p.current].Type != tokenRPAREN {
		// Each term definition should start with TERM
		if p.tokens[p.current].Type != tokenTERM {
			return nil, newParseError("expected TERM in variable definition",
				p.tokens[p.current].Position, nil)
		}

		term, err := p.parseTermDefinition()
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)

		// After a term definition, expect a comma or closing parenthesis
		if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenCOMMA {
			p.current++ // Skip comma
		}
	}

	// Expect closing parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
		return nil, newParseError("expected ) at end of variable definition",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Expect semicolon
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenSEMI {
		return nil, newParseError("expected ; after variable definition",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Create variable with parsed terms
	return fuzzy.NewVariable(variableName, terms...), nil
}

// parseTermDefinition parses a term definition (TERM name FUNCTION_TYPE (params))
func (p *Parser) parseTermDefinition() (*fuzzy.Term, error) {
	// Skip TERM token
	termToken := p.tokens[p.current]
	p.current++

	// Get term name
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected term name", termToken.Position, nil)
	}
	termName := p.tokens[p.current].Value
	p.current++

	// Next token should be the membership function type
	if p.current >= len(p.tokens) {
		return nil, newParseError("expected membership function type",
			p.tokens[p.current-1].Position, nil)
	}

	// Parse membership function
	membership, err := p.parseMembershipFunction()
	if err != nil {
		return nil, err
	}

	return fuzzy.NewTerm(termName, membership), nil
}
