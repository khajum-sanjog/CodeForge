package kzm

import (
	"testing"
)

func TestLexerBasic(t *testing.T) {
	input := `project "My API"
version "1.0"

watch github "myuser/my-api" on branch "main"

before deploy:
  run "npm install"
  run "npm test" must pass

deploy to lambda "my-api-prod":
  region "ap-south-1"
  env NODE_ENV = "production"
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("failed to tokenize: %v", err)
	}

	expectedTypes := []TokenType{
		TokKeyword, TokString, TokNewline, // project "My API"
		TokKeyword, TokString, TokNewline, // version "1.0"
		TokKeywordPhrase, TokString, TokKeywordPhrase, TokString, TokNewline, // watch github "myuser/my-api" on branch "main"
		TokKeywordPhrase, TokColon, TokNewline, // before deploy:
		TokIndent,
		TokKeyword, TokString, TokNewline, // run "npm install"
		TokKeyword, TokString, TokKeywordPhrase, TokNewline, // run "npm test" must pass
		TokDedent,
		TokKeywordPhrase, TokIdentifier, TokString, TokColon, TokNewline, // deploy to lambda "my-api-prod":
		TokIndent,
		TokIdentifier, TokString, TokNewline, // region "ap-south-1"
		TokKeyword, TokIdentifier, TokEquals, TokString, TokNewline, // env NODE_ENV = "production"
		TokDedent,
		TokEOF,
	}

	// Filter out extra Newlines if any, or verify exactly.
	// In our lexer, we add TokNewline at the end of each non-comment/non-empty line.
	// Let's verify that the tokens match our expected sequence.
	if len(tokens) != len(expectedTypes) {
		t.Logf("Tokens obtained:")
		for i, tok := range tokens {
			t.Logf("  [%d] Type: %s, Value: %q, Line: %d", i, tok.Type, tok.Value, tok.Line)
		}
		t.Fatalf("token count mismatch: expected %d, got %d", len(expectedTypes), len(tokens))
	}

	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token [%d] mismatch: expected type %s, got %s (value %q) at line %d", i, exp, tokens[i].Type, tokens[i].Value, tokens[i].Line)
		}
	}
}

func TestLexerIndentationErrors(t *testing.T) {
	input := `before deploy:
  run "step1"
    run "step2"
 run "step3" # invalid dedent
`
	lexer := NewLexer(input)
	_, err := lexer.Tokenize()
	if err == nil {
		t.Fatal("expected indentation error, but got nil")
	}
}
