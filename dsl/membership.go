package dsl

import (
	"fmt"

	"github.com/bornholm/go-fuzzy"
	"github.com/pkg/errors"
)

type MembershipParser interface {
	ParseMembership(tokens []Token, current int, parse ParseMembershipFunc) (membership fuzzy.Membership, newCurrent int, err error)
}

type ParseMembershipFunc func(tokens []Token, current int, parse ParseMembershipFunc) (fuzzy.Membership, int, error)

func (fn ParseMembershipFunc) ParseMembership(tokens []Token, current int, parse ParseMembershipFunc) (membership fuzzy.Membership, newCurrent int, err error) {
	return fn(tokens, current, parse)
}

// parseMembershipFunction parses a membership function definition
func (p *Parser) parseMembershipFunction() (fuzzy.Membership, error) {
	// Get function type
	if p.current >= len(p.tokens) {
		return nil, newParseError("expected membership function type",
			p.tokens[p.current-1].Position, nil)
	}

	var parseMembership ParseMembershipFunc = func(tokens []Token, current int, parse ParseMembershipFunc) (membership fuzzy.Membership, newCurrent int, err error) {
		funcTypeToken := tokens[current]
		funcType := funcTypeToken.Type

		current++

		membershipParser, exists := p.memberships[funcType]
		if !exists {
			return nil, current, newParseError(
				fmt.Sprintf("unknown membership function type: %s", tokens[current-1].Value),
				funcTypeToken.Position, nil)
		}

		return membershipParser.ParseMembership(tokens, current, parse)
	}

	membership, current, err := parseMembership(p.tokens, p.current, parseMembership)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	p.current = current

	return membership, nil
}

const (
	tokenLINEAR     string = "LINEAR"
	tokenTRIANGULAR string = "TRIANGULAR"
	tokenTRAPEZOID  string = "TRAPEZOID"
	tokenINVERTED   string = "INVERTED"
)

var DefaultMemberships = map[string]MembershipParser{
	tokenLINEAR:     ParseMembershipFunc(ParseLinear),
	tokenTRIANGULAR: ParseMembershipFunc(ParseTriangular),
	tokenTRAPEZOID:  ParseMembershipFunc(ParseTrapezoid),
	tokenINVERTED:   ParseMembershipFunc(ParseInverted),
}

// ParseLinear parses a LINEAR(x1, x2) membership function
func ParseLinear(tokens []Token, current int, parse ParseMembershipFunc) (fuzzy.Membership, int, error) {
	// Expect open parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenLPAREN {
		return nil, current, newParseError("expected ( after LINEAR",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse first parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected first parameter for LINEAR",
			tokens[current-1].Position, nil)
	}
	x1Str := tokens[current].Value
	x1, err := parseFloat(x1Str, tokens[current].Position)
	if err != nil {
		return nil, current, err
	}
	current++

	// Expect comma
	if current >= len(tokens) || tokens[current].Type != tokenCOMMA {
		return nil, current, newParseError("expected , between parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse second parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected second parameter for LINEAR",
			tokens[current-1].Position, nil)
	}
	x2Str := tokens[current].Value
	x2, err := parseFloat(x2Str, tokens[current].Position)
	if err != nil {
		return nil, current, err
	}
	current++

	// Expect closing parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenRPAREN {
		return nil, current, newParseError("expected ) after LINEAR parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// The Linear function has issues with descending linear functions (x1 > x2)
	// Create a function that behaves correctly for both ascending and descending cases
	if x1 < x2 {
		// Ascending function: 0 at x1, 1 at x2
		return fuzzy.Linear(x1, x2), current, nil
	} else if x1 > x2 {
		// Descending function: 1 at x2, 0 at x1
		// For a descending function like LINEAR(10, 0), we create Linear(0, 10) and invert it
		return fuzzy.Inverted(fuzzy.Linear(x2, x1)), current, nil
	} else {
		// x1 == x2 case - step function
		return fuzzy.Step(x1), current, nil
	}
}

// ParseTriangular parses a TRIANGULAR(x1, x2, x3) membership function
func ParseTriangular(tokens []Token, current int, parse ParseMembershipFunc) (fuzzy.Membership, int, error) {
	// Expect open parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenLPAREN {
		return nil, current, newParseError("expected ( after TRIANGULAR",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse first parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected first parameter for TRIANGULAR",
			tokens[current-1].Position, nil)
	}
	x1Str := tokens[current].Value
	x1, err := parseFloat(x1Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect comma
	if current >= len(tokens) || tokens[current].Type != tokenCOMMA {
		return nil, current, newParseError("expected , between parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse second parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected second parameter for TRIANGULAR",
			tokens[current-1].Position, nil)
	}
	x2Str := tokens[current].Value
	x2, err := parseFloat(x2Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect comma
	if current >= len(tokens) || tokens[current].Type != tokenCOMMA {
		return nil, current, newParseError("expected , between parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse third parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected third parameter for TRIANGULAR",
			tokens[current-1].Position, nil)
	}
	x3Str := tokens[current].Value
	x3, err := parseFloat(x3Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect closing parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenRPAREN {
		return nil, current, newParseError("expected ) after TRIANGULAR parameters",
			tokens[current-1].Position, nil)
	}
	current++

	return fuzzy.Triangular(x1, x2, x3), current, nil
}

// ParseTrapezoid parses a TRAPEZOID(x1, x2, x3, x4) membership function
func ParseTrapezoid(tokens []Token, current int, parse ParseMembershipFunc) (fuzzy.Membership, int, error) {
	// Expect open parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenLPAREN {
		return nil, current, newParseError("expected ( after TRAPEZOID",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse first parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected first parameter for TRAPEZOID",
			tokens[current-1].Position, nil)
	}
	x1Str := tokens[current].Value
	x1, err := parseFloat(x1Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect comma
	if current >= len(tokens) || tokens[current].Type != tokenCOMMA {
		return nil, current, newParseError("expected , between parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse second parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected second parameter for TRAPEZOID",
			tokens[current-1].Position, nil)
	}
	x2Str := tokens[current].Value
	x2, err := parseFloat(x2Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect comma
	if current >= len(tokens) || tokens[current].Type != tokenCOMMA {
		return nil, current, newParseError("expected , between parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse third parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected third parameter for TRAPEZOID",
			tokens[current-1].Position, nil)
	}
	x3Str := tokens[current].Value
	x3, err := parseFloat(x3Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect comma
	if current >= len(tokens) || tokens[current].Type != tokenCOMMA {
		return nil, current, newParseError("expected , between parameters",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse fourth parameter
	if current >= len(tokens) || tokens[current].Type != tokenVAR {
		return nil, current, newParseError("expected fourth parameter for TRAPEZOID",
			tokens[current-1].Position, nil)
	}
	x4Str := tokens[current].Value
	x4, err := parseFloat(x4Str, tokens[current].Position)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}
	current++

	// Expect closing parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenRPAREN {
		return nil, current, newParseError("expected ) after TRAPEZOID parameters",
			tokens[current-1].Position, nil)
	}
	current++

	return fuzzy.Trapezoid(x1, x2, x3, x4), current, nil
}

// ParseInverted parses an INVERTED(function) membership function
func ParseInverted(tokens []Token, current int, parse ParseMembershipFunc) (fuzzy.Membership, int, error) {
	// Expect open parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenLPAREN {
		return nil, current, newParseError("expected ( after INVERTED",
			tokens[current-1].Position, nil)
	}
	current++

	// Parse the inner membership function
	innerFunc, current, err := parse(tokens, current, parse)
	if err != nil {
		return nil, current, errors.WithStack(err)
	}

	// Expect closing parenthesis
	if current >= len(tokens) || tokens[current].Type != tokenRPAREN {
		return nil, current, newParseError("expected ) after INVERTED function",
			tokens[current-1].Position, nil)
	}
	current++

	return fuzzy.Inverted(innerFunc), current, nil
}
