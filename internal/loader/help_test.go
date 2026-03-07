package loader

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestGenerateHelp(t *testing.T) {
	def := &APIDefinition{
		Use:   "create-disks",
		Short: "Creates one or more disks",
		Long:  "Create one or more disks with specified configurations.",
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string", Required: true, Description: "ID of availability zone."},
			{Name: "disk-name", Type: "string", Required: true, Description: "Disk name."},
			{Name: "disk-size", Type: "integer", Required: true, Description: "Storage space in GB."},
		},
	}

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("profile", "", "profile name")

	help := GenerateHelp(def, flags)

	// Check NAME section
	if !strings.Contains(help, "NAME") {
		t.Error("help should contain NAME section")
	}
	if !strings.Contains(help, "create-disks -") {
		t.Error("help should contain command name")
	}

	// Check DESCRIPTION section
	if !strings.Contains(help, "DESCRIPTION") {
		t.Error("help should contain DESCRIPTION section")
	}
	if !strings.Contains(help, "Create one or more disks") {
		t.Error("help should contain long description")
	}

	// Check SYNOPSIS section
	if !strings.Contains(help, "SYNOPSIS") {
		t.Error("help should contain SYNOPSIS section")
	}
	if !strings.Contains(help, "--zone-id <value>") {
		t.Error("help should contain required parameter in synopsis")
	}

	// Check OPTIONS section
	if !strings.Contains(help, "OPTIONS") {
		t.Error("help should contain OPTIONS section")
	}
	if !strings.Contains(help, "--zone-id (string)") {
		t.Error("help should contain parameter with type")
	}

	// Check GLOBAL OPTIONS section
	if !strings.Contains(help, "GLOBAL OPTIONS") {
		t.Error("help should contain GLOBAL OPTIONS section")
	}
}

func TestGenerateHelp_ObjectParameter(t *testing.T) {
	def := &APIDefinition{
		Use:   "create-instance",
		Short: "Creates an instance",
		Long:  "Create an instance.",
		Parameters: []Parameter{
			{
				Name:        "system-disk",
				Type:        "object",
				Description: "Boot disk configuration.",
				ObjectSchema: []SchemaField{
					{Name: "diskCategory", Type: "string", Description: "Disk type."},
					{Name: "diskSize", Type: "integer", Description: "Disk size in GB."},
				},
			},
		},
	}

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	help := GenerateHelp(def, flags)

	// Check structure display
	if !strings.Contains(help, "--system-disk (structure)") {
		t.Error("help should show object as (structure)")
	}

	// Check field documentation
	if !strings.Contains(help, "diskCategory -> (string)") {
		t.Error("help should contain field with arrow notation")
	}
	if !strings.Contains(help, "diskSize -> (integer)") {
		t.Error("help should contain diskSize field")
	}

	// Check Shorthand Syntax
	if !strings.Contains(help, "Shorthand Syntax:") {
		t.Error("help should contain Shorthand Syntax section")
	}
	if !strings.Contains(help, "diskCategory=string") {
		t.Error("help should contain shorthand example")
	}

	// Check JSON Syntax
	if !strings.Contains(help, "JSON Syntax:") {
		t.Error("help should contain JSON Syntax section")
	}
}

func TestGenerateHelp_ObjectArrayParameter(t *testing.T) {
	def := &APIDefinition{
		Use:   "create-tags",
		Short: "Creates tags",
		Long:  "Create tags for resources.",
		Parameters: []Parameter{
			{
				Name:        "tags",
				Type:        "object-array",
				Description: "Tags to apply.",
				ItemSchema: []SchemaField{
					{Name: "key", Type: "string", Description: "Tag key."},
					{Name: "value", Type: "string", Description: "Tag value."},
				},
			},
		},
	}

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	help := GenerateHelp(def, flags)

	// Check list display
	if !strings.Contains(help, "--tags (list)") {
		t.Error("help should show object-array as (list)")
	}

	// Check (structure) indicator
	if !strings.Contains(help, "(structure)") {
		t.Error("help should contain (structure) for object-array items")
	}

	// Check field documentation
	if !strings.Contains(help, "key -> (string)") {
		t.Error("help should contain key field")
	}
	if !strings.Contains(help, "value -> (string)") {
		t.Error("help should contain value field")
	}

	// Check shorthand with ... for arrays
	if !strings.Contains(help, "key=string,value=string ...") {
		t.Error("help should contain shorthand with ... for arrays")
	}
}

func TestGenerateHelp_RequiredFields(t *testing.T) {
	def := &APIDefinition{
		Use:   "test-cmd",
		Short: "Test command",
		Long:  "Test command description.",
		Parameters: []Parameter{
			{
				Name:        "config",
				Type:        "object",
				Description: "Configuration.",
				ObjectSchema: []SchemaField{
					{Name: "name", Type: "string", Description: "Name.", Required: true},
					{Name: "value", Type: "string", Description: "Value."},
				},
			},
		},
	}

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	help := GenerateHelp(def, flags)

	// Check [required] marker
	if !strings.Contains(help, "name -> (string) [required]") {
		t.Error("help should show [required] for required fields")
	}
}

func TestGenerateHelp_EnumField(t *testing.T) {
	def := &APIDefinition{
		Use:   "test-cmd",
		Short: "Test command",
		Long:  "Test command description.",
		Parameters: []Parameter{
			{
				Name:        "config",
				Type:        "object",
				Description: "Configuration.",
				ObjectSchema: []SchemaField{
					{
						Name:        "status",
						Type:        "enum",
						Description: "Status.",
						EnumValues: EnumOptions{
							{Value: "Active", Description: "Resource is active"},
							{Value: "Inactive", Description: "Resource is inactive"},
						},
					},
				},
			},
		},
	}

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	help := GenerateHelp(def, flags)

	// Check Possible values
	if !strings.Contains(help, "Possible values:") {
		t.Error("help should contain Possible values section")
	}
	if !strings.Contains(help, "Active - Resource is active") {
		t.Error("help should contain enum value with description")
	}
	if !strings.Contains(help, "Inactive - Resource is inactive") {
		t.Error("help should contain all enum values")
	}
}

func TestBuildSynopsis(t *testing.T) {
	def := &APIDefinition{
		Use: "test-cmd",
		Parameters: []Parameter{
			{Name: "required-param", Type: "string", Required: true},
			{Name: "optional-param", Type: "string", Required: false},
		},
	}

	synopsis := buildSynopsis(def)

	// Required params should not have brackets
	if !strings.Contains(synopsis, "--required-param <value>") {
		t.Error("required param should not have brackets")
	}

	// Optional params should have brackets
	if !strings.Contains(synopsis, "[--optional-param <value>]") {
		t.Error("optional param should have brackets")
	}
}

func TestMapTypeToDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"string", "string"},
		{"integer", "integer"},
		{"float", "float"},
		{"boolean", "boolean"},
		{"enum", "string"},
		{"string-array", "list"},
		{"integer-array", "list"},
		{"object", "structure"},
		{"object-array", "list"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		got := mapTypeToDisplay(tt.input)
		if got != tt.want {
			t.Errorf("mapTypeToDisplay(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWrapText(t *testing.T) {
	t.Run("short text unchanged", func(t *testing.T) {
		result := wrapText("short text", "  ", 80)
		if result != "  short text" {
			t.Errorf("wrapText() = %q, want '  short text'", result)
		}
	})

	t.Run("long text wraps", func(t *testing.T) {
		longText := "This is a very long text that should be wrapped to fit within the specified width limit."
		result := wrapText(longText, "  ", 40)
		lines := strings.Split(result, "\n")
		if len(lines) < 2 {
			t.Error("long text should wrap to multiple lines")
		}
		for _, line := range lines {
			if len(line) > 40 {
				t.Errorf("line too long: %q (len=%d)", line, len(line))
			}
		}
	})

	t.Run("empty text", func(t *testing.T) {
		result := wrapText("", "  ", 80)
		if result != "" {
			t.Errorf("wrapText('') = %q, want ''", result)
		}
	})

	t.Run("preserves newlines", func(t *testing.T) {
		text := "Line one.\nLine two."
		result := wrapText(text, "  ", 80)
		if !strings.Contains(result, "\n") {
			t.Error("should preserve newlines")
		}
	})
}

func TestGetTypePlaceholder(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"string", "string"},
		{"enum", "string"},
		{"integer", "integer"},
		{"float", "float"},
		{"boolean", "true|false"},
		{"unknown", "value"},
	}

	for _, tt := range tests {
		got := getTypePlaceholder(tt.input)
		if got != tt.want {
			t.Errorf("getTypePlaceholder(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
