package tiny_v2_parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const Separator = "\t"

// Lexer maintains the parsing state
type Lexer struct {
	reader     *bufio.Scanner
	lineNumber int
	state      lexerState
	namespaces int // From header
}

type lexerState int

const (
	stateStart lexerState = iota
	stateHeaderParsed
	stateInBody
)

// NewLexer creates a new lexer for Tiny V2 format
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewScanner(r),
		state:  stateStart,
	}
}

// NextToken returns the next parsed token
func (l *Lexer) NextToken() Token {
	if !l.reader.Scan() {
		if err := l.reader.Err(); err != nil {
			return l.errorToken(fmt.Sprintf(ErrIo, err))
		}
		return Token{Type: TokenEOF}
	}

	l.lineNumber++
	line := l.reader.Text()

	switch l.state {
	case stateStart:
		return l.parseHeader(line)
	case stateHeaderParsed:
		return l.parsePossibleProperty(line)
	default:
		return l.parseBodyLine(line)
	}
}

func (l *Lexer) parseHeader(line string) Token {
	if !strings.HasPrefix(line, "tiny\t") {
		return l.errorToken(ErrInvalidHeader)
	}

	parts := strings.Split(line, Separator)
	if len(parts) < 5 {
		return l.errorToken(ErrHeaderTooShort)
	}

	l.namespaces = len(parts) - 4 // Calculate namespace count
	l.state = stateHeaderParsed

	return Token{
		Type:  TokenHeader,
		Parts: parts,
		Line:  l.lineNumber,
	}
}

func (l *Lexer) parsePossibleProperty(line string) Token {
	indent := countIndent(line)
	trimmed := strings.TrimLeft(line, Separator)

	// Check if we're still in properties section
	if indent == 1 {
		parts := strings.SplitN(trimmed, Separator, 3)
		if len(parts) < 2 {
			return l.errorToken(ErrInvalidProperty)
		}

		prop := Token{
			Type:   TokenProperty,
			Indent: indent,
			Line:   l.lineNumber,
			Parts:  parts,
		}
		return prop
	}

	// Transition to body parsing
	l.state = stateInBody
	return l.parseBodyLine(line)
}

func (l *Lexer) parseBodyLine(line string) Token {
	indent := countIndent(line)
	trimmed := strings.TrimLeft(line, Separator)
	if trimmed == "" {
		return l.errorToken(ErrEmptyLine)
	}

	parts := strings.Split(trimmed, Separator)
	if len(parts) == 0 {
		return l.errorToken(ErrEmptyLineIndent)
	}

	identifier := parts[0]
	tokenType := TokenError

	switch identifier {
	case "c":
		if indent == 0 {
			tokenType = TokenClass
		} else {
			tokenType = TokenComment
		}
	case "f":
		if indent == 1 {
			tokenType = TokenField
		}
	case "m":
		if indent == 1 {
			tokenType = TokenMethod
		}
	case "p":
		if indent == 2 {
			tokenType = TokenParameter
		}
	case "v":
		if indent == 2 {
			tokenType = TokenLocalVariable
		}
	default:
		// There are some discrepancy concerning comments, the safe bet it to
		// assume that any invalid identifier is a comment
		tokenType = TokenComment
	}

	if tokenType == TokenError {
		return l.errorToken(fmt.Sprintf(ErrInvalidIdentifier, identifier, indent))
	}

	return Token{
		Type:   tokenType,
		Indent: indent,
		Parts:  parts,
		Line:   l.lineNumber,
	}
}

func countIndent(line string) int {
	count := 0
	for _, c := range line {
		if c == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

func (l *Lexer) errorToken(err string) Token {
	return Token{
		Type:  TokenError,
		Line:  l.lineNumber,
		Error: err,
	}
}
