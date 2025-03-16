package tiny_v2_parser

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed mappings.tiny
var input string

func TestParser_Parse(t *testing.T) {
	parser := NewParser(strings.NewReader(input))
	ast, errors := parser.Parse()

	if len(errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range errors {
			fmt.Printf("Line %d: %s\n", err.Line, err.Message)
		}

		t.Fail()
	}

	fmt.Println("Classes:")
	for _, class := range ast.Classes {
		fmt.Printf("- %s → %s\n", class.Names[0], class.Names[1])
		fmt.Println("  Comments:", class.Comments)

		for _, method := range class.Methods {
			fmt.Printf("  Method: %s → %s\n", method.Names[0], method.Names[1])
			fmt.Println("    Parameters:", method.Parameters)
		}
	}
}

func TestParser_SimpleMap(t *testing.T) {
	result := make(map[string]string)

	parser := NewParser(strings.NewReader(input))
	ast, errors := parser.Parse()

	if len(errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range errors {
			fmt.Printf("Line %d: %s\n", err.Line, err.Message)
		}

		t.Fail()
	}

	for _, class := range ast.Classes {
		for _, method := range class.Methods {
			result[method.Names[0]] = method.Names[1]
		}

		for _, field := range class.Fields {
			result[field.Names[0]] = field.Names[1]
		}
	}

	fmt.Println(result)
}
