package dsl

import "strings"

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