package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTableFormatter_ImplementsFormatter(t *testing.T) {
	var _ Formatter = &TableFormatter{}
}

func TestTableFormatter_Format_MapStringString(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		wantNoop bool
		contains []string
	}{
		{
			name:     "empty map returns no output",
			data:     map[string]string{},
			wantNoop: true,
		},
		{
			name:     "single key-value pair",
			data:     map[string]string{"key": "value"},
			contains: []string{"key", "value"},
		},
		{
			name:     "multiple key-value pairs rendered sorted",
			data:     map[string]string{"name": "Alice", "age": "30"},
			contains: []string{"age", "name", "30", "Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TableFormatter{}
			var buf bytes.Buffer
			if err := f.Format(&buf, tt.data); err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if tt.wantNoop {
				if buf.Len() != 0 {
					t.Errorf("Format() wrote %q, want no output", buf.String())
				}
				return
			}
			got := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Format() output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestTableFormatter_Format_MapStringInterface(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		wantNoop bool
		contains []string
	}{
		{
			name:     "empty map returns no output",
			data:     map[string]interface{}{},
			wantNoop: true,
		},
		{
			name:     "single string value",
			data:     map[string]interface{}{"key": "value"},
			contains: []string{"key", "value"},
		},
		{
			name:     "integer value",
			data:     map[string]interface{}{"count": 42},
			contains: []string{"count", "42"},
		},
		{
			name: "nested object renders as section",
			data: map[string]interface{}{
				"name": "Alice",
				"disk": map[string]interface{}{"size": 50, "category": "ssd"},
			},
			contains: []string{"name", "Alice", "disk", "size", "ssd"},
		},
		{
			name: "array of objects renders as inline",
			data: map[string]interface{}{
				"instances": []interface{}{
					map[string]interface{}{"id": "ins-1", "status": "RUNNING"},
					map[string]interface{}{"id": "ins-2", "status": "STOPPED"},
				},
			},
			// Small object arrays are now rendered inline: "ins-1(RUNNING), ins-2(STOPPED)".
			// Field names are not repeated as column headers.
			contains: []string{"instances", "ins-1", "RUNNING", "ins-2", "STOPPED"},
		},
		{
			name: "scalar array renders as list",
			data: map[string]interface{}{
				"tags": []interface{}{"env=prod", "region=asia"},
			},
			contains: []string{"tags", "env=prod", "region=asia"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TableFormatter{}
			var buf bytes.Buffer
			if err := f.Format(&buf, tt.data); err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if tt.wantNoop {
				if buf.Len() != 0 {
					t.Errorf("Format() wrote %q, want no output", buf.String())
				}
				return
			}
			got := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Format() output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestTableFormatter_Format_Nil(t *testing.T) {
	f := &TableFormatter{}
	var buf bytes.Buffer
	if err := f.Format(&buf, nil); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if !strings.Contains(buf.String(), "<nil>") {
		t.Errorf("Format(nil) = %q, want output containing '<nil>'", buf.String())
	}
}

func TestTableFormatter_Format_Scalar(t *testing.T) {
	f := &TableFormatter{}
	var buf bytes.Buffer
	if err := f.Format(&buf, 42); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if !strings.Contains(buf.String(), "42") {
		t.Errorf("Format(42) = %q, want output containing '42'", buf.String())
	}
}

func TestTableFormatter_Format_ObjectArrayWithNestedFields(t *testing.T) {
	f := &TableFormatter{}
	var buf bytes.Buffer
	data := map[string]interface{}{
		"list": []interface{}{
			map[string]interface{}{
				"id":   "a",
				"meta": map[string]interface{}{"region": "us"},
			},
		},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "id") || !strings.Contains(got, "meta") {
		t.Errorf("Format() = %q, want 'id' and 'meta'", got)
	}
}

// TestTableFormatter_BorderStyle verifies the tccli-style box-drawing borders.
func TestTableFormatter_BorderStyle(t *testing.T) {
	f := &TableFormatter{}
	var buf bytes.Buffer
	data := map[string]interface{}{
		"instances": []interface{}{
			map[string]interface{}{"id": "ins-1", "status": "RUNNING"},
			map[string]interface{}{"id": "ins-2", "status": "STOPPED"},
		},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	got := buf.String()

	// Top border should be a run of dashes.
	firstLine := strings.SplitN(got, "\n", 2)[0]
	if !strings.HasPrefix(firstLine, "---") {
		t.Errorf("expected top border of dashes, got first line: %q", firstLine)
	}

	// Output must contain box-drawing characters.
	if !strings.Contains(got, "+") {
		t.Error("expected '+' corners in row separators")
	}
	if !strings.Contains(got, "|") {
		t.Error("expected '|' column dividers")
	}

	// Section title "instances" must be present.
	if !strings.Contains(got, "instances") {
		t.Error("expected section title 'instances'")
	}
}

// TestTableFormatter_VerticalKeyValue verifies that a single-scalar dict
// renders as a 2-column key | value table (no horizontal header).
func TestTableFormatter_VerticalKeyValue(t *testing.T) {
	f := &TableFormatter{}
	var buf bytes.Buffer
	data := map[string]interface{}{"region": "ap-southeast-1"}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "region") || !strings.Contains(got, "ap-southeast-1") {
		t.Errorf("Format() = %q, want 'region' and 'ap-southeast-1'", got)
	}
}

// TestTableFormatter_FieldOrder verifies that columns follow the order specified
// in FieldOrder, with unrecognised keys appended alphabetically at the end.
func TestTableFormatter_FieldOrder(t *testing.T) {
	fieldOrder := map[string][]string{
		"": {"totalCount", "dataSet"},
		"dataSet": {"instanceId", "status", "name"},
	}
	f := &TableFormatter{FieldOrder: fieldOrder}
	var buf bytes.Buffer
	// 5 fields per item forces non-inline (nested table) rendering.
	data := map[string]interface{}{
		"requestId":  "req-abc",
		"totalCount": 1,
		"dataSet": []interface{}{
			map[string]interface{}{
				"name":       "alice",
				"status":     "RUNNING",
				"instanceId": "ins-1",
				"zone":       "ap-southeast-1a",
				"region":     "ap-southeast-1",
			},
		},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	got := buf.String()

	// totalCount must appear before requestId (unknown field → appended last).
	idxTotal := strings.Index(got, "totalCount")
	idxRequest := strings.Index(got, "requestId")
	if idxTotal < 0 || idxRequest < 0 {
		t.Fatalf("expected both totalCount and requestId in output, got:\n%s", got)
	}
	if idxTotal > idxRequest {
		t.Errorf("totalCount (schema field) should appear before requestId (unknown field)")
	}

	// instanceId must appear before name in the dataSet section.
	idxID := strings.Index(got, "instanceId")
	idxName := strings.Index(got, "name")
	if idxID < 0 || idxName < 0 {
		t.Fatalf("expected both instanceId and name in output, got:\n%s", got)
	}
	if idxID > idxName {
		t.Errorf("instanceId (first in schema) should appear before name")
	}
}

// TestTableFormatter_HiddenFields verifies that fields listed in HiddenFields
// are excluded from table output at the top level, while nested fields are
// unaffected and all other top-level fields still appear.
func TestTableFormatter_HiddenFields(t *testing.T) {
	t.Run("hidden field absent from output", func(t *testing.T) {
		f := &TableFormatter{HiddenFields: []string{"requestId"}}
		var buf bytes.Buffer
		data := map[string]interface{}{
			"requestId":  "req-abc",
			"totalCount": 2,
		}
		if err := f.Format(&buf, data); err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		if strings.Contains(got, "requestId") {
			t.Errorf("expected requestId to be hidden, but found it in:\n%s", got)
		}
		if !strings.Contains(got, "totalCount") {
			t.Errorf("expected totalCount to be visible, but not found in:\n%s", got)
		}
	})

	t.Run("nested field with same name is not hidden", func(t *testing.T) {
		f := &TableFormatter{HiddenFields: []string{"requestId"}}
		var buf bytes.Buffer
		data := map[string]interface{}{
			"requestId": "req-abc",
			"items": []interface{}{
				map[string]interface{}{"requestId": "nested-req", "name": "alice"},
			},
		}
		if err := f.Format(&buf, data); err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		// Top-level requestId hidden; nested requestId in items should still appear.
		if !strings.Contains(got, "nested-req") {
			t.Errorf("expected nested requestId value 'nested-req' to appear, got:\n%s", got)
		}
	})

	t.Run("multiple hidden fields", func(t *testing.T) {
		f := &TableFormatter{HiddenFields: []string{"requestId", "totalCount"}}
		var buf bytes.Buffer
		data := map[string]interface{}{
			"requestId":  "req-abc",
			"totalCount": 2,
			"name":       "alice",
		}
		if err := f.Format(&buf, data); err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		if strings.Contains(got, "requestId") {
			t.Errorf("expected requestId to be hidden, got:\n%s", got)
		}
		if strings.Contains(got, "totalCount") {
			t.Errorf("expected totalCount to be hidden, got:\n%s", got)
		}
		if !strings.Contains(got, "alice") {
			t.Errorf("expected name value 'alice' to appear, got:\n%s", got)
		}
	})

	t.Run("no HiddenFields shows all fields", func(t *testing.T) {
		f := &TableFormatter{}
		var buf bytes.Buffer
		data := map[string]interface{}{
			"requestId":  "req-abc",
			"totalCount": 2,
		}
		if err := f.Format(&buf, data); err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, "requestId") {
			t.Errorf("expected requestId to appear when HiddenFields is empty, got:\n%s", got)
		}
	})
}

// TestTableFormatter_ConsistentLineWidth verifies every non-empty line has the
// same rune-width (the global max width), ensuring a clean rectangular table.
func TestTableFormatter_ConsistentLineWidth(t *testing.T) {
	f := &TableFormatter{}
	var buf bytes.Buffer
	data := map[string]interface{}{
		"totalCount": 2,
		"instanceSet": []interface{}{
			map[string]interface{}{"id": "ins-001", "status": "RUNNING"},
			map[string]interface{}{"id": "ins-002", "status": "STOPPED"},
		},
	}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("expected output, got none")
	}

	expected := len([]rune(lines[0]))
	for i, line := range lines {
		if got := len([]rune(line)); got != expected {
			t.Errorf("line %d width = %d, want %d: %q", i, got, expected, line)
		}
	}
}
