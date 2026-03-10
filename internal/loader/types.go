// Package loader scans the embedded apis/ filesystem, parses YAML definitions,
// and dynamically registers cobra commands for every product/API combination.
package loader

// APIDefinition represents a single YAML API definition file.
type APIDefinition struct {
	Name       string      `yaml:"name"`
	Product    string      `yaml:"product"`
	Use        string      `yaml:"use"`
	Short      string      `yaml:"short"`
	Long       string      `yaml:"long"`
	Examples   []Example   `yaml:"examples"`
	SDK        SDKInfo     `yaml:"sdk"`
	Parameters []Parameter `yaml:"parameters"`
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
	EnumValues   []string      `yaml:"enum-values"`
	ArrayStyle   string        `yaml:"array-style"`
	ObjectSchema []SchemaField `yaml:"object-schema"`
	ItemSchema   []SchemaField `yaml:"item-schema"`
}

// SchemaField describes a field inside an object or object-array parameter.
type SchemaField struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}
