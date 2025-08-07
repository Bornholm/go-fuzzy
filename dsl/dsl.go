package dsl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bornholm/go-fuzzy"
	"github.com/pkg/errors"
)

// Position represents a position in the source text
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
}

// String returns a string representation of the position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Msg      string
	Pos      Position
	cause    error
	stackErr error // Error with stack trace
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s at %s: %v", e.Msg, e.Pos, e.cause)
	}
	return fmt.Sprintf("%s at %s", e.Msg, e.Pos)
}

// Line returns the line number where the error occurred
func (e *ParseError) Line() int {
	return e.Pos.Line
}

// Column returns the column number where the error occurred
func (e *ParseError) Column() int {
	return e.Pos.Column
}

// Unwrap returns the underlying cause of the error
func (e *ParseError) Unwrap() error {
	return e.stackErr
}

// newParseError creates a new ParseError
func newParseError(msg string, pos Position, cause error) *ParseError {
	var stackErr error
	if cause != nil {
		stackErr = errors.Wrap(cause, msg)
	} else {
		stackErr = errors.New(msg)
	}

	return &ParseError{
		Msg:      msg,
		Pos:      pos,
		cause:    cause,
		stackErr: stackErr,
	}
}

// DSL tokens
const (
	tokenIF     = "IF"
	tokenIS     = "IS"
	tokenTHEN   = "THEN"
	tokenAND    = "AND"
	tokenOR     = "OR"
	tokenNOT    = "NOT"
	tokenSEMI   = ";"
	tokenVAR    = "VARIABLE"
	tokenTERM   = "TERM"
	tokenLPAREN = "("
	tokenRPAREN = ")"
)

// Token represents a lexical token in the DSL
type Token struct {
	Type     string
	Value    string
	Position Position // Position in the source text
}

// tokenize breaks down the input string into tokens with position information
func tokenize(input string) ([]Token, error) {
	var tokens []Token

	// We need to process the input character by character to track positions
	lines := strings.Split(input, "\n")
	var tokenPositions []struct {
		word string
		pos  Position
	}

	// First pass: identify words and their positions
	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers
		column := 1

		// Replace semicolons with spaces around them
		line = strings.ReplaceAll(line, ";", " ; ")
		// Replace parentheses with spaces around them
		line = strings.ReplaceAll(line, "(", " ( ")
		line = strings.ReplaceAll(line, ")", " ) ")

		// Split line into words
		words := strings.Fields(line)

		for _, word := range words {
			if word == "" {
				continue
			}

			// Find the actual column position (skipping leading whitespace)
			wordPos := strings.Index(line[column-1:], word)
			if wordPos >= 0 {
				column += wordPos
			}

			tokenPositions = append(tokenPositions, struct {
				word string
				pos  Position
			}{
				word: word,
				pos:  Position{Line: lineNum, Column: column},
			})

			// Move column position past this word
			column += len(word) + 1
		}
	}

	// Second pass: create tokens with positions
	for _, tp := range tokenPositions {
		word := tp.word
		pos := tp.pos

		var tokenType string
		switch strings.ToUpper(word) {
		case "IF":
			tokenType = tokenIF
		case "IS":
			tokenType = tokenIS
		case "THEN":
			tokenType = tokenTHEN
		case "AND":
			tokenType = tokenAND
		case "OR":
			tokenType = tokenOR
		case "NOT":
			tokenType = tokenNOT
		case "(":
			tokenType = tokenLPAREN
		case ")":
			tokenType = tokenRPAREN
		case ";":
			tokenType = tokenSEMI
		default:
			// If it's not a keyword, it's a variable or term name
			tokenType = tokenVAR
		}

		tokens = append(tokens, Token{
			Type:     tokenType,
			Value:    word,
			Position: pos,
		})
	}

	return tokens, nil
}

// splitWhilePreservingParentheses splits the input string by whitespace
// while preserving parentheses as separate tokens
func splitWhilePreservingParentheses(input string) []string {
	// Replace parentheses with spaces around them to ensure they're separate tokens
	input = strings.ReplaceAll(input, "(", " ( ")
	input = strings.ReplaceAll(input, ")", " ) ")

	// Use regexp to split by whitespace, preserving quoted strings if any
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	matches := r.FindAllString(input, -1)

	// Filter out empty strings
	var result []string
	for _, match := range matches {
		match = strings.TrimSpace(match)
		if match != "" {
			result = append(result, match)
		}
	}

	return result
}

// Parser holds the state during parsing
type Parser struct {
	tokens  []Token
	current int
}

// parse processes the tokens and produces rules
func (p *Parser) parse() ([]*fuzzy.Rule, error) {
	var rules []*fuzzy.Rule
	var errs []string

	for p.current < len(p.tokens) {
		rule, err := p.parseRule()
		if err != nil {
			// Collect errors but keep trying to parse other rules
			errs = append(errs, err.Error())
		}
		if rule != nil {
			rules = append(rules, rule)
		}

		// If we've reached the end of tokens, break
		if p.current >= len(p.tokens) {
			break
		}
	}

	// If we encountered any errors, return them all together
	// Note: We're returning errors even if we successfully parsed some rules
	if len(errs) > 0 {
		return nil, errors.Errorf("parsing errors: %s", strings.Join(errs, "; "))
	}

	return rules, nil
}

// parseRule parses a single rule
func (p *Parser) parseRule() (*fuzzy.Rule, error) {
	// Each rule should start with IF
	if p.current >= len(p.tokens) {
		return nil, nil // End of tokens, just return nil
	}

	if p.tokens[p.current].Type != tokenIF {
		// If we find a token that's not IF, we should report an error
		// But first, let's try to skip to the next semicolon to recover
		tokenPos := p.tokens[p.current].Position
		errorToken := p.tokens[p.current].Value

		for p.current < len(p.tokens) && p.tokens[p.current].Type != tokenSEMI {
			p.current++
		}

		// Skip the semicolon if found
		if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenSEMI {
			p.current++
		}

		return nil, newParseError(
			fmt.Sprintf("expected rule to start with IF, found %s", errorToken),
			tokenPos,
			nil,
		)
	}

	// Skip IF token
	p.current++ // Skip IF

	// Parse premise expression
	premise, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// After premise comes THEN
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenTHEN {
		var pos Position
		if p.current < len(p.tokens) {
			pos = p.tokens[p.current].Position
		} else if p.current > 0 && p.current-1 < len(p.tokens) {
			pos = p.tokens[p.current-1].Position
		} else {
			pos = Position{Line: 1, Column: 1} // Fallback
		}

		return nil, newParseError("expected THEN after premise", pos, nil)
	}
	p.current++ // Skip THEN

	// Parse conclusion (which is always an IS expression)
	variable, term, err := p.parseIsExpression()
	if err != nil {
		return nil, err
	}

	// End of rule should be semicolon
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenSEMI {
		// Missing semicolon at the end of the rule
		var pos Position
		if p.current < len(p.tokens) {
			pos = p.tokens[p.current].Position
		} else if p.current > 0 && p.current-1 < len(p.tokens) {
			pos = p.tokens[p.current-1].Position
		} else {
			pos = Position{Line: 1, Column: 1} // Fallback
		}

		// Save the current state to create the rule even without a semicolon
		ruleWithoutSemicolon := fuzzy.If(premise).Then(variable, term)

		// Try to find the next IF token to continue parsing
		for p.current < len(p.tokens) && p.tokens[p.current].Type != tokenIF {
			p.current++
		}

		return ruleWithoutSemicolon, newParseError("missing semicolon at end of rule", pos, nil)
	}
	p.current++ // Skip semicolon

	// Create and return the rule
	rule := fuzzy.If(premise).Then(variable, term)
	return rule, nil
}

// parseExpression parses an expression (which can be an IS expression or a logical combination)
func (p *Parser) parseExpression() (fuzzy.Expr, error) {
	// Handle NOT
	if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenNOT {
		p.current++ // Skip NOT

		// Handle parentheses after NOT
		if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenLPAREN {
			p.current++ // Skip (
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
				var pos Position
				if p.current < len(p.tokens) {
					pos = p.tokens[p.current].Position
				} else if p.current > 0 && p.current-1 < len(p.tokens) {
					pos = p.tokens[p.current-1].Position
				} else {
					pos = Position{Line: 1, Column: 1} // Fallback
				}
				return nil, newParseError("missing closing parenthesis", pos, nil)
			}
			p.current++ // Skip )

			// Apply NOT to the expression and check for logical combinations
			notExpr := fuzzy.Not(expr)
			return p.parseLogicalCombination(notExpr)
		}

		// Parse the next expression (which could be a simple expression or another complex one)
		// For "NOT pressure IS low", we need to properly handle it as a simple expression
		var expr fuzzy.Expr
		var err error

		// Check if next token is a variable (indicating a simple expression like "pressure IS low")
		if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenVAR {
			expr, err = p.parseSimpleExpression()
		} else {
			expr, err = p.parseExpression()
		}

		if err != nil {
			return nil, err
		}

		// Apply NOT and check for logical combinations (AND/OR) that might follow
		notExpr := fuzzy.Not(expr)
		return p.parseLogicalCombination(notExpr)
	}

	// Handle parenthesized expression
	if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenLPAREN {
		p.current++ // Skip (
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenRPAREN {
			var pos Position
			if p.current < len(p.tokens) {
				pos = p.tokens[p.current].Position
			} else if p.current > 0 && p.current-1 < len(p.tokens) {
				pos = p.tokens[p.current-1].Position
			} else {
				pos = Position{Line: 1, Column: 1} // Fallback
			}
			return nil, newParseError("missing closing parenthesis", pos, nil)
		}
		p.current++ // Skip )

		// Check for AND or OR after this expression
		return p.parseLogicalCombination(expr)
	}

	// Parse a simple expression (like "temperature IS hot")
	expr, err := p.parseSimpleExpression()
	if err != nil {
		return nil, err
	}

	// Check for AND or OR after this expression
	return p.parseLogicalCombination(expr)
}

// parseSimpleExpression parses a simple expression (variable IS term)
func (p *Parser) parseSimpleExpression() (fuzzy.Expr, error) {
	variable, term, err := p.parseIsExpression()
	if err != nil {
		return nil, err
	}

	return fuzzy.Is(variable, term), nil
}

// parseIsExpression parses a variable IS term expression and returns the variable and term
func (p *Parser) parseIsExpression() (string, string, error) {
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		var pos Position
		if p.current < len(p.tokens) {
			pos = p.tokens[p.current].Position
		} else if p.current > 0 && p.current-1 < len(p.tokens) {
			pos = p.tokens[p.current-1].Position
		} else {
			pos = Position{Line: 1, Column: 1} // Fallback
		}
		return "", "", newParseError("expected variable name", pos, nil)
	}
	variable := p.tokens[p.current].Value
	varToken := p.tokens[p.current]
	p.current++ // Skip variable

	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenIS {
		// Use the position right after the variable name
		pos := Position{
			Line:   varToken.Position.Line,
			Column: varToken.Position.Column + len(varToken.Value) + 1,
		}
		return "", "", newParseError("expected IS after variable", pos, nil)
	}
	p.current++ // Skip IS

	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenVAR {
		var pos Position
		if p.current < len(p.tokens) {
			pos = p.tokens[p.current].Position
		} else if p.current > 0 && p.current-1 < len(p.tokens) {
			pos = p.tokens[p.current-1].Position
		} else {
			pos = Position{Line: 1, Column: 1} // Fallback
		}
		return "", "", newParseError("expected term name after IS", pos, nil)
	}
	term := p.tokens[p.current].Value
	p.current++ // Skip term

	return variable, term, nil
}

// parseLogicalCombination handles AND/OR combinations
func (p *Parser) parseLogicalCombination(left fuzzy.Expr) (fuzzy.Expr, error) {
	// Check if there's an AND or OR following
	if p.current < len(p.tokens) {
		if p.tokens[p.current].Type == tokenAND {
			p.current++ // Skip AND

			// Parse the right side expression
			right, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			// If left is already an AndExpr, add to it
			if leftAnd, ok := left.(*fuzzy.AndExpr); ok {
				// Create new slice to avoid modifying the original
				newExprs := make([]fuzzy.Expr, len(leftAnd.Exprs())+1)
				copy(newExprs, leftAnd.Exprs())

				// Check if right is also an AndExpr
				if rightAnd, ok := right.(*fuzzy.AndExpr); ok {
					// Flatten nested AND expressions
					newExprs = append(newExprs[:len(leftAnd.Exprs())], rightAnd.Exprs()...)
				} else {
					newExprs[len(leftAnd.Exprs())] = right
				}

				return fuzzy.And(newExprs...), nil
			}

			// Check if right is an AndExpr
			if rightAnd, ok := right.(*fuzzy.AndExpr); ok {
				// Create new slice with left as first element
				newExprs := make([]fuzzy.Expr, len(rightAnd.Exprs())+1)
				newExprs[0] = left
				copy(newExprs[1:], rightAnd.Exprs())

				return fuzzy.And(newExprs...), nil
			}

			return fuzzy.And(left, right), nil
		} else if p.tokens[p.current].Type == tokenOR {
			p.current++ // Skip OR

			// Parse the right side expression
			right, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			// If left is already an OrExpr, add to it
			if leftOr, ok := left.(*fuzzy.OrExpr); ok {
				// Create new slice to avoid modifying the original
				newExprs := make([]fuzzy.Expr, len(leftOr.Exprs())+1)
				copy(newExprs, leftOr.Exprs())

				// Check if right is also an OrExpr
				if rightOr, ok := right.(*fuzzy.OrExpr); ok {
					// Flatten nested OR expressions
					newExprs = append(newExprs[:len(leftOr.Exprs())], rightOr.Exprs()...)
				} else {
					newExprs[len(leftOr.Exprs())] = right
				}

				return fuzzy.Or(newExprs...), nil
			}

			// Check if right is an OrExpr
			if rightOr, ok := right.(*fuzzy.OrExpr); ok {
				// Create new slice with left as first element
				newExprs := make([]fuzzy.Expr, len(rightOr.Exprs())+1)
				newExprs[0] = left
				copy(newExprs[1:], rightOr.Exprs())

				return fuzzy.Or(newExprs...), nil
			}

			return fuzzy.Or(left, right), nil
		}
	}

	// No AND/OR, just return the expression
	return left, nil
}

// ParseRules parses DSL text into a slice of Rule objects
func ParseRules(dsl string) ([]*fuzzy.Rule, error) {
	tokens, err := tokenize(dsl)
	if err != nil {
		return nil, errors.Wrap(err, "tokenization error")
	}

	parser := &Parser{
		tokens:  tokens,
		current: 0,
	}

	rules, err := parser.parse()
	if err != nil {
		return nil, errors.Wrap(err, "parsing error")
	}

	return rules, nil
}

// ParseRulesOrPanic parses DSL text into a slice of Rule objects or panics on error
func ParseRulesOrPanic(dsl string) []*fuzzy.Rule {
	rules, err := ParseRules(dsl)
	if err != nil {
		panic(fmt.Sprintf("failed to parse rules: %v", err))
	}
	return rules
}
