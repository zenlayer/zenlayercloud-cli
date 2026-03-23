package loader

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/term"
)

const (
	helpIndent       = "       "
	helpDoubleIndent = "          "
	helpTripleIndent = "              "
	helpQuadIndent   = "                  "
	helpLineWidth    = 80
)

// GenerateHelp generates structured help documentation for an API command.
func GenerateHelp(def *APIDefinition, inheritedFlags *pflag.FlagSet) string {
	var b strings.Builder

	// NAME
	b.WriteString("NAME\n")
	b.WriteString(helpIndent)
	b.WriteString(def.Use)
	b.WriteString(" -\n\n")

	// DESCRIPTION
	b.WriteString("DESCRIPTION\n")
	b.WriteString(wrapText(def.Long, helpIndent, helpLineWidth))
	b.WriteString("\n")

	// SYNOPSIS
	b.WriteString("\nSYNOPSIS\n")
	b.WriteString(buildSynopsis(def))

	// OPTIONS
	b.WriteString("\nOPTIONS\n")
	for _, param := range def.Parameters {
		b.WriteString(buildOptionDoc(param))
	}

	// EXAMPLES
	if len(def.Examples) > 0 {
		b.WriteString("\nEXAMPLES\n")
		for _, ex := range def.Examples {
			if ex.Desc != "" {
				b.WriteString(helpIndent)
				b.WriteString(ex.Desc)
				b.WriteString("\n\n")
			}
			b.WriteString(helpDoubleIndent)
			b.WriteString(ex.Cmd)
			b.WriteString("\n\n")
		}
	}

	// GLOBAL OPTIONS
	b.WriteString("GLOBAL OPTIONS\n")
	b.WriteString(formatGlobalFlags(inheritedFlags))

	return b.String()
}

// buildSynopsis generates the SYNOPSIS section.
func buildSynopsis(def *APIDefinition) string {
	var b strings.Builder
	b.WriteString(helpDoubleIndent)
	b.WriteString(def.Use)
	b.WriteString("\n")

	for _, param := range def.Parameters {
		b.WriteString(helpDoubleIndent)
		if param.Required {
			b.WriteString(fmt.Sprintf("--%s <value>\n", param.Name))
		} else {
			b.WriteString(fmt.Sprintf("[--%s <value>]\n", param.Name))
		}
	}
	return b.String()
}

// buildOptionDoc generates documentation for a single parameter.
func buildOptionDoc(param Parameter) string {
	var b strings.Builder

	// Parameter header: --name (type)
	b.WriteString(helpIndent)
	b.WriteString("--")
	b.WriteString(param.Name)
	b.WriteString(" (")
	b.WriteString(mapTypeToDisplay(param.Type))
	b.WriteString(")\n")

	// Description
	b.WriteString(wrapText(param.Description, helpDoubleIndent, helpLineWidth))
	b.WriteString("\n")

	// Handle enum values for top-level enum parameters
	if param.Type == "enum" && len(param.EnumValues) > 0 {
		b.WriteString("\n")
		b.WriteString(helpDoubleIndent)
		b.WriteString("Possible values:\n\n")
		for _, opt := range param.EnumValues {
			b.WriteString(helpDoubleIndent)
			b.WriteString("o ")
			b.WriteString(opt.Value)
			if opt.Description != "" {
				b.WriteString(" - ")
				b.WriteString(opt.Description)
			}
			b.WriteString("\n\n")
		}
	}

	// Handle object type
	if param.Type == "object" && len(param.ObjectSchema) > 0 {
		b.WriteString("\n")
		b.WriteString(buildSchemaDoc(param.ObjectSchema, helpDoubleIndent))
		b.WriteString(buildShorthandSyntax(param.ObjectSchema, false, helpIndent))
		b.WriteString(buildJSONSyntax(param.ObjectSchema, false, helpIndent))
	}

	// Handle object-array type
	if param.Type == "object-array" && len(param.ItemSchema) > 0 {
		b.WriteString("\n")
		b.WriteString(helpDoubleIndent)
		b.WriteString("(structure)\n")
		b.WriteString(buildSchemaDoc(param.ItemSchema, helpTripleIndent))
		b.WriteString(buildShorthandSyntax(param.ItemSchema, true, helpIndent))
		b.WriteString(buildJSONSyntax(param.ItemSchema, true, helpIndent))
	}

	b.WriteString("\n")
	return b.String()
}

// buildSchemaDoc generates documentation for schema fields.
func buildSchemaDoc(schema []SchemaField, indent string) string {
	var b strings.Builder
	nextIndent := indent + "    "

	for _, field := range schema {
		b.WriteString(indent)
		b.WriteString(field.Name)
		b.WriteString(" -> (")
		b.WriteString(mapTypeToDisplay(field.Type))
		b.WriteString(")")
		if field.Required {
			b.WriteString(" [required]")
		}
		b.WriteString("\n")

		// Field description
		if field.Description != "" {
			b.WriteString(wrapText(field.Description, nextIndent, helpLineWidth))
			b.WriteString("\n")
		}

		// Enum values for schema fields
		if field.Type == "enum" && len(field.EnumValues) > 0 {
			b.WriteString("\n")
			b.WriteString(nextIndent)
			b.WriteString("Possible values:\n\n")
			for _, opt := range field.EnumValues {
				b.WriteString(nextIndent)
				b.WriteString("o ")
				b.WriteString(opt.Value)
				if opt.Description != "" {
					b.WriteString(" - ")
					b.WriteString(opt.Description)
				}
				b.WriteString("\n\n")
			}
		} else {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// buildShorthandSyntax generates the Shorthand Syntax section.
func buildShorthandSyntax(schema []SchemaField, isArray bool, indent string) string {
	var b strings.Builder
	b.WriteString(indent)
	b.WriteString("Shorthand Syntax:\n\n")
	b.WriteString(indent)
	b.WriteString("   ")

	var parts []string
	for _, field := range schema {
		parts = append(parts, fmt.Sprintf("%s=%s", field.Name, getTypePlaceholder(field.Type)))
	}
	b.WriteString(strings.Join(parts, ","))
	if isArray {
		b.WriteString(" ...")
	}
	b.WriteString("\n\n")
	return b.String()
}

// buildJSONSyntax generates the JSON Syntax section.
func buildJSONSyntax(schema []SchemaField, isArray bool, indent string) string {
	var b strings.Builder
	b.WriteString(indent)
	b.WriteString("JSON Syntax:\n\n")

	// Build example object
	obj := make(map[string]interface{})
	for _, field := range schema {
		obj[field.Name] = getJSONExample(field)
	}

	var jsonBytes []byte
	var err error
	if isArray {
		jsonBytes, err = json.MarshalIndent([]interface{}{obj}, indent+"   ", "  ")
	} else {
		jsonBytes, err = json.MarshalIndent(obj, indent+"   ", "  ")
	}
	if err != nil {
		return ""
	}

	// Add indent to each line
	lines := strings.Split(string(jsonBytes), "\n")
	for i, line := range lines {
		if i == 0 {
			b.WriteString(indent)
			b.WriteString("   ")
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// formatGlobalFlags formats inherited flags for display.
func formatGlobalFlags(flags *pflag.FlagSet) string {
	var b strings.Builder
	flags.VisitAll(func(f *pflag.Flag) {
		b.WriteString(helpIndent)
		b.WriteString("--")
		b.WriteString(f.Name)
		b.WriteString(" (")
		b.WriteString(f.Value.Type())
		b.WriteString(")\n")
		if f.Usage != "" {
			b.WriteString(wrapText(f.Usage, helpDoubleIndent, helpLineWidth))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	})
	return b.String()
}

// mapTypeToDisplay converts internal type names to display format.
func mapTypeToDisplay(typ string) string {
	switch typ {
	case "string":
		return "string"
	case "integer":
		return "integer"
	case "float":
		return "float"
	case "boolean":
		return "boolean"
	case "enum":
		return "string"
	case "string-array":
		return "list"
	case "integer-array":
		return "list"
	case "object":
		return "structure"
	case "object-array":
		return "list"
	default:
		return typ
	}
}

// getTypePlaceholder returns a placeholder string for a type.
func getTypePlaceholder(typ string) string {
	switch typ {
	case "string", "enum":
		return "string"
	case "integer":
		return "integer"
	case "float":
		return "float"
	case "boolean":
		return "true|false"
	default:
		return "value"
	}
}

// getJSONExample returns an example value for JSON syntax.
func getJSONExample(field SchemaField) interface{} {
	switch field.Type {
	case "string":
		return "string"
	case "enum":
		if len(field.EnumValues) > 0 {
			// Join enum values with |
			var values []string
			for _, opt := range field.EnumValues {
				values = append(values, opt.Value)
			}
			return strings.Join(values, "|")
		}
		return "string"
	case "integer":
		return 123
	case "float":
		return 1.5
	case "boolean":
		return true
	default:
		return "value"
	}
}

// wrapText wraps text to fit within the specified width, adding indent to each line.
func wrapText(text string, indent string, width int) string {
	if text == "" {
		return ""
	}

	// Handle multi-line text (preserve paragraph breaks)
	paragraphs := strings.Split(text, "\n")
	var result []string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			result = append(result, "")
			continue
		}
		result = append(result, wrapParagraph(para, indent, width))
	}

	return strings.Join(result, "\n")
}

// wrapParagraph wraps a single paragraph.
func wrapParagraph(text string, indent string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := indent

	for _, word := range words {
		if len(currentLine)+len(word)+1 > width && currentLine != indent {
			lines = append(lines, currentLine)
			currentLine = indent + word
		} else if currentLine == indent {
			currentLine = indent + word
		} else {
			currentLine += " " + word
		}
	}
	if currentLine != indent {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// OutputWithPager outputs the help text using a pager (like less) if the output
// is a terminal and the content is longer than the terminal height.
// Falls back to direct output if pager is not available or not a TTY.
func OutputWithPager(w io.Writer, content string) {
	// Check if output is a terminal
	file, ok := w.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		fmt.Fprint(w, content)
		return
	}

	// Get terminal height
	_, height, err := term.GetSize(int(file.Fd()))
	if err != nil {
		fmt.Fprint(w, content)
		return
	}

	// Count lines in content
	lineCount := strings.Count(content, "\n") + 1
	if lineCount <= height {
		fmt.Fprint(w, content)
		return
	}

	// Try to use pager
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	// Try less with options for better UX
	var cmd *exec.Cmd
	if pager == "less" {
		// -R: interpret ANSI color codes
		// -F: quit if content fits on one screen
		// No --mouse to allow normal text selection for copying
		cmd = exec.Command("less", "-R", "-F")
	} else {
		cmd = exec.Command(pager)
	}

	cmd.Stdout = file
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprint(w, content)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprint(w, content)
		return
	}

	_, _ = io.WriteString(stdin, content)
	stdin.Close()

	_ = cmd.Wait()
}
