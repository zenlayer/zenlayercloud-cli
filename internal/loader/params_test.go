package loader

import (
	"strings"
	"testing"
)

func TestKebabToCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"instance-ids", "instanceIds"},
		{"zone-id", "zoneId"},
		{"page-size", "pageSize"},
		{"page-num", "pageNum"},
		{"nic-network-type", "nicNetworkType"},
		{"system-disk", "systemDisk"},
		{"name", "name"},
		{"", ""},
	}
	for _, tt := range tests {
		got := kebabToCamel(tt.input)
		if got != tt.want {
			t.Errorf("kebabToCamel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	enums := []string{"Auto", "VirtioOnly", "E1000Only"}

	if err := validateEnum("Auto", enums); err != nil {
		t.Errorf("validateEnum(Auto) = %v, want nil", err)
	}
	if err := validateEnum("invalid", enums); err == nil {
		t.Error("validateEnum(invalid) = nil, want error")
	}
}

func TestParseObjectValue(t *testing.T) {
	schema := []SchemaField{
		{Name: "diskCategory", Type: "string"},
		{Name: "diskSize", Type: "integer"},
	}

	t.Run("valid kv string", func(t *testing.T) {
		result, err := parseObjectValue("diskCategory=ssd,diskSize=50", schema)
		if err != nil {
			t.Fatalf("parseObjectValue() error = %v", err)
		}
		if result["diskCategory"] != "ssd" {
			t.Errorf("diskCategory = %v, want ssd", result["diskCategory"])
		}
		if result["diskSize"] != 50 {
			t.Errorf("diskSize = %v, want 50", result["diskSize"])
		}
	})

	t.Run("invalid kv pair missing equals", func(t *testing.T) {
		_, err := parseObjectValue("diskCategoryssd", schema)
		if err == nil {
			t.Error("expected error for missing '='")
		}
	})

	t.Run("empty string returns empty map", func(t *testing.T) {
		result, err := parseObjectValue("", schema)
		if err != nil {
			t.Fatalf("parseObjectValue('') error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("integer field type coercion error", func(t *testing.T) {
		_, err := parseObjectValue("diskSize=notanint", schema)
		if err == nil {
			t.Error("expected error for non-integer value in integer field")
		}
	})
}

func TestCollectParams(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string"},
			{Name: "page-size", Type: "integer"},
			{Name: "enabled", Type: "boolean"},
		},
	}
	store := newFlagStore()
	zoneID := "asia-east-1a"
	store.strings["zone-id"] = &zoneID
	pageSize := 10
	store.ints["page-size"] = &pageSize
	enabled := true
	store.bools["enabled"] = &enabled

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if params["zoneId"] != "asia-east-1a" {
		t.Errorf("zoneId = %v, want 'asia-east-1a'", params["zoneId"])
	}
	if params["pageSize"] != 10 {
		t.Errorf("pageSize = %v, want 10", params["pageSize"])
	}
	if params["enabled"] != true {
		t.Errorf("enabled = %v, want true", params["enabled"])
	}
}

func TestCollectParams_ZeroValuesOmitted(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string"},
			{Name: "page-size", Type: "integer"},
		},
	}
	store := newFlagStore()
	empty := ""
	store.strings["zone-id"] = &empty
	zero := 0
	store.ints["page-size"] = &zero

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestCollectParams_EnumValidation(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "nic-type", Type: "enum", EnumValues: EnumOptions{{Value: "Auto"}, {Value: "Manual"}}},
		},
	}
	store := newFlagStore()

	t.Run("valid enum value", func(t *testing.T) {
		v := "Auto"
		store.strings["nic-type"] = &v
		_, err := collectParams(store, def, nil)
		if err != nil {
			t.Errorf("collectParams() unexpected error: %v", err)
		}
	})

	t.Run("invalid enum value", func(t *testing.T) {
		v := "Unknown"
		store.strings["nic-type"] = &v
		_, err := collectParams(store, def, nil)
		if err == nil {
			t.Error("collectParams() expected error for invalid enum")
		}
	})
}

func TestCollectParams_ObjectArray(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name:     "data-disk",
				Type:     "object-array",
				ItemSchema: []SchemaField{
					{Name: "diskCategory", Type: "string"},
					{Name: "diskSize", Type: "integer"},
				},
			},
		},
	}
	store := newFlagStore()
	disks := []string{"diskCategory=ssd,diskSize=100"}
	store.stringArrays["data-disk"] = &disks

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	arr, ok := params["dataDisk"].([]map[string]interface{})
	if !ok {
		t.Fatalf("dataDisk type = %T, want []map[string]interface{}", params["dataDisk"])
	}
	if len(arr) != 1 {
		t.Fatalf("dataDisk len = %d, want 1", len(arr))
	}
	if arr[0]["diskCategory"] != "ssd" {
		t.Errorf("diskCategory = %v, want 'ssd'", arr[0]["diskCategory"])
	}
}

func TestCollectParams_SDKWrapper(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name:       "tag",
				Type:       "object-array",
				SDKWrapper: "tags",
				ItemSchema: []SchemaField{
					{Name: "key", Type: "string"},
					{Name: "value", Type: "string"},
				},
			},
		},
	}
	store := newFlagStore()
	tags := []string{"key=env,value=prod"}
	store.stringArrays["tag"] = &tags

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	// sdk-wrapper wraps the result
	wrapper, ok := params["tag"].(map[string]interface{})
	if !ok {
		t.Fatalf("tag type = %T, want map[string]interface{}", params["tag"])
	}
	if wrapper["tags"] == nil {
		t.Error("expected wrapper to contain 'tags' key")
	}
}

func TestCollectParams_StringArray(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "instance-ids", Type: "string-array"},
		},
	}
	store := newFlagStore()
	ids := []string{"ins-001", "ins-002"}
	store.stringArrays["instance-ids"] = &ids

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	got, ok := params["instanceIds"].([]string)
	if !ok {
		t.Fatalf("instanceIds type = %T, want []string", params["instanceIds"])
	}
	if len(got) != 2 || got[0] != "ins-001" {
		t.Errorf("instanceIds = %v, want [ins-001 ins-002]", got)
	}
}

func TestCollectParams_StringArray_Empty(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "instance-ids", Type: "string-array"},
		},
	}
	store := newFlagStore()
	empty := []string{}
	store.stringArrays["instance-ids"] = &empty

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if _, ok := params["instanceIds"]; ok {
		t.Error("empty string-array should be omitted from params")
	}
}

func TestCollectParams_IntegerArray(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "port-ids", Type: "integer-array"},
		},
	}
	store := newFlagStore()
	ports := []int{80, 443}
	store.intArrays["port-ids"] = &ports

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	got, ok := params["portIds"].([]int)
	if !ok {
		t.Fatalf("portIds type = %T, want []int", params["portIds"])
	}
	if len(got) != 2 || got[0] != 80 {
		t.Errorf("portIds = %v, want [80 443]", got)
	}
}

func TestCollectParams_IntegerArray_Empty(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "port-ids", Type: "integer-array"},
		},
	}
	store := newFlagStore()
	empty := []int{}
	store.intArrays["port-ids"] = &empty

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if _, ok := params["portIds"]; ok {
		t.Error("empty integer-array should be omitted from params")
	}
}

func TestCollectParams_Float(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "cpu-ratio", Type: "float"},
		},
	}
	store := newFlagStore()
	ratio := 1.5
	store.floats["cpu-ratio"] = &ratio

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if params["cpuRatio"] != 1.5 {
		t.Errorf("cpuRatio = %v, want 1.5", params["cpuRatio"])
	}
}

func TestCollectParams_Float_Zero(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "cpu-ratio", Type: "float"},
		},
	}
	store := newFlagStore()
	zero := 0.0
	store.floats["cpu-ratio"] = &zero

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if _, ok := params["cpuRatio"]; ok {
		t.Error("zero float should be omitted from params")
	}
}

func TestCollectParams_Object(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name: "system-disk",
				Type: "object",
				ObjectSchema: []SchemaField{
					{Name: "diskCategory", Type: "string"},
					{Name: "diskSize", Type: "integer"},
				},
			},
		},
	}
	store := newFlagStore()
	v := "diskCategory=ssd,diskSize=100"
	store.strings["system-disk"] = &v

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	obj, ok := params["systemDisk"].(map[string]interface{})
	if !ok {
		t.Fatalf("systemDisk type = %T, want map[string]interface{}", params["systemDisk"])
	}
	if obj["diskCategory"] != "ssd" {
		t.Errorf("diskCategory = %v, want 'ssd'", obj["diskCategory"])
	}
	if obj["diskSize"] != 100 {
		t.Errorf("diskSize = %v, want 100", obj["diskSize"])
	}
}

func TestCollectParams_Object_WithSDKWrapper(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name:       "system-disk",
				Type:       "object",
				SDKWrapper: "disk",
				ObjectSchema: []SchemaField{
					{Name: "diskCategory", Type: "string"},
				},
			},
		},
	}
	store := newFlagStore()
	v := "diskCategory=ssd"
	store.strings["system-disk"] = &v

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	wrapper, ok := params["systemDisk"].(map[string]interface{})
	if !ok {
		t.Fatalf("systemDisk type = %T, want map[string]interface{}", params["systemDisk"])
	}
	if wrapper["disk"] == nil {
		t.Error("expected wrapper to contain 'disk' key")
	}
}

func TestCollectParams_Object_ParseError(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "system-disk", Type: "object", ObjectSchema: []SchemaField{{Name: "diskSize", Type: "integer"}}},
		},
	}
	store := newFlagStore()
	v := "diskSize=notanint"
	store.strings["system-disk"] = &v

	_, err := collectParams(store, def, nil)
	if err == nil {
		t.Error("expected error for invalid object value")
	}
}

func TestCollectParams_ObjectArray_ParseError(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name: "data-disk",
				Type: "object-array",
				ItemSchema: []SchemaField{
					{Name: "diskSize", Type: "integer"},
				},
			},
		},
	}
	store := newFlagStore()
	disks := []string{"diskSize=notanint"}
	store.stringArrays["data-disk"] = &disks

	_, err := collectParams(store, def, nil)
	if err == nil {
		t.Error("expected error for invalid object-array value")
	}
}

func TestCollectParams_Sensitive(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "password", Type: "string", Sensitive: true},
		},
	}
	store := newFlagStore()
	sensitiveValues := map[string]string{"password": "s3cr3t"}

	params, err := collectParams(store, def, sensitiveValues)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if params["password"] != "s3cr3t" {
		t.Errorf("password = %v, want 's3cr3t'", params["password"])
	}
}

func TestCollectParams_Sensitive_Missing(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "password", Type: "string", Sensitive: true},
		},
	}
	store := newFlagStore()

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	// Sensitive value not provided → should be omitted.
	if _, ok := params["password"]; ok {
		t.Error("missing sensitive value should be omitted from params")
	}
}

func TestCollectParams_SDKField(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "instance-id", Type: "string", SDKField: "InstanceId"},
		},
	}
	store := newFlagStore()
	v := "ins-abc"
	store.strings["instance-id"] = &v

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	if params["InstanceId"] != "ins-abc" {
		t.Errorf("InstanceId = %v, want 'ins-abc'", params["InstanceId"])
	}
	if _, ok := params["instanceId"]; ok {
		t.Error("sdk-field should override camelCase key")
	}
}

func TestSdkKey(t *testing.T) {
	tests := []struct {
		param Parameter
		want  string
	}{
		{Parameter{Name: "zone-id"}, "zoneId"},
		{Parameter{Name: "zone-id", SDKField: "ZoneId"}, "ZoneId"},
		{Parameter{Name: "name"}, "name"},
		{Parameter{Name: "name", SDKField: "Name"}, "Name"},
	}
	for _, tt := range tests {
		got := sdkKey(&tt.param)
		if got != tt.want {
			t.Errorf("sdkKey(%+v) = %q, want %q", tt.param, got, tt.want)
		}
	}
}

func TestParseObjectValue_FloatCoercion(t *testing.T) {
	schema := []SchemaField{
		{Name: "ratio", Type: "float"},
	}
	result, err := parseObjectValue("ratio=3.14", schema)
	if err != nil {
		t.Fatalf("parseObjectValue() error = %v", err)
	}
	if result["ratio"] != 3.14 {
		t.Errorf("ratio = %v, want 3.14", result["ratio"])
	}
}

func TestParseObjectValue_FloatCoercionError(t *testing.T) {
	schema := []SchemaField{
		{Name: "ratio", Type: "float"},
	}
	_, err := parseObjectValue("ratio=notafloat", schema)
	if err == nil {
		t.Error("expected error for invalid float value")
	}
}

func TestParseObjectValue_BooleanCoercion(t *testing.T) {
	schema := []SchemaField{
		{Name: "enabled", Type: "boolean"},
	}
	tests := []struct {
		raw  string
		want bool
	}{
		{"enabled=true", true},
		{"enabled=TRUE", true},
		{"enabled=True", true},
		{"enabled=1", true},
		{"enabled=false", false},
		{"enabled=0", false},
	}
	for _, tt := range tests {
		result, err := parseObjectValue(tt.raw, schema)
		if err != nil {
			t.Fatalf("parseObjectValue(%q) error = %v", tt.raw, err)
		}
		if result["enabled"] != tt.want {
			t.Errorf("parseObjectValue(%q) enabled = %v, want %v", tt.raw, result["enabled"], tt.want)
		}
	}
}

func TestParseObjectValue_EmptyKey(t *testing.T) {
	_, err := parseObjectValue("=value", nil)
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestValidateRequired(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string", Required: true},
			{Name: "optional", Type: "string", Required: false},
		},
	}

	t.Run("missing required field returns error", func(t *testing.T) {
		store := newFlagStore()
		empty := ""
		store.strings["zone-id"] = &empty
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required field")
		}
	})

	t.Run("all required fields present", func(t *testing.T) {
		store := newFlagStore()
		v := "asia-east-1a"
		store.strings["zone-id"] = &v
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_StringArray(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "instance-ids", Type: "string-array", Required: true},
		},
	}

	t.Run("missing", func(t *testing.T) {
		store := newFlagStore()
		empty := []string{}
		store.stringArrays["instance-ids"] = &empty
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required string-array")
		}
	})

	t.Run("provided", func(t *testing.T) {
		store := newFlagStore()
		ids := []string{"ins-001"}
		store.stringArrays["instance-ids"] = &ids
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_Integer(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "page-size", Type: "integer", Required: true},
		},
	}

	t.Run("missing (zero)", func(t *testing.T) {
		store := newFlagStore()
		zero := 0
		store.ints["page-size"] = &zero
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required integer (zero)")
		}
	})

	t.Run("provided", func(t *testing.T) {
		store := newFlagStore()
		v := 10
		store.ints["page-size"] = &v
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_IntegerArray(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "port-ids", Type: "integer-array", Required: true},
		},
	}

	t.Run("missing", func(t *testing.T) {
		store := newFlagStore()
		empty := []int{}
		store.intArrays["port-ids"] = &empty
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required integer-array")
		}
	})

	t.Run("provided", func(t *testing.T) {
		store := newFlagStore()
		ports := []int{80}
		store.intArrays["port-ids"] = &ports
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_Float(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "ratio", Type: "float", Required: true},
		},
	}

	t.Run("missing (zero)", func(t *testing.T) {
		store := newFlagStore()
		zero := 0.0
		store.floats["ratio"] = &zero
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required float (zero)")
		}
	})

	t.Run("provided", func(t *testing.T) {
		store := newFlagStore()
		v := 1.5
		store.floats["ratio"] = &v
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_Boolean(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "enabled", Type: "boolean", Required: true},
		},
	}
	// Boolean is always considered "provided" regardless of its value.
	store := newFlagStore()
	v := false
	store.bools["enabled"] = &v
	if err := validateRequired(store, def, nil); err != nil {
		t.Errorf("boolean required should always pass, got error: %v", err)
	}
}

func TestValidateRequired_Object(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "system-disk", Type: "object", Required: true},
		},
	}

	t.Run("missing", func(t *testing.T) {
		store := newFlagStore()
		empty := ""
		store.strings["system-disk"] = &empty
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required object")
		}
	})

	t.Run("provided", func(t *testing.T) {
		store := newFlagStore()
		v := "diskCategory=ssd"
		store.strings["system-disk"] = &v
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_ObjectArray(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "data-disk", Type: "object-array", Required: true},
		},
	}

	t.Run("missing", func(t *testing.T) {
		store := newFlagStore()
		empty := []string{}
		store.stringArrays["data-disk"] = &empty
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required object-array")
		}
	})

	t.Run("provided", func(t *testing.T) {
		store := newFlagStore()
		disks := []string{"diskCategory=ssd"}
		store.stringArrays["data-disk"] = &disks
		if err := validateRequired(store, def, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_SensitiveString(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "password", Type: "string", Required: true, Sensitive: true},
		},
	}

	t.Run("missing sensitive", func(t *testing.T) {
		store := newFlagStore()
		if err := validateRequired(store, def, nil); err == nil {
			t.Error("expected error for missing required sensitive string")
		}
	})

	t.Run("provided via sensitiveValues", func(t *testing.T) {
		store := newFlagStore()
		sensitiveValues := map[string]string{"password": "s3cr3t"}
		if err := validateRequired(store, def, sensitiveValues); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateRequired_MultipleErrors(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string", Required: true},
			{Name: "instance-name", Type: "string", Required: true},
		},
	}
	store := newFlagStore()
	e1 := ""
	store.strings["zone-id"] = &e1
	e2 := ""
	store.strings["instance-name"] = &e2

	err := validateRequired(store, def, nil)
	if err == nil {
		t.Fatal("expected error for multiple missing required fields")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "--zone-id") || !strings.Contains(errMsg, "--instance-name") {
		t.Errorf("error %q should list all missing fields", errMsg)
	}
}

// JSON format parsing tests

func TestParseObjectValue_JSON(t *testing.T) {
	schema := []SchemaField{
		{Name: "diskCategory", Type: "string"},
		{Name: "diskSize", Type: "integer"},
	}

	t.Run("valid JSON object", func(t *testing.T) {
		result, err := parseObjectValue(`{"diskCategory":"ssd","diskSize":100}`, schema)
		if err != nil {
			t.Fatalf("parseObjectValue() error = %v", err)
		}
		if result["diskCategory"] != "ssd" {
			t.Errorf("diskCategory = %v, want 'ssd'", result["diskCategory"])
		}
		// JSON numbers are float64
		if result["diskSize"] != float64(100) {
			t.Errorf("diskSize = %v, want 100", result["diskSize"])
		}
	})

	t.Run("JSON with whitespace", func(t *testing.T) {
		result, err := parseObjectValue(`  { "diskCategory" : "nvme" }  `, schema)
		if err != nil {
			t.Fatalf("parseObjectValue() error = %v", err)
		}
		if result["diskCategory"] != "nvme" {
			t.Errorf("diskCategory = %v, want 'nvme'", result["diskCategory"])
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := parseObjectValue(`{"diskCategory":"ssd"`, schema)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestParseObjectArrayValue_JSONArray(t *testing.T) {
	schema := []SchemaField{
		{Name: "key", Type: "string"},
		{Name: "value", Type: "string"},
	}

	t.Run("valid JSON array", func(t *testing.T) {
		rawItems := []string{`[{"key":"env","value":"prod"},{"key":"team","value":"backend"}]`}
		result, err := parseObjectArrayValue(rawItems, schema)
		if err != nil {
			t.Fatalf("parseObjectArrayValue() error = %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("len(result) = %d, want 2", len(result))
		}
		if result[0]["key"] != "env" {
			t.Errorf("result[0][key] = %v, want 'env'", result[0]["key"])
		}
		if result[1]["key"] != "team" {
			t.Errorf("result[1][key] = %v, want 'team'", result[1]["key"])
		}
	})

	t.Run("JSON array with whitespace", func(t *testing.T) {
		rawItems := []string{`  [ { "key" : "env" , "value" : "dev" } ]  `}
		result, err := parseObjectArrayValue(rawItems, schema)
		if err != nil {
			t.Fatalf("parseObjectArrayValue() error = %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("len(result) = %d, want 1", len(result))
		}
		if result[0]["key"] != "env" {
			t.Errorf("result[0][key] = %v, want 'env'", result[0]["key"])
		}
	})

	t.Run("invalid JSON array", func(t *testing.T) {
		rawItems := []string{`[{"key":"env"`}
		_, err := parseObjectArrayValue(rawItems, schema)
		if err == nil {
			t.Error("expected error for invalid JSON array")
		}
	})

	t.Run("KV format still works", func(t *testing.T) {
		rawItems := []string{"key=env,value=prod", "key=team,value=backend"}
		result, err := parseObjectArrayValue(rawItems, schema)
		if err != nil {
			t.Fatalf("parseObjectArrayValue() error = %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("len(result) = %d, want 2", len(result))
		}
		if result[0]["key"] != "env" {
			t.Errorf("result[0][key] = %v, want 'env'", result[0]["key"])
		}
	})

	t.Run("single JSON object element", func(t *testing.T) {
		rawItems := []string{`{"key":"name","value":"test"}`}
		result, err := parseObjectArrayValue(rawItems, schema)
		if err != nil {
			t.Fatalf("parseObjectArrayValue() error = %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("len(result) = %d, want 1", len(result))
		}
		if result[0]["key"] != "name" {
			t.Errorf("result[0][key] = %v, want 'name'", result[0]["key"])
		}
	})
}

func TestCollectParams_ObjectJSON(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name: "system-disk",
				Type: "object",
				ObjectSchema: []SchemaField{
					{Name: "diskCategory", Type: "string"},
					{Name: "diskSize", Type: "integer"},
				},
			},
		},
	}
	store := newFlagStore()
	v := `{"diskCategory":"ssd","diskSize":100}`
	store.strings["system-disk"] = &v

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	obj, ok := params["systemDisk"].(map[string]interface{})
	if !ok {
		t.Fatalf("systemDisk type = %T, want map[string]interface{}", params["systemDisk"])
	}
	if obj["diskCategory"] != "ssd" {
		t.Errorf("diskCategory = %v, want 'ssd'", obj["diskCategory"])
	}
}

func TestCollectParams_ObjectArrayJSON(t *testing.T) {
	def := &APIDefinition{
		Parameters: []Parameter{
			{
				Name: "tags",
				Type: "object-array",
				ItemSchema: []SchemaField{
					{Name: "key", Type: "string"},
					{Name: "value", Type: "string"},
				},
			},
		},
	}
	store := newFlagStore()
	tags := []string{`[{"key":"env","value":"prod"},{"key":"team","value":"backend"}]`}
	store.stringArrays["tags"] = &tags

	params, err := collectParams(store, def, nil)
	if err != nil {
		t.Fatalf("collectParams() error = %v", err)
	}
	arr, ok := params["tags"].([]map[string]interface{})
	if !ok {
		t.Fatalf("tags type = %T, want []map[string]interface{}", params["tags"])
	}
	if len(arr) != 2 {
		t.Fatalf("tags len = %d, want 2", len(arr))
	}
	if arr[0]["key"] != "env" {
		t.Errorf("tags[0][key] = %v, want 'env'", arr[0]["key"])
	}
}
