// Package loader scans the embedded apis/ filesystem, parses YAML definitions,
// and dynamically registers cobra commands for every product/API combination.
package loader

import "gopkg.in/yaml.v3"

// APIDefinition represents a single YAML API definition file.
type APIDefinition struct {
	Name       string        `yaml:"name"`
	Product    string        `yaml:"product"`
	Use        string        `yaml:"use"`
	Short      string        `yaml:"short"`
	Long       string        `yaml:"long"`
	Examples   []Example     `yaml:"examples"`
	SDK        SDKInfo       `yaml:"sdk"`
	Parameters []Parameter   `yaml:"parameters"`
	Response   []SchemaField `yaml:"response"`
}

// Example is a CLI usage example shown in the command's help text.
type Example struct {
	Cmd  string `yaml:"cmd"`
	Desc string `yaml:"desc"`
}

// SDKInfo holds the service/version/action triple used to call the Zenlayer API.
type SDKInfo struct {
	Service string `yaml:"service"`
	Version string `yaml:"version"`
	Action  string `yaml:"action"`
}

// Parameter describes a single CLI flag derived from an API parameter.
type Parameter struct {
	Name         string        `yaml:"name"`
	Type         string        `yaml:"type"`
	Required     bool          `yaml:"required"`
	SDKField     string        `yaml:"sdk-field"`
	SDKWrapper   string        `yaml:"sdk-wrapper"`
	Description  string        `yaml:"description"`
	Sensitive    bool          `yaml:"sensitive"`
	Deprecated   bool          `yaml:"deprecated"`
	EnumValues   EnumOptions   `yaml:"enum-values"`
	ArrayStyle   string        `yaml:"array-style"`
	ObjectSchema []SchemaField `yaml:"object-schema"`
	ItemSchema   []SchemaField `yaml:"item-schema"`
}

// EnumOption represents an enum value with optional description.
type EnumOption struct {
	Value       string `yaml:"value"`
	Description string `yaml:"desc"`
}

// EnumOptions is a slice of EnumOption that supports both formats:
// - Simple: ["value1", "value2"]
// - With desc: [{value: "value1", desc: "desc1"}]
type EnumOptions []EnumOption

// UnmarshalYAML implements custom unmarshaling to support both formats.
func (e *EnumOptions) UnmarshalYAML(node *yaml.Node) error {
	// Try structured format first: [{value: "v1", desc: "d1"}]
	var structured []EnumOption
	if err := node.Decode(&structured); err == nil && len(structured) > 0 && structured[0].Value != "" {
		*e = structured
		return nil
	}
	// Fall back to simple string array: ["v1", "v2"]
	var simple []string
	if err := node.Decode(&simple); err != nil {
		return err
	}
	*e = make(EnumOptions, len(simple))
	for i, v := range simple {
		(*e)[i] = EnumOption{Value: v}
	}
	return nil
}

// Values returns a slice of just the enum values (without descriptions).
func (e EnumOptions) Values() []string {
	result := make([]string, len(e))
	for i, opt := range e {
		result[i] = opt.Value
	}
	return result
}

// Equal compares two EnumOptions for equality (values only).
func (e EnumOptions) Equal(other EnumOptions) bool {
	if len(e) != len(other) {
		return false
	}
	for i, opt := range e {
		if opt.Value != other[i].Value {
			return false
		}
	}
	return true
}

// SchemaField describes a field inside an object or object-array parameter.
type SchemaField struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Required    bool        `yaml:"required"`
	EnumValues  EnumOptions `yaml:"enum-values"`
}
