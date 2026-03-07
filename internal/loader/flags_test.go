package loader

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewFlagStore(t *testing.T) {
	store := newFlagStore()
	if store.strings == nil {
		t.Error("strings map is nil")
	}
	if store.stringArrays == nil {
		t.Error("stringArrays map is nil")
	}
	if store.ints == nil {
		t.Error("ints map is nil")
	}
	if store.intArrays == nil {
		t.Error("intArrays map is nil")
	}
	if store.floats == nil {
		t.Error("floats map is nil")
	}
	if store.bools == nil {
		t.Error("bools map is nil")
	}
	if store.arrayFlagNames == nil {
		t.Error("arrayFlagNames is nil")
	}
}

func TestBuildFlagDescription(t *testing.T) {
	tests := []struct {
		name  string
		param Parameter
		want  string
	}{
		{
			name:  "plain string",
			param: Parameter{Type: "string", Description: "A string param"},
			want:  "A string param",
		},
		{
			name:  "enum with values",
			param: Parameter{Type: "enum", Description: "A mode", EnumValues: EnumOptions{{Value: "fast"}, {Value: "slow"}}},
			want:  "A mode (fast|slow)",
		},
		{
			name:  "enum without values",
			param: Parameter{Type: "enum", Description: "A mode"},
			want:  "A mode",
		},
		{
			name:  "string-array",
			param: Parameter{Type: "string-array", Description: "Instance IDs"},
			want:  "Instance IDs (one or more strings separated by spaces, quote items containing spaces)",
		},
		{
			name:  "integer-array",
			param: Parameter{Type: "integer-array", Description: "Port numbers"},
			want:  "Port numbers (one or more integers separated by spaces)",
		},
		{
			name: "object with empty ArrayStyle (default kv)",
			param: Parameter{
				Type:        "object",
				Description: "Disk config",
				ArrayStyle:  "",
				ObjectSchema: []SchemaField{
					{Name: "size", Type: "integer"},
					{Name: "category", Type: "string"},
				},
			},
			want: "Disk config (e.g. size=<int>,category=<value>)",
		},
		{
			name: "object with explicit kv style",
			param: Parameter{
				Type:        "object",
				Description: "Disk config",
				ArrayStyle:  "kv",
				ObjectSchema: []SchemaField{
					{Name: "ratio", Type: "float"},
				},
			},
			want: "Disk config (e.g. ratio=<float>)",
		},
		{
			name: "object with no schema fields",
			param: Parameter{
				Type:         "object",
				Description:  "Config",
				ObjectSchema: nil,
			},
			want: "Config",
		},
		{
			name: "object-array with empty ArrayStyle",
			param: Parameter{
				Type:        "object-array",
				Description: "Disk list",
				ArrayStyle:  "",
				ItemSchema: []SchemaField{
					{Name: "size", Type: "integer"},
				},
			},
			want: "Disk list (e.g. size=<int>)",
		},
		{
			name: "object-array with kv style and boolean field",
			param: Parameter{
				Type:        "object-array",
				Description: "Flags",
				ArrayStyle:  "kv",
				ItemSchema: []SchemaField{
					{Name: "enabled", Type: "boolean"},
				},
			},
			want: "Flags (e.g. enabled=<bool>)",
		},
		{
			name:  "sensitive param",
			param: Parameter{Type: "string", Description: "Password", Sensitive: true},
			want:  "Password (sensitive)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFlagDescription(&tt.param)
			if got != tt.want {
				t.Errorf("buildFlagDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractSchemaKeys(t *testing.T) {
	fields := []SchemaField{
		{Name: "diskCategory"},
		{Name: "diskSize"},
	}
	keys := extractSchemaKeys(fields)
	if len(keys) != 2 {
		t.Fatalf("len(keys) = %d, want 2", len(keys))
	}
	if keys[0] != "diskCategory=" {
		t.Errorf("keys[0] = %q, want 'diskCategory='", keys[0])
	}
	if keys[1] != "diskSize=" {
		t.Errorf("keys[1] = %q, want 'diskSize='", keys[1])
	}
}

func TestExtractSchemaKeys_Empty(t *testing.T) {
	keys := extractSchemaKeys(nil)
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}
}

func TestObjectExampleParts(t *testing.T) {
	fields := []SchemaField{
		{Name: "name", Type: "string"},
		{Name: "count", Type: "integer"},
		{Name: "ratio", Type: "float"},
		{Name: "enabled", Type: "boolean"},
		{Name: "other", Type: "unknown"},
	}
	parts := objectExampleParts(fields)
	expected := []string{"name=<value>", "count=<int>", "ratio=<float>", "enabled=<bool>", "other=<value>"}
	if len(parts) != len(expected) {
		t.Fatalf("len(parts) = %d, want %d", len(parts), len(expected))
	}
	for i, want := range expected {
		if parts[i] != want {
			t.Errorf("parts[%d] = %q, want %q", i, parts[i], want)
		}
	}
}

func TestObjectExampleParts_Empty(t *testing.T) {
	parts := objectExampleParts(nil)
	if len(parts) != 0 {
		t.Errorf("expected empty parts, got %v", parts)
	}
}

func TestNoFileComp(t *testing.T) {
	completions, directive := noFileComp(nil, nil, "")
	if completions != nil {
		t.Errorf("noFileComp() completions = %v, want nil", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("noFileComp() directive = %v, want NoFileComp", directive)
	}
}

func TestSchemaKeyComp(t *testing.T) {
	keys := []string{"name=", "size="}
	comp := schemaKeyComp(keys)
	completions, directive := comp(nil, nil, "")
	if len(completions) != 2 {
		t.Fatalf("len(completions) = %d, want 2", len(completions))
	}
	if completions[0] != "name=" {
		t.Errorf("completions[0] = %q, want 'name='", completions[0])
	}
	if directive != cobra.ShellCompDirectiveNoSpace {
		t.Errorf("directive = %v, want NoSpace", directive)
	}
}

func TestSchemaKeyComp_Empty(t *testing.T) {
	comp := schemaKeyComp(nil)
	completions, directive := comp(nil, nil, "")
	if len(completions) != 0 {
		t.Errorf("expected empty completions, got %v", completions)
	}
	if directive != cobra.ShellCompDirectiveNoSpace {
		t.Errorf("directive = %v, want NoSpace", directive)
	}
}

func TestBindFlags_AllTypes(t *testing.T) {
	params := []Parameter{
		{Name: "zone-id", Type: "string", Description: "Zone"},
		{Name: "mode", Type: "enum", EnumValues: EnumOptions{{Value: "auto"}, {Value: "manual"}}, Description: "Mode"},
		{Name: "tags", Type: "string-array", Description: "Tags"},
		{Name: "count", Type: "integer", Description: "Count"},
		{Name: "ports", Type: "integer-array", Description: "Ports"},
		{Name: "ratio", Type: "float", Description: "Ratio"},
		{Name: "enabled", Type: "boolean", Description: "Enabled"},
		{Name: "disk", Type: "object", Description: "Disk", ObjectSchema: []SchemaField{{Name: "size", Type: "integer"}}},
		{Name: "disks", Type: "object-array", Description: "Disks", ItemSchema: []SchemaField{{Name: "size", Type: "integer"}}},
	}
	def := &APIDefinition{Parameters: params}
	cmd := &cobra.Command{Use: "test"}
	store := bindFlags(cmd, def)

	// Verify all flags are registered on the command.
	for _, p := range params {
		if cmd.Flags().Lookup(p.Name) == nil {
			t.Errorf("flag --%s not registered", p.Name)
		}
	}

	// Verify store maps contain the correct entries.
	if _, ok := store.strings["zone-id"]; !ok {
		t.Error("store.strings missing 'zone-id'")
	}
	if _, ok := store.strings["mode"]; !ok {
		t.Error("store.strings missing 'mode' (enum stored as string)")
	}
	if _, ok := store.stringArrays["tags"]; !ok {
		t.Error("store.stringArrays missing 'tags'")
	}
	if _, ok := store.ints["count"]; !ok {
		t.Error("store.ints missing 'count'")
	}
	if _, ok := store.intArrays["ports"]; !ok {
		t.Error("store.intArrays missing 'ports'")
	}
	if _, ok := store.floats["ratio"]; !ok {
		t.Error("store.floats missing 'ratio'")
	}
	if _, ok := store.bools["enabled"]; !ok {
		t.Error("store.bools missing 'enabled'")
	}
	if _, ok := store.strings["disk"]; !ok {
		t.Error("store.strings missing 'disk' (object stored as string)")
	}
	if _, ok := store.stringArrays["disks"]; !ok {
		t.Error("store.stringArrays missing 'disks' (object-array stored as []string)")
	}
}

func TestBindFlag_UnknownType(t *testing.T) {
	// Unknown types should not panic; just no flag is registered.
	param := Parameter{Name: "mystery", Type: "unknown-type", Description: "?"}
	cmd := &cobra.Command{Use: "test"}
	store := newFlagStore()
	bindFlag(cmd, &param, store)
	if cmd.Flags().Lookup("mystery") != nil {
		t.Error("expected no flag registered for unknown type")
	}
}

func TestBindFlag_Deprecated(t *testing.T) {
	param := Parameter{
		Name:        "old-flag",
		Type:        "string",
		Description: "Legacy flag",
		Deprecated:  true,
	}
	cmd := &cobra.Command{Use: "test"}
	store := newFlagStore()
	bindFlag(cmd, &param, store)

	flag := cmd.Flags().Lookup("old-flag")
	if flag == nil {
		t.Fatal("flag --old-flag not registered")
	}
	// cobra marks deprecated flags with a non-empty Deprecated field.
	if flag.Deprecated == "" {
		t.Error("expected flag.Deprecated to be non-empty for a deprecated flag")
	}
}

func TestBindFlag_EnumCompletion(t *testing.T) {
	param := Parameter{
		Name:       "nic-type",
		Type:       "enum",
		EnumValues: EnumOptions{{Value: "Auto"}, {Value: "Manual"}},
		Description: "NIC type",
	}
	cmd := &cobra.Command{Use: "test"}
	store := newFlagStore()
	bindFlag(cmd, &param, store)

	// Verify the flag is registered.
	if cmd.Flags().Lookup("nic-type") == nil {
		t.Fatal("flag --nic-type not registered")
	}
	// Verify the enum values appear in the flag description.
	flag := cmd.Flags().Lookup("nic-type")
	if !strings.Contains(flag.Usage, "Auto") {
		t.Errorf("flag usage %q should contain enum values", flag.Usage)
	}
}

func TestBindFlag_ObjectSchemaCompletion(t *testing.T) {
	param := Parameter{
		Name:        "system-disk",
		Type:        "object",
		Description: "System disk",
		ObjectSchema: []SchemaField{
			{Name: "diskCategory", Type: "string"},
			{Name: "diskSize", Type: "integer"},
		},
	}
	cmd := &cobra.Command{Use: "test"}
	store := newFlagStore()
	bindFlag(cmd, &param, store)

	if cmd.Flags().Lookup("system-disk") == nil {
		t.Fatal("flag --system-disk not registered")
	}
}

func TestBindFlag_ObjectArraySchemaCompletion(t *testing.T) {
	param := Parameter{
		Name:        "data-disks",
		Type:        "object-array",
		Description: "Data disks",
		ItemSchema: []SchemaField{
			{Name: "diskCategory", Type: "string"},
		},
	}
	cmd := &cobra.Command{Use: "test"}
	store := newFlagStore()
	bindFlag(cmd, &param, store)

	if cmd.Flags().Lookup("data-disks") == nil {
		t.Fatal("flag --data-disks not registered")
	}
}

func TestStringSliceValue_Set(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
		want   []string
	}{
		{
			name:   "single value",
			inputs: []string{"a"},
			want:   []string{"a"},
		},
		{
			name:   "multiple calls",
			inputs: []string{"a", "b", "c"},
			want:   []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v []string
			s := newStringSliceValue(&v)
			for _, input := range tt.inputs {
				if err := s.Set(input); err != nil {
					t.Fatalf("Set(%q) error = %v", input, err)
				}
			}
			if len(v) != len(tt.want) {
				t.Fatalf("got %v, want %v", v, tt.want)
			}
			for i, got := range v {
				if got != tt.want[i] {
					t.Errorf("v[%d] = %q, want %q", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestStringSliceValue_String(t *testing.T) {
	var v []string
	s := newStringSliceValue(&v)
	if s.String() != "" {
		t.Errorf("empty slice String() = %q, want empty", s.String())
	}
	v = []string{"a", "b", "c"}
	if s.String() != "a b c" {
		t.Errorf("String() = %q, want 'a b c'", s.String())
	}
}

func TestStringSliceValue_Type(t *testing.T) {
	var v []string
	s := newStringSliceValue(&v)
	if s.Type() != "strings" {
		t.Errorf("Type() = %q, want 'strings'", s.Type())
	}
}

func TestIntSliceValue_Set(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		want    []int
		wantErr bool
	}{
		{
			name:   "single value",
			inputs: []string{"1"},
			want:   []int{1},
		},
		{
			name:   "multiple calls",
			inputs: []string{"1", "2", "3"},
			want:   []int{1, 2, 3},
		},
		{
			name:    "invalid integer",
			inputs:  []string{"abc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v []int
			s := newIntSliceValue(&v)
			var err error
			for _, input := range tt.inputs {
				if err = s.Set(input); err != nil {
					break
				}
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
			if len(v) != len(tt.want) {
				t.Fatalf("got %v, want %v", v, tt.want)
			}
			for i, got := range v {
				if got != tt.want[i] {
					t.Errorf("v[%d] = %d, want %d", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestIntSliceValue_String(t *testing.T) {
	var v []int
	s := newIntSliceValue(&v)
	if s.String() != "" {
		t.Errorf("empty slice String() = %q, want empty", s.String())
	}
	v = []int{1, 2, 3}
	if s.String() != "1 2 3" {
		t.Errorf("String() = %q, want '1 2 3'", s.String())
	}
}

func TestIntSliceValue_Type(t *testing.T) {
	var v []int
	s := newIntSliceValue(&v)
	if s.Type() != "ints" {
		t.Errorf("Type() = %q, want 'ints'", s.Type())
	}
}

func TestExpandTrailingArgs(t *testing.T) {
	tests := []struct {
		name      string
		flagArgs  []string
		trailing  []string
		wantVals  []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "no trailing args",
			flagArgs: []string{"--ids", "a"},
			trailing: nil,
			wantVals: []string{"a"},
		},
		{
			name:     "with trailing args",
			flagArgs: []string{"--ids", "a"},
			trailing: []string{"b", "c"},
			wantVals: []string{"a", "b", "c"},
		},
		{
			name:      "trailing args without array flag",
			flagArgs:  nil,
			trailing:  []string{"a", "b"},
			wantErr:   true,
			errSubstr: "unexpected arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a command with a string-array flag.
			cmd := &cobra.Command{Use: "test"}
			var ids []string
			store := newFlagStore()
			store.stringArrays["ids"] = &ids
			store.arrayFlagNames = []string{"ids"}
			cmd.Flags().Var(newStringSliceValue(&ids), "ids", "IDs")

			// Parse the flags.
			if len(tt.flagArgs) > 0 {
				if err := cmd.Flags().Parse(tt.flagArgs); err != nil {
					t.Fatalf("Flags().Parse() error = %v", err)
				}
			}

			// Expand trailing args.
			err := expandTrailingArgs(cmd, tt.trailing, store)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("expandTrailingArgs() error = %v", err)
			}

			if len(ids) != len(tt.wantVals) {
				t.Fatalf("got %v, want %v", ids, tt.wantVals)
			}
			for i, got := range ids {
				if got != tt.wantVals[i] {
					t.Errorf("ids[%d] = %q, want %q", i, got, tt.wantVals[i])
				}
			}
		})
	}
}

func TestArrayFlagNamesTracked(t *testing.T) {
	params := []Parameter{
		{Name: "zone-id", Type: "string", Description: "Zone"},
		{Name: "instance-ids", Type: "string-array", Description: "Instance IDs"},
		{Name: "ports", Type: "integer-array", Description: "Ports"},
		{Name: "count", Type: "integer", Description: "Count"},
	}
	def := &APIDefinition{Parameters: params}
	cmd := &cobra.Command{Use: "test"}
	store := bindFlags(cmd, def)

	// Verify arrayFlagNames contains the array types.
	if len(store.arrayFlagNames) != 2 {
		t.Fatalf("len(arrayFlagNames) = %d, want 2", len(store.arrayFlagNames))
	}
	found := make(map[string]bool)
	for _, name := range store.arrayFlagNames {
		found[name] = true
	}
	if !found["instance-ids"] {
		t.Error("arrayFlagNames missing 'instance-ids'")
	}
	if !found["ports"] {
		t.Error("arrayFlagNames missing 'ports'")
	}
}
