package tiny_v2_parser

// TokenType represents different elements in the Tiny V2 format
type TokenType int

const (
	TokenHeader TokenType = iota
	TokenProperty
	TokenClass
	TokenField
	TokenMethod
	TokenParameter
	TokenLocalVariable
	TokenComment
	TokenError
	TokenEOF
)

// Token represents a parsed element with metadata
type Token struct {
	Type   TokenType
	Indent int      // Number of tabs indentation
	Parts  []string // Split parts of the line
	Line   int      // Line number in source
	Error  string
}
