package kzm

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType defines the category of a lexical token.
type TokenType string

const (
	TokKeyword       TokenType = "KEYWORD"
	TokKeywordPhrase TokenType = "KEYWORD_PHRASE"
	TokString        TokenType = "STRING"
	TokNumber        TokenType = "NUMBER"
	TokBoolean       TokenType = "BOOLEAN"
	TokIdentifier    TokenType = "IDENTIFIER"
	TokEquals        TokenType = "EQUALS"
	TokColon         TokenType = "COLON"
	TokComma         TokenType = "COMMA"
	TokIndent        TokenType = "INDENT"
	TokDedent        TokenType = "DEDENT"
	TokNewline       TokenType = "NEWLINE"
	TokEOF           TokenType = "EOF"
	TokError         TokenType = "ERROR"
)

// Token represents a single token in the KZM source code.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Lexer processes KZM source text into a slice of tokens.
type Lexer struct {
	input       string
	pos         int
	line        int
	col         int
	indentStack []int
	tokens      []Token
}

// NewLexer creates a new Lexer instance.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:       input,
		line:        1,
		col:         1,
		indentStack: []int{0},
	}
}

// Tokenize scans the entire input string and returns all tokens or an error.
func (l *Lexer) Tokenize() ([]Token, error) {
	lines := strings.Split(l.input, "\n")
	for i, lineStr := range lines {
		l.line = i + 1
		l.col = 1

		// 1. Check if line is empty or just a comment
		trimmed := strings.TrimSpace(lineStr)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 2. Count leading spaces
		leadingSpaces := 0
		for _, r := range lineStr {
			if r == ' ' {
				leadingSpaces++
			} else if r == '\t' {
				// Tabs are treated as 2 spaces
				leadingSpaces += 2
			} else {
				break
			}
		}

		// 3. Indent/Dedent logic
		currentIndent := l.indentStack[len(l.indentStack)-1]
		if leadingSpaces > currentIndent {
			l.indentStack = append(l.indentStack, leadingSpaces)
			l.tokens = append(l.tokens, Token{Type: TokIndent, Value: "", Line: l.line, Col: 1})
		} else if leadingSpaces < currentIndent {
			for len(l.indentStack) > 0 && l.indentStack[len(l.indentStack)-1] > leadingSpaces {
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
				l.tokens = append(l.tokens, Token{Type: TokDedent, Value: "", Line: l.line, Col: 1})
			}
			if len(l.indentStack) == 0 || l.indentStack[len(l.indentStack)-1] != leadingSpaces {
				return nil, fmt.Errorf("inconsistent indentation on line %d: got %d spaces", l.line, leadingSpaces)
			}
		}

		// 4. Tokenize the line content
		l.pos = 0
		// We use lineStr from the first non-space character
		lineRunes := []rune(lineStr)
		l.pos = leadingSpaces

		for l.pos < len(lineRunes) {
			r := lineRunes[l.pos]
			l.col = l.pos + 1

			// Skip spaces and tabs
			if r == ' ' || r == '\t' {
				l.pos++
				continue
			}

			// Comment
			if r == '#' {
				break // Skip rest of the line
			}

			// Colon
			if r == ':' {
				l.tokens = append(l.tokens, Token{Type: TokColon, Value: ":", Line: l.line, Col: l.col})
				l.pos++
				continue
			}

			// Equals
			if r == '=' {
				l.tokens = append(l.tokens, Token{Type: TokEquals, Value: "=", Line: l.line, Col: l.col})
				l.pos++
				continue
			}

			// Comma
			if r == ',' {
				l.tokens = append(l.tokens, Token{Type: TokComma, Value: ",", Line: l.line, Col: l.col})
				l.pos++
				continue
			}

			// String
			if r == '"' {
				strVal, consumed, err := l.scanString(lineRunes[l.pos:])
				if err != nil {
					return nil, err
				}
				l.tokens = append(l.tokens, Token{Type: TokString, Value: strVal, Line: l.line, Col: l.col})
				l.pos += consumed
				continue
			}

			// Number
			if unicode.IsDigit(r) {
				numVal, consumed := l.scanNumber(lineRunes[l.pos:])
				l.tokens = append(l.tokens, Token{Type: TokNumber, Value: numVal, Line: l.line, Col: l.col})
				l.pos += consumed
				continue
			}

			// Words (keywords, keyword phrases, booleans, identifiers)
			matchedPhrase, consumedPhrase := l.matchKeywordPhrase(lineRunes[l.pos:])
			if matchedPhrase != "" {
				l.tokens = append(l.tokens, Token{Type: TokKeywordPhrase, Value: matchedPhrase, Line: l.line, Col: l.col})
				l.pos += consumedPhrase
				continue
			}

			word, consumed := l.scanWord(lineRunes[l.pos:])
			if word != "" {
				l.tokens = append(l.tokens, l.categorizeWord(word, l.line, l.col))
				l.pos += consumed
				continue
			}

			// Fallback: character error
			return nil, fmt.Errorf("lexical error: unexpected character %q at line %d, col %d", r, l.line, l.col)
		}

		// Emit Newline at the end of the line
		l.tokens = append(l.tokens, Token{Type: TokNewline, Value: "\n", Line: l.line, Col: len(lineRunes) + 1})
	}

	// 5. Deduce remaining indentations
	for len(l.indentStack) > 1 {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		l.tokens = append(l.tokens, Token{Type: TokDedent, Value: "", Line: l.line, Col: 1})
	}

	// 6. Emit EOF
	l.tokens = append(l.tokens, Token{Type: TokEOF, Value: "", Line: l.line, Col: 1})

	return l.tokens, nil
}

func (l *Lexer) scanString(runes []rune) (string, int, error) {
	// runes[0] is '"'
	var sb strings.Builder
	pos := 1
	escaped := false
	for pos < len(runes) {
		r := runes[pos]
		if escaped {
			sb.WriteRune(r)
			escaped = false
			pos++
			continue
		}
		if r == '\\' {
			escaped = true
			pos++
			continue
		}
		if r == '"' {
			return sb.String(), pos + 1, nil
		}
		sb.WriteRune(r)
		pos++
	}
	return "", 0, fmt.Errorf("unterminated string literal starting at line %d", l.line)
}

func (l *Lexer) scanNumber(runes []rune) (string, int) {
	var sb strings.Builder
	pos := 0
	for pos < len(runes) && unicode.IsDigit(runes[pos]) {
		sb.WriteRune(runes[pos])
		pos++
	}
	return sb.String(), pos
}

func (l *Lexer) scanWord(runes []rune) (string, int) {
	var sb strings.Builder
	pos := 0
	for pos < len(runes) {
		r := runes[pos]
		// Identifiers can contain letters, digits, underscores, hyphens, dots, forward slashes, backslashes, tilde, and @ signs.
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' || r == '/' || r == '\\' || r == '~' || r == '@' {
			sb.WriteRune(r)
			pos++
		} else {
			break
		}
	}
	return sb.String(), pos
}

func (l *Lexer) matchKeywordPhrase(runes []rune) (string, int) {
	phrases := []string{
		"must pass or rollback",
		"skip if no changes",
		"use secrets from",
		"before deploy",
		"after deploy",
		"watch github",
		"watch folder",
		"watch gitlab",
		"on trigger",
		"deploy to",
		"on branch",
		"keep last",
		"must pass",
		"if env is",
	}

	remainingStr := string(runes)
	for _, p := range phrases {
		if strings.HasPrefix(remainingStr, p) {
			length := len(p)
			// Ensure it ends on a word boundary
			if len(remainingStr) == length || isBoundary(rune(remainingStr[length])) {
				// Convert byte length to rune length
				runeLen := len([]rune(p))
				return p, runeLen
			}
		}
	}
	return "", 0
}

func isBoundary(r rune) bool {
	return unicode.IsSpace(r) || r == ':' || r == '=' || r == ',' || r == '#' || r == '"'
}

func (l *Lexer) categorizeWord(word string, line, col int) Token {
	keywords := map[string]bool{
		"project":      true,
		"version":      true,
		"description":  true,
		"watch":        true,
		"every":        true,
		"notify":       true,
		"run":          true,
		"copy":         true,
		"set":          true,
		"plugin":       true,
		"environments": true,
		"if":           true,
		"env":          true,
	}

	booleans := map[string]bool{
		"yes":   true,
		"no":    true,
		"true":  true,
		"false": true,
	}

	lowerWord := strings.ToLower(word)
	if keywords[lowerWord] {
		return Token{Type: TokKeyword, Value: lowerWord, Line: line, Col: col}
	}
	if booleans[lowerWord] {
		return Token{Type: TokBoolean, Value: lowerWord, Line: line, Col: col}
	}
	// Defaults to Identifier
	return Token{Type: TokIdentifier, Value: word, Line: line, Col: col}
}
