package dsl

import (
	"fmt"

	"github.com/bornholm/go-fuzzy"
)

// parseMembershipFunction parses a membership function definition
func (p *Parser) parseMembershipFunction() (fuzzy.Membership, error) {
	// Get function type
	if p.current >= len(p.tokens) {
		return nil, newParseError("expected membership function type",
			p.tokens[p.current-1].Position, nil)
	}

	funcTypeToken := p.tokens[p.current]
	funcType := funcTypeToken.Type
	p.current++

	// Handle different function types
	switch funcType {
	case tokenLINEAR:
		// Parse LINEAR(x1, x2)
		return p.parseLinearFunction()

	case tokenTRIANGULAR:
		// Parse TRIANGULAR(x1, x2, x3)
		return p.parseTriangularFunction()

	case tokenTRAPEZOID:
		// Parse TRAPEZOID(x1, x2, x3, x4)
		return p.parseTrapezoidFunction()

	case tokenINVERTED:
		// Parse INVERTED(function)
		return p.parseInvertedFunction()

	default:
		return nil, newParseError(
			fmt.Sprintf("unknown membership function type: %s", p.tokens[p.current-1].Value),
			funcTypeToken.Position, nil)
	}
}

// parseLinearFunction parses a LINEAR(x1, x2) membership function
func (p *Parser) parseLinearFunction() (fuzzy.Membership, error) {
	// Expect open parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenLPAREN {
		return nil, newParseError("expected ( after LINEAR",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse first parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected first parameter for LINEAR",
			p.tokens[p.current-1].Position, nil)
	}
	x1Str := p.tokens[p.current].Value
	x1, err := parseFloat(x1Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect comma
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenCOMMA {
		return nil, newParseError("expected , between parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse second parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected second parameter for LINEAR",
			p.tokens[p.current-1].Position, nil)
	}
	x2Str := p.tokens[p.current].Value
	x2, err := parseFloat(x2Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect closing parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
		return nil, newParseError("expected ) after LINEAR parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// The Linear function has issues with descending linear functions (x1 > x2)
	// Create a function that behaves correctly for both ascending and descending cases
	if x1 < x2 {
		// Ascending function: 0 at x1, 1 at x2
		return fuzzy.Linear(x1, x2), nil
	} else if x1 > x2 {
		// Descending function: 1 at x2, 0 at x1
		// For a descending function like LINEAR(10, 0), we create Linear(0, 10) and invert it
		return fuzzy.Inverted(fuzzy.Linear(x2, x1)), nil
	} else {
		// x1 == x2 case - step function
		return fuzzy.Step(x1), nil
	}
}

// parseTriangularFunction parses a TRIANGULAR(x1, x2, x3) membership function
func (p *Parser) parseTriangularFunction() (fuzzy.Membership, error) {
	// Expect open parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenLPAREN {
		return nil, newParseError("expected ( after TRIANGULAR",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse first parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected first parameter for TRIANGULAR",
			p.tokens[p.current-1].Position, nil)
	}
	x1Str := p.tokens[p.current].Value
	x1, err := parseFloat(x1Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect comma
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenCOMMA {
		return nil, newParseError("expected , between parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse second parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected second parameter for TRIANGULAR",
			p.tokens[p.current-1].Position, nil)
	}
	x2Str := p.tokens[p.current].Value
	x2, err := parseFloat(x2Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect comma
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenCOMMA {
		return nil, newParseError("expected , between parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse third parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected third parameter for TRIANGULAR",
			p.tokens[p.current-1].Position, nil)
	}
	x3Str := p.tokens[p.current].Value
	x3, err := parseFloat(x3Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect closing parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
		return nil, newParseError("expected ) after TRIANGULAR parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	return fuzzy.Triangular(x1, x2, x3), nil
}

// parseTrapezoidFunction parses a TRAPEZOID(x1, x2, x3, x4) membership function
func (p *Parser) parseTrapezoidFunction() (fuzzy.Membership, error) {
	// Expect open parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenLPAREN {
		return nil, newParseError("expected ( after TRAPEZOID",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse first parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected first parameter for TRAPEZOID",
			p.tokens[p.current-1].Position, nil)
	}
	x1Str := p.tokens[p.current].Value
	x1, err := parseFloat(x1Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect comma
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenCOMMA {
		return nil, newParseError("expected , between parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse second parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected second parameter for TRAPEZOID",
			p.tokens[p.current-1].Position, nil)
	}
	x2Str := p.tokens[p.current].Value
	x2, err := parseFloat(x2Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect comma
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenCOMMA {
		return nil, newParseError("expected , between parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse third parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected third parameter for TRAPEZOID",
			p.tokens[p.current-1].Position, nil)
	}
	x3Str := p.tokens[p.current].Value
	x3, err := parseFloat(x3Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect comma
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenCOMMA {
		return nil, newParseError("expected , between parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse fourth parameter
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		return nil, newParseError("expected fourth parameter for TRAPEZOID",
			p.tokens[p.current-1].Position, nil)
	}
	x4Str := p.tokens[p.current].Value
	x4, err := parseFloat(x4Str, p.tokens[p.current].Position)
	if err != nil {
		return nil, err
	}
	p.current++

	// Expect closing parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
		return nil, newParseError("expected ) after TRAPEZOID parameters",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	return fuzzy.Trapezoid(x1, x2, x3, x4), nil
}

// parseInvertedFunction parses an INVERTED(function) membership function
func (p *Parser) parseInvertedFunction() (fuzzy.Membership, error) {
	// Expect open parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenLPAREN {
		return nil, newParseError("expected ( after INVERTED",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	// Parse the inner membership function
	innerFunc, err := p.parseMembershipFunction()
	if err != nil {
		return nil, err
	}

	// Expect closing parenthesis
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
		return nil, newParseError("expected ) after INVERTED function",
			p.tokens[p.current-1].Position, nil)
	}
	p.current++

	return fuzzy.Inverted(innerFunc), nil
}

// parseFloat is defined in parser.go and available within the package
