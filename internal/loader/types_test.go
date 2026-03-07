package loader

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnumOptions_UnmarshalYAML_StructuredFormat(t *testing.T) {
	yamlData := `
- value: Active
  desc: Resource is active
- value: Inactive
  desc: Resource is inactive
`
	var opts EnumOptions
	if err := yaml.Unmarshal([]byte(yamlData), &opts); err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	if len(opts) != 2 {
		t.Fatalf("len(opts) = %d, want 2", len(opts))
	}
	if opts[0].Value != "Active" {
		t.Errorf("opts[0].Value = %q, want 'Active'", opts[0].Value)
	}
	if opts[0].Description != "Resource is active" {
		t.Errorf("opts[0].Description = %q, want 'Resource is active'", opts[0].Description)
	}
	if opts[1].Value != "Inactive" {
		t.Errorf("opts[1].Value = %q, want 'Inactive'", opts[1].Value)
	}
}

func TestEnumOptions_UnmarshalYAML_SimpleFormat(t *testing.T) {
	yamlData := `
- Active
- Inactive
- Pending
`
	var opts EnumOptions
	if err := yaml.Unmarshal([]byte(yamlData), &opts); err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	if len(opts) != 3 {
		t.Fatalf("len(opts) = %d, want 3", len(opts))
	}
	if opts[0].Value != "Active" {
		t.Errorf("opts[0].Value = %q, want 'Active'", opts[0].Value)
	}
	if opts[0].Description != "" {
		t.Errorf("opts[0].Description = %q, want ''", opts[0].Description)
	}
	if opts[2].Value != "Pending" {
		t.Errorf("opts[2].Value = %q, want 'Pending'", opts[2].Value)
	}
}

func TestEnumOptions_UnmarshalYAML_InlineSimple(t *testing.T) {
	yamlData := `[Active, Inactive, Pending]`
	var opts EnumOptions
	if err := yaml.Unmarshal([]byte(yamlData), &opts); err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	if len(opts) != 3 {
		t.Fatalf("len(opts) = %d, want 3", len(opts))
	}
	if opts[0].Value != "Active" {
		t.Errorf("opts[0].Value = %q, want 'Active'", opts[0].Value)
	}
}

func TestSchemaField_WithEnumValues(t *testing.T) {
	yamlData := `
name: status
type: enum
description: Resource status
enum-values:
  - value: Active
    desc: Resource is active
  - value: Inactive
    desc: Resource is inactive
`
	var field SchemaField
	if err := yaml.Unmarshal([]byte(yamlData), &field); err != nil {
		t.Fatalf("Unmarshal SchemaField error = %v", err)
	}

	if field.Name != "status" {
		t.Errorf("field.Name = %q, want 'status'", field.Name)
	}
	if field.Type != "enum" {
		t.Errorf("field.Type = %q, want 'enum'", field.Type)
	}
	if len(field.EnumValues) != 2 {
		t.Fatalf("len(field.EnumValues) = %d, want 2", len(field.EnumValues))
	}
	if field.EnumValues[0].Value != "Active" {
		t.Errorf("field.EnumValues[0].Value = %q, want 'Active'", field.EnumValues[0].Value)
	}
	if field.EnumValues[0].Description != "Resource is active" {
		t.Errorf("field.EnumValues[0].Description = %q, want 'Resource is active'", field.EnumValues[0].Description)
	}
}

func TestSchemaField_WithRequired(t *testing.T) {
	yamlData := `
name: cidrBlock
type: string
required: true
description: CIDR block
`
	var field SchemaField
	if err := yaml.Unmarshal([]byte(yamlData), &field); err != nil {
		t.Fatalf("Unmarshal SchemaField error = %v", err)
	}

	if !field.Required {
		t.Error("field.Required = false, want true")
	}
}

func TestSchemaField_BackwardCompatibility(t *testing.T) {
	yamlData := `
name: networkType
type: enum
enum-values: [PremiumBGP, StandardBGP, CN2]
description: Network type
`
	var field SchemaField
	if err := yaml.Unmarshal([]byte(yamlData), &field); err != nil {
		t.Fatalf("Unmarshal SchemaField error = %v", err)
	}

	if len(field.EnumValues) != 3 {
		t.Fatalf("len(field.EnumValues) = %d, want 3", len(field.EnumValues))
	}
	if field.EnumValues[0].Value != "PremiumBGP" {
		t.Errorf("field.EnumValues[0].Value = %q, want 'PremiumBGP'", field.EnumValues[0].Value)
	}
	if field.EnumValues[0].Description != "" {
		t.Errorf("field.EnumValues[0].Description = %q, want '' (empty)", field.EnumValues[0].Description)
	}
}
