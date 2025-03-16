package tiny_v2_parser

var (
	ErrIo                = "IO error: %v"
	ErrHeaderTooShort    = "header too short"
	ErrInvalidHeader     = "invalid header"
	ErrInvalidProperty   = "invalid property format"
	ErrEmptyLine         = "empty line"
	ErrEmptyLineIndent   = "empty line after indent"
	ErrInvalidIdentifier = "invalid identifier %q at indent %d"
)
