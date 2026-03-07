package output

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

// TableFormatter formats output as a table.
type TableFormatter struct{}

// Format implements Formatter interface for table output.
func (f *TableFormatter) Format(w io.Writer, data interface{}) error {
	switch v := data.(type) {
	case map[string]string:
		return f.formatKeyValue(w, v)
	case map[string]interface{}:
		return f.formatMapInterface(w, v)
	default:
		return f.formatStruct(w, data)
	}
}

// formatKeyValue formats a simple key-value map.
func (f *TableFormatter) formatKeyValue(w io.Writer, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	// Find max key length for alignment
	maxKeyLen := 0
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	sort.Strings(keys)

	// Print header
	fmt.Fprintf(w, "%-*s  %s\n", maxKeyLen, "Key", "Value")
	fmt.Fprintf(w, "%-*s  %s\n", maxKeyLen, strings.Repeat("-", maxKeyLen), "-----")

	// Print rows
	for _, k := range keys {
		fmt.Fprintf(w, "%-*s  %s\n", maxKeyLen, k, data[k])
	}

	return nil
}

// formatMapInterface formats a map[string]interface{}.
func (f *TableFormatter) formatMapInterface(w io.Writer, data map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Convert to string map for simple cases
	strMap := make(map[string]string)
	for k, v := range data {
		strMap[k] = fmt.Sprintf("%v", v)
	}
	return f.formatKeyValue(w, strMap)
}

// formatStruct formats a struct as a table.
func (f *TableFormatter) formatStruct(w io.Writer, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// Fallback to simple string representation
		fmt.Fprintln(w, data)
		return nil
	}

	t := v.Type()
	maxKeyLen := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath == "" { // exported field
			name := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					name = parts[0]
				}
			}
			if len(name) > maxKeyLen {
				maxKeyLen = len(name)
			}
		}
	}

	// Print header
	fmt.Fprintf(w, "%-*s  %s\n", maxKeyLen, "Key", "Value")
	fmt.Fprintf(w, "%-*s  %s\n", maxKeyLen, strings.Repeat("-", maxKeyLen), "-----")

	// Print fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath == "" { // exported field
			name := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					name = parts[0]
				}
			}
			value := v.Field(i).Interface()
			fmt.Fprintf(w, "%-*s  %v\n", maxKeyLen, name, value)
		}
	}

	return nil
}
