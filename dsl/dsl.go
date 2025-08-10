package dsl

import (
	"fmt"
	"strconv"
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

	// New tokens for variable definitions
	tokenDEFINE     = "DEFINE"
	tokenCOMMA      = ","
	tokenLINEAR     = "LINEAR"
	tokenTRIANGULAR = "TRIANGULAR"
	tokenTRAPEZOID  = "TRAPEZOID"
	tokenINVERTED   = "INVERTED"
)

// Token represents a lexical token in the DSL
type Token struct {
	Type     string
	Value    string
	Position Position // Position in the source text
}

// removeComments removes all comments from the input text while precisely preserving code structure
func removeComments(input string) string {
	var result strings.Builder
	inMultilineComment := false
	i := 0

	for i < len(input) {
		// If we're in a multi-line comment, look for the end
		if inMultilineComment {
			if i+1 < len(input) && input[i] == '*' && input[i+1] == '/' {
				inMultilineComment = false
				i += 2 // Skip the */

				// Always add a space to ensure tokens don't merge
				result.WriteByte(' ')
				continue
			}

			// Preserve all newlines in multi-line comments to maintain line numbers
			if input[i] == '\n' {
				result.WriteByte('\n')
			} else {
				// Replace other comment characters with spaces to maintain token separation
				result.WriteByte(' ')
			}

			i++ // Move to next character
			continue
		}

		// Check for start of single-line comment
		if i+1 < len(input) && input[i] == '/' && input[i+1] == '/' {
			// Skip to the end of this line
			endOfLine := strings.IndexByte(input[i:], '\n')
			if endOfLine == -1 {
				// No more newlines (end of file), we're done
				// Add a newline to ensure proper parsing of the last line
				result.WriteByte('\n')
				break
			}

			// Replace all characters in the comment with spaces
			// This preserves column alignment and ensures proper token separation
			for j := 0; j < endOfLine; j++ {
				result.WriteByte(' ')
			}

			// Move to the newline
			i += endOfLine

			// Don't skip the newline itself, preserve it
			result.WriteByte('\n')
			i++
			continue
		}

		// Check for start of multi-line comment
		if i+1 < len(input) && input[i] == '/' && input[i+1] == '*' {
			inMultilineComment = true
			i += 2 // Skip the /*

			// Add a space to ensure tokens don't merge
			result.WriteByte(' ')
			continue
		}

		// Not in a comment, add this character to the result
		result.WriteByte(input[i])
		i++
	}

	return result.String()
}

// tokenize breaks down the input string into tokens with position information
func tokenize(input string) ([]Token, error) {
	// First, remove all comments while preserving structure
	cleanedInput := removeComments(input)

	var tokens []Token
	var tokenPositions []struct {
		word string
		pos  Position
	}

	// Split input into lines
	lines := strings.Split(cleanedInput, "\n")

	// Process each line
	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers
		column := 1

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Prepare line for tokenization
		// Replace special characters with spaces around them
		line = strings.ReplaceAll(line, ";", " ; ")
		line = strings.ReplaceAll(line, "(", " ( ")
		line = strings.ReplaceAll(line, ")", " ) ")
		line = strings.ReplaceAll(line, ",", " , ")

		// Split line into words
		words := strings.Fields(line)

		for _, word := range words {
			if word == "" {
				continue
			}

			// Find the actual column position in the line
			// Using a safer approach to avoid out-of-bounds errors
			wordPos := strings.Index(line, word)
			if wordPos >= 0 {
				column = wordPos + 1 // 1-based column indexing
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
		case "DEFINE":
			tokenType = tokenDEFINE
		case "TERM":
			tokenType = tokenTERM
		case "LINEAR":
			tokenType = tokenLINEAR
		case "TRIANGULAR":
			tokenType = tokenTRIANGULAR
		case "TRAPEZOID":
			tokenType = tokenTRAPEZOID
		case "INVERTED":
			tokenType = tokenINVERTED
		case "(":
			tokenType = tokenLPAREN
		case ")":
			tokenType = tokenRPAREN
		case ";":
			tokenType = tokenSEMI
		case ",":
			tokenType = tokenCOMMA
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

// ParseResult contains both rules and variables parsed from the DSL
type ParseResult struct {
	Rules     []*fuzzy.Rule
	Variables []*fuzzy.Variable
}

// Parser holds the state during parsing
type Parser struct {
	tokens  []Token
	current int
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

// parseFloat parses a string to a float64
func parseFloat(s string, pos Position) (float64, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, newParseError(fmt.Sprintf("invalid number: %s", s), pos, err)
	}
	return val, nil
}

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
