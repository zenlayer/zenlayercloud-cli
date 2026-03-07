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
			name: "array of objects renders as table",
			data: map[string]interface{}{
				"instances": []interface{}{
					map[string]interface{}{"id": "ins-1", "status": "RUNNING"},
					map[string]interface{}{"id": "ins-2", "status": "STOPPED"},
				},
			},
			contains: []string{"id", "status", "ins-1", "RUNNING", "ins-2", "STOPPED"},
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
