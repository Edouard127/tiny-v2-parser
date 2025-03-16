package tiny_v2_parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser Implementation
type Parser struct {
	lexer      *Lexer
	currentTok Token
	errors     []ParseError
	metadata   Metadata
	classes    []*Class
	stack      []*parseNode
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		lexer:  NewLexer(r),
		errors: make([]ParseError, 0),
		stack:  make([]*parseNode, 0),
	}
}

func (p *Parser) Parse() (*Mapping, []ParseError) {
	tok := p.lexer.NextToken()
	if err := p.parseHeader(tok); err != nil {
		return nil, p.errors
	}

	for {
		tok = p.lexer.NextToken()
		if tok.Type == TokenEOF {
			break
		}
		p.currentTok = tok
		p.adjustStack(tok.Indent)

		p.parseBody(tok)
	}

	return &Mapping{
		Metadata: p.metadata,
		Classes:  p.classes,
	}, p.errors
}

func (p *Parser) adjustStack(currentIndent int) {
	for len(p.stack) > 0 && p.stack[len(p.stack)-1].indent >= currentIndent {
		p.stack = p.stack[:len(p.stack)-1]
	}
}

func (p *Parser) parseHeader(tok Token) error {
	if tok.Type != TokenHeader {
		p.addError("missing header")
		return fmt.Errorf("invalid format")
	}

	parts := tok.Parts
	if len(parts) < 5 {
		p.addError("invalid header format")
		return fmt.Errorf("invalid header")
	}

	var err error
	p.metadata.MajorVersion, err = strconv.Atoi(parts[1])
	if err != nil {
		p.addError("invalid major version")
	}
	p.metadata.MinorVersion, err = strconv.Atoi(parts[2])
	if err != nil {
		p.addError("invalid minor version")
	}

	// Parse namespaces
	p.metadata.Namespaces = parts[3:]
	p.metadata.Properties = make(map[string]string)

	for {
		tok = p.lexer.NextToken()
		if tok.Indent == 0 {
			// Hack, if there are no properties, change the current token
			// and pass the token argument to the parseBody
			p.currentTok = tok
			p.parseBody(tok)
			return nil
		}

		if len(tok.Parts) < 2 {
			p.addError(tok.Error)
			continue
		}

		var value string
		if len(tok.Parts) >= 3 {
			value = strings.Join(tok.Parts[2:], "\t")
		}

		p.metadata.Properties[tok.Parts[1]] = value
	}

	return nil
}

func (p *Parser) parseBody(tok Token) {
	switch tok.Type {
	case TokenClass:
		p.parseClass()
	case TokenField:
		p.parseField()
	case TokenMethod:
		p.parseMethod()
	case TokenParameter:
		p.parseParameter()
	case TokenLocalVariable:
		p.parseLocalVariable()
	case TokenComment:
		p.handleComment()
	case TokenError:
		p.addError(tok.Error)
	}
}

func (p *Parser) parseClass() {
	if p.currentTok.Indent != 0 {
		p.addError("class must be at root level")
		return
	}

	class := &Class{
		Names:    p.parseNames(p.currentTok.Parts[1:]),
		Comments: make([]string, 0),
	}

	p.classes = append(p.classes, class)
	p.pushStack(class, p.currentTok.Indent)
}

func (p *Parser) parseMethod() {
	parent := p.getCurrentParent()
	if parent == nil || p.currentTok.Indent != 1 {
		p.addError("method must be inside a class")
		return
	}

	method := &Method{
		Descriptor: p.unescape(p.currentTok.Parts[1]),
		Names:      p.parseNames(p.currentTok.Parts[2:]),
		Comments:   make([]string, 0),
	}

	class, ok := parent.(*Class)
	if !ok {
		p.addError("method parent must be a class")
		return
	}

	class.Methods = append(class.Methods, method)
	p.pushStack(method, p.currentTok.Indent)
}

func (p *Parser) parseField() {
	parent := p.getCurrentParent()
	if parent == nil || p.currentTok.Indent != 1 {
		p.addError("field must be inside a class")
		return
	}

	field := &Field{
		Descriptor: p.unescape(p.currentTok.Parts[1]),
		Names:      p.parseNames(p.currentTok.Parts[2:]),
		Comments:   make([]string, 0),
	}

	class, ok := parent.(*Class)
	if !ok {
		p.addError("field parent must be a class")
		return
	}

	class.Fields = append(class.Fields, field)
	p.pushStack(field, p.currentTok.Indent)
}

func (p *Parser) parseParameter() {
	parent := p.getCurrentParent()
	if parent == nil || p.currentTok.Indent != 2 {
		p.addError("parameter must be inside a method")
		return
	}

	index, err := strconv.Atoi(p.currentTok.Parts[1])
	if err != nil {
		p.addError("invalid parameter index")
		return
	}

	param := &Parameter{
		Index:    index,
		Names:    p.parseNames(p.currentTok.Parts[2:]),
		Comments: make([]string, 0),
	}

	method, ok := parent.(*Method)
	if !ok {
		p.addError("parameter parent must be a method")
		return
	}

	method.Parameters = append(method.Parameters, param)
	p.pushStack(param, p.currentTok.Indent)
}

func (p *Parser) parseLocalVariable() {
	parent := p.getCurrentParent()
	if parent == nil || p.currentTok.Indent != 2 {
		p.addError("local variable must be inside a method")
		return
	}

	if len(p.currentTok.Parts) < 4 {
		p.addError("invalid local variable format")
		return
	}

	index, err1 := strconv.Atoi(p.currentTok.Parts[1])
	startOffset, err2 := strconv.Atoi(p.currentTok.Parts[2])
	lvtIndex, err3 := strconv.Atoi(p.currentTok.Parts[3])

	if err1 != nil || err2 != nil || err3 != nil {
		p.addError("invalid local variable numbers")
		return
	}

	localVar := &LocalVariable{
		Index:       index,
		StartOffset: startOffset,
		LvtIndex:    lvtIndex,
		Names:       p.parseNames(p.currentTok.Parts[4:]),
		Comments:    make([]string, 0),
	}

	method, ok := parent.(*Method)
	if !ok {
		p.addError("local variable parent must be a method")
		return
	}

	method.LocalVars = append(method.LocalVars, localVar)
	p.pushStack(localVar, p.currentTok.Indent)
}

func (p *Parser) handleComment() {
	// Find appropriate parent
	for i := len(p.stack) - 1; i >= 0; i-- {
		if p.stack[i].indent == p.currentTok.Indent-1 {
			switch elem := p.stack[i].element.(type) {
			case *Class:
				elem.Comments = append(elem.Comments, strings.Join(p.currentTok.Parts, "\t"))
			case *Field:
				elem.Comments = append(elem.Comments, strings.Join(p.currentTok.Parts, "\t"))
			case *Method:
				elem.Comments = append(elem.Comments, strings.Join(p.currentTok.Parts, "\t"))
			case *Parameter:
				elem.Comments = append(elem.Comments, strings.Join(p.currentTok.Parts, "\t"))
			case *LocalVariable:
				elem.Comments = append(elem.Comments, strings.Join(p.currentTok.Parts, "\t"))
			}
			return
		}
	}

	// Global comment
	p.metadata.GlobalComments = append(p.metadata.GlobalComments,
		fmt.Sprintf("L%d: %s", p.currentTok.Line, strings.Join(p.currentTok.Parts, "\t")))
}

func (p *Parser) parseNames(parts []string) []string {
	names := make([]string, len(p.metadata.Namespaces))
	for i := range p.metadata.Namespaces {
		if i < len(parts) {
			names[i] = p.unescape(parts[i])
		}
	}
	return names
}

func (p *Parser) unescape(s string) string {
	if p.metadata.Properties["escaped-names"] != "" {
		return unescapeString(s)
	}
	return s
}

func (p *Parser) pushStack(element any, indent int) {
	p.stack = append(p.stack, &parseNode{
		indent:  indent,
		element: element,
	})
}

func (p *Parser) getCurrentParent() any {
	if len(p.stack) == 0 {
		return nil
	}
	return p.stack[len(p.stack)-1].element
}

func (p *Parser) addError(err string) {
	p.errors = append(p.errors, ParseError{
		Line:    p.currentTok.Line,
		Message: err,
	})
}

func (p *Parser) getClasses() []*Class {
	return p.classes
}

func unescapeString(s string) string {
	// Implement actual unescaping logic based on TinyV2 spec
	return strings.ReplaceAll(s, "\\n", "\n")
}

type Mapping struct {
	Metadata Metadata
	Classes  []*Class
}

type Metadata struct {
	MajorVersion   int
	MinorVersion   int
	Namespaces     []string
	Properties     map[string]string
	GlobalComments []string
}

type Class struct {
	Names    []string
	Comments []string
	Fields   []*Field
	Methods  []*Method
}

type Field struct {
	Descriptor string
	Names      []string
	Comments   []string
}

type Method struct {
	Descriptor string
	Names      []string
	Comments   []string
	Parameters []*Parameter
	LocalVars  []*LocalVariable
}

type Parameter struct {
	Index    int
	Names    []string
	Comments []string
}

type LocalVariable struct {
	Index       int
	StartOffset int
	LvtIndex    int
	Names       []string
	Comments    []string
}

type parseNode struct {
	indent  int
	element any
}

type ParseError struct {
	Line    int
	Message string
}
