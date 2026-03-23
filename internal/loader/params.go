package loader

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// collectParams gathers flag values from the store into a map ready to pass to
// the API client. Zero-valued flags are omitted to avoid polluting the request body.
// Sensitive string flags must be read separately (via readSensitive) before this call.
func collectParams(store *flagStore, def *APIDefinition, sensitiveValues map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for i := range def.Parameters {
		param := &def.Parameters[i]
		name := param.Name
		key := sdkKey(param)

		switch param.Type {
		case "string", "enum":
			if param.Sensitive {
				if v, ok := sensitiveValues[name]; ok && v != "" {
					result[key] = v
				}
				continue
			}
			if v, ok := store.strings[name]; ok && *v != "" {
				if param.Type == "enum" {
					if err := validateEnum(*v, param.EnumValues.Values()); err != nil {
						return nil, fmt.Errorf("flag --%s: %w", name, err)
					}
				}
				result[key] = *v
			}

		case "string-array":
			if v, ok := store.stringArrays[name]; ok && len(*v) > 0 {
				result[key] = *v
			}

		case "integer":
			if v, ok := store.ints[name]; ok && *v != 0 {
				result[key] = *v
			}

		case "integer-array":
			if v, ok := store.intArrays[name]; ok && len(*v) > 0 {
				result[key] = *v
			}

		case "float":
			if v, ok := store.floats[name]; ok && *v != 0 {
				result[key] = *v
			}

		case "boolean":
			if v, ok := store.bools[name]; ok && *v {
				result[key] = *v
			}

		case "object":
			if v, ok := store.strings[name]; ok && *v != "" {
				parsed, err := parseObjectValue(*v, param.ObjectSchema)
				if err != nil {
					return nil, fmt.Errorf("flag --%s: %w", name, err)
				}
				if param.SDKWrapper != "" {
					result[key] = map[string]interface{}{param.SDKWrapper: parsed}
				} else {
					result[key] = parsed
				}
			}

		case "object-array":
			if v, ok := store.stringArrays[name]; ok && len(*v) > 0 {
				items, err := parseObjectArrayValue(*v, param.ItemSchema)
				if err != nil {
					return nil, fmt.Errorf("flag --%s: %w", name, err)
				}
				if param.SDKWrapper != "" {
					result[key] = map[string]interface{}{param.SDKWrapper: items}
				} else {
					result[key] = items
				}
			}
		}
	}

	return result, nil
}

// validateRequired checks that all required parameters have been provided.
func validateRequired(store *flagStore, def *APIDefinition, sensitiveValues map[string]string) error {
	var missing []string
	for i := range def.Parameters {
		param := &def.Parameters[i]
		if !param.Required {
			continue
		}
		name := param.Name
		provided := false

		switch param.Type {
		case "string", "enum":
			if param.Sensitive {
				if v, ok := sensitiveValues[name]; ok && v != "" {
					provided = true
				}
			} else if v, ok := store.strings[name]; ok && *v != "" {
				provided = true
			}
		case "string-array":
			if v, ok := store.stringArrays[name]; ok && len(*v) > 0 {
				provided = true
			}
		case "integer":
			if v, ok := store.ints[name]; ok && *v != 0 {
				provided = true
			}
		case "integer-array":
			if v, ok := store.intArrays[name]; ok && len(*v) > 0 {
				provided = true
			}
		case "float":
			if v, ok := store.floats[name]; ok && *v != 0 {
				provided = true
			}
		case "boolean":
			provided = true // booleans are always "provided" (false is a valid value)
		case "object":
			if v, ok := store.strings[name]; ok && *v != "" {
				provided = true
			}
		case "object-array":
			if v, ok := store.stringArrays[name]; ok && len(*v) > 0 {
				provided = true
			}
		}

		if !provided {
			missing = append(missing, "--"+name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required flags not set: %s", strings.Join(missing, ", "))
	}
	return nil
}

// sdkKey returns the JSON key to use in the request body for a parameter.
// If sdk-field is set, use that; otherwise convert kebab-case to camelCase.
func sdkKey(param *Parameter) string {
	if param.SDKField != "" {
		return param.SDKField
	}
	return kebabToCamel(param.Name)
}

// kebabToCamel converts "instance-ids" → "instanceIds".
func kebabToCamel(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	for i, p := range parts {
		if i == 0 {
			b.WriteString(p)
		} else if len(p) > 0 {
			runes := []rune(p)
			runes[0] = unicode.ToUpper(runes[0])
			b.WriteString(string(runes))
		}
	}
	return b.String()
}

// parseObjectValue parses a kv string like "category=ssd,size=50" into a map.
// If the input starts with '{' it is treated as JSON.
func parseObjectValue(raw string, schema []SchemaField) (map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)

	// JSON format detection
	if strings.HasPrefix(raw, "{") {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return result, nil
	}

	// KV format parsing
	result := make(map[string]interface{})

	// Build type lookup from schema
	typeFor := make(map[string]string, len(schema))
	for _, f := range schema {
		typeFor[f.Name] = f.Type
	}

	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		idx := strings.IndexByte(pair, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid kv pair %q (expected key=value)", pair)
		}
		k := strings.TrimSpace(pair[:idx])
		v := strings.TrimSpace(pair[idx+1:])
		if k == "" {
			return nil, fmt.Errorf("empty key in kv pair %q", pair)
		}

		// Type coercion based on schema
		switch typeFor[k] {
		case "integer":
			var n int
			if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
				return nil, fmt.Errorf("field %q expects integer, got %q", k, v)
			}
			result[k] = n
		case "float":
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
				return nil, fmt.Errorf("field %q expects float, got %q", k, v)
			}
			result[k] = f
		case "boolean":
			result[k] = strings.EqualFold(v, "true") || v == "1"
		default:
			result[k] = v
		}
	}
	return result, nil
}

// parseObjectArrayValue parses object-array values supporting both JSON array and KV formats.
// If the input is a single element starting with '[', it's parsed as a JSON array.
// Otherwise, each element is parsed as a KV string.
func parseObjectArrayValue(rawItems []string, schema []SchemaField) ([]map[string]interface{}, error) {
	// Check for JSON array format: single element starting with '['
	if len(rawItems) == 1 {
		trimmed := strings.TrimSpace(rawItems[0])
		if strings.HasPrefix(trimmed, "[") {
			var result []map[string]interface{}
			if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
				return nil, fmt.Errorf("invalid JSON array: %w", err)
			}
			return result, nil
		}
	}

	// Parse each element as KV or JSON object
	var items []map[string]interface{}
	for _, raw := range rawItems {
		parsed, err := parseObjectValue(raw, schema)
		if err != nil {
			return nil, err
		}
		items = append(items, parsed)
	}
	return items, nil
}

// validateEnum checks that value is one of the allowed enumValues.
func validateEnum(value string, enumValues []string) error {
	for _, e := range enumValues {
		if e == value {
			return nil
		}
	}
	return fmt.Errorf("invalid value %q, allowed values: %s", value, strings.Join(enumValues, ", "))
}
