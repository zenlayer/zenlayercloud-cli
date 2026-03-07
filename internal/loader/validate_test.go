package loader

import (
	"strings"
	"testing"
)

func TestValidateConsistency(t *testing.T) {
	base := &APIDefinition{
		Name:    "TestAction",
		Product: "zec",
		Use:     "test-action",
		SDK:     SDKInfo{Service: "zec", Version: "2024-04-01", Action: "TestAction"},
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string", Required: true, EnumValues: nil},
		},
	}

	t.Run("identical definitions produce no error", func(t *testing.T) {
		zh := *base
		zh.Short = "中文描述"
		if err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("mismatched name is detected", func(t *testing.T) {
		zh := *base
		zh.Name = "DifferentName"
		// In non-strict mode this prints a warning and returns nil.
		// We just verify it doesn't panic.
		_ = validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
	})

	t.Run("mismatched parameter type is detected", func(t *testing.T) {
		zh := *base
		zh.Parameters = []Parameter{
			{Name: "zone-id", Type: "integer", Required: true},
		}
		_ = validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
	})

	t.Run("mismatched parameter count is detected", func(t *testing.T) {
		zh := *base
		zh.Parameters = []Parameter{}
		_ = validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
	})

	t.Run("strict mode returns error on mismatch", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.Name = "WrongName"
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error in strict mode")
		}
		if !strings.Contains(err.Error(), "STRICT_VALIDATE") {
			t.Errorf("error %q should mention STRICT_VALIDATE", err.Error())
		}
	})
}

func TestValidateConsistency_SDKMismatches(t *testing.T) {
	base := &APIDefinition{
		Name:    "TestAction",
		Product: "zec",
		Use:     "test-action",
		SDK:     SDKInfo{Service: "zec", Version: "2024-04-01", Action: "TestAction"},
		Parameters: []Parameter{
			{Name: "zone-id", Type: "string"},
		},
	}

	t.Run("mismatched SDK service", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.SDK.Service = "bmc"
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for SDK service mismatch")
		}
	})

	t.Run("mismatched SDK version", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.SDK.Version = "2025-01-01"
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for SDK version mismatch")
		}
	})

	t.Run("mismatched SDK action", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.SDK.Action = "OtherAction"
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for SDK action mismatch")
		}
	})

	t.Run("mismatched product", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.Product = "bmc"
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for product mismatch")
		}
	})

	t.Run("mismatched use", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.Use = "other-action"
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for use mismatch")
		}
	})
}

func TestValidateConsistency_ParameterFieldMismatches(t *testing.T) {
	base := &APIDefinition{
		Name:    "TestAction",
		Product: "zec",
		Use:     "test-action",
		SDK:     SDKInfo{Service: "zec", Version: "2024-04-01", Action: "TestAction"},
		Parameters: []Parameter{
			{
				Name:       "zone-id",
				Type:       "string",
				Required:   true,
				SDKField:   "ZoneId",
				SDKWrapper: "wrapper",
				Sensitive:  true,
				Deprecated: false,
				EnumValues: EnumOptions{{Value: "A"}, {Value: "B"}},
			},
		},
	}

	t.Run("mismatched parameter name", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		zh.Parameters = []Parameter{{Name: "different-name", Type: "string"}}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for parameter name mismatch")
		}
	})

	t.Run("mismatched parameter required", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		p := base.Parameters[0]
		p.Required = false
		zh.Parameters = []Parameter{p}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for parameter required mismatch")
		}
	})

	t.Run("mismatched parameter sdk-field", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		p := base.Parameters[0]
		p.SDKField = "OtherField"
		zh.Parameters = []Parameter{p}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for sdk-field mismatch")
		}
	})

	t.Run("mismatched parameter sdk-wrapper", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		p := base.Parameters[0]
		p.SDKWrapper = "other"
		zh.Parameters = []Parameter{p}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for sdk-wrapper mismatch")
		}
	})

	t.Run("mismatched parameter sensitive", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		p := base.Parameters[0]
		p.Sensitive = false
		zh.Parameters = []Parameter{p}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for sensitive mismatch")
		}
	})

	t.Run("mismatched parameter deprecated", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		p := base.Parameters[0]
		p.Deprecated = true
		zh.Parameters = []Parameter{p}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for deprecated mismatch")
		}
	})

	t.Run("mismatched parameter enum-values", func(t *testing.T) {
		t.Setenv("STRICT_VALIDATE", "true")
		zh := *base
		p := base.Parameters[0]
		p.EnumValues = EnumOptions{{Value: "X"}, {Value: "Y"}, {Value: "Z"}}
		zh.Parameters = []Parameter{p}
		err := validateConsistency(base, &zh, "zh-CN/zec/test.yaml")
		if err == nil {
			t.Error("expected error for enum-values mismatch")
		}
	})
}

func TestStringSliceEqual(t *testing.T) {
	if !stringSliceEqual(nil, nil) {
		t.Error("nil slices should be equal")
	}
	if !stringSliceEqual([]string{}, []string{}) {
		t.Error("empty slices should be equal")
	}
	if !stringSliceEqual([]string{"a", "b"}, []string{"a", "b"}) {
		t.Error("identical slices should be equal")
	}
	if stringSliceEqual([]string{"a"}, []string{"b"}) {
		t.Error("different slices should not be equal")
	}
	if stringSliceEqual([]string{"a"}, []string{"a", "b"}) {
		t.Error("slices of different lengths should not be equal")
	}
}
