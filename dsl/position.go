package dsl

import "fmt"

// Position represents a position in the source text
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
}

// String returns a string representation of the position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}