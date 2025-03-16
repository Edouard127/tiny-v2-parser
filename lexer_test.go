package tiny_v2_parser

import (
	"log"
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	lexer := NewLexer(strings.NewReader(input))

	for {
		tok := lexer.NextToken()
		if tok.Type == TokenEOF {
			break
		}
		if tok.Type == TokenError {
			log.Fatalf("Error on line %d: %s\n", tok.Line, tok.Error)
		}

		// Process token
		switch tok.Type {
		case TokenHeader:
			log.Println("Header:", tok.Parts)
		case TokenProperty:
			log.Println("Property:", tok.Parts)
		case TokenClass:
			log.Printf("Class at indent %d: %v\n", tok.Indent, tok.Parts)
		case TokenField:
			log.Println("Field:", tok.Parts)
		case TokenMethod:
			log.Println("Method:", tok.Parts)
		case TokenParameter:
			log.Println("Parameter:", tok.Parts)
		case TokenLocalVariable:
			log.Println("LocalVariable:", tok.Parts)
		case TokenComment:
			log.Println("Comment:", tok.Parts)
		}
	}
}
