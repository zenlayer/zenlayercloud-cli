package loader

import (
	"fmt"
	"os"
	"strings"
)

// validateConsistency checks that structural fields in zh-CN are identical to
// en-US (the source of truth). Text-only fields (short/long/examples/description)
// are allowed to differ.
//
// When STRICT_VALIDATE=true the function returns a fatal error; otherwise it
// prints warnings and returns nil so the CLI continues to function.
func validateConsistency(enDef, zhDef *APIDefinition, path string) error {
	strict := os.Getenv("STRICT_VALIDATE") == "true"

	var issues []string

	if enDef.Name != zhDef.Name {
		issues = append(issues, fmt.Sprintf("name: en-US=%q, zh-CN=%q", enDef.Name, zhDef.Name))
	}
	if enDef.Product != zhDef.Product {
		issues = append(issues, fmt.Sprintf("product: en-US=%q, zh-CN=%q", enDef.Product, zhDef.Product))
	}
	if enDef.Use != zhDef.Use {
		issues = append(issues, fmt.Sprintf("use: en-US=%q, zh-CN=%q", enDef.Use, zhDef.Use))
	}
	if enDef.SDK.Service != zhDef.SDK.Service {
		issues = append(issues, fmt.Sprintf("sdk.service: en-US=%q, zh-CN=%q", enDef.SDK.Service, zhDef.SDK.Service))
	}
	if enDef.SDK.Version != zhDef.SDK.Version {
		issues = append(issues, fmt.Sprintf("sdk.version: en-US=%q, zh-CN=%q", enDef.SDK.Version, zhDef.SDK.Version))
	}
	if enDef.SDK.Action != zhDef.SDK.Action {
		issues = append(issues, fmt.Sprintf("sdk.action: en-US=%q, zh-CN=%q", enDef.SDK.Action, zhDef.SDK.Action))
	}

	if len(enDef.Parameters) != len(zhDef.Parameters) {
		issues = append(issues, fmt.Sprintf("parameter count: en-US=%d, zh-CN=%d", len(enDef.Parameters), len(zhDef.Parameters)))
	} else {
		for i := range enDef.Parameters {
			ep := enDef.Parameters[i]
			zp := zhDef.Parameters[i]
			prefix := fmt.Sprintf("parameter[%d]", i)

			if ep.Name != zp.Name {
				issues = append(issues, fmt.Sprintf("%s.name: en-US=%q, zh-CN=%q", prefix, ep.Name, zp.Name))
			}
			if ep.Type != zp.Type {
				issues = append(issues, fmt.Sprintf("%s.type: en-US=%q, zh-CN=%q", prefix, ep.Type, zp.Type))
			}
			if ep.Required != zp.Required {
				issues = append(issues, fmt.Sprintf("%s.required: en-US=%v, zh-CN=%v", prefix, ep.Required, zp.Required))
			}
			if ep.SDKField != zp.SDKField {
				issues = append(issues, fmt.Sprintf("%s.sdk-field: en-US=%q, zh-CN=%q", prefix, ep.SDKField, zp.SDKField))
			}
			if ep.SDKWrapper != zp.SDKWrapper {
				issues = append(issues, fmt.Sprintf("%s.sdk-wrapper: en-US=%q, zh-CN=%q", prefix, ep.SDKWrapper, zp.SDKWrapper))
			}
			if ep.Sensitive != zp.Sensitive {
				issues = append(issues, fmt.Sprintf("%s.sensitive: en-US=%v, zh-CN=%v", prefix, ep.Sensitive, zp.Sensitive))
			}
			if ep.Deprecated != zp.Deprecated {
				issues = append(issues, fmt.Sprintf("%s.deprecated: en-US=%v, zh-CN=%v", prefix, ep.Deprecated, zp.Deprecated))
			}
			if !ep.EnumValues.Equal(zp.EnumValues) {
				issues = append(issues, fmt.Sprintf("%s.enum-values: en-US=%v, zh-CN=%v", prefix, ep.EnumValues.Values(), zp.EnumValues.Values()))
			}
		}
	}

	if len(issues) == 0 {
		return nil
	}

	msg := fmt.Sprintf("[WARNING] %s structure mismatch with en-US:\n  %s\n",
		path, strings.Join(issues, "\n  "))

	if strict {
		return fmt.Errorf("STRICT_VALIDATE: %s", msg)
	}
	fmt.Fprint(os.Stderr, msg)
	return nil
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
