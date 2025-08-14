package dsl

import (
	"strings"
)

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

	// Tokens for variable definitions
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