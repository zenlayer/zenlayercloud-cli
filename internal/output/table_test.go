package output

import (
	"bytes"
	"strings"
	"testing"
)

type testPerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type testNoTag struct {
	ID   int
	Name string
}

func TestTableFormatter_Format_MapStringString(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		wantOut  string
		wantNoop bool
	}{
		{
			name:     "empty map returns no output",
			data:     map[string]string{},
			wantNoop: true,
		},
		{
			name: "single key-value pair",
			data: map[string]string{"key": "value"},
			wantOut: "Key  Value\n" +
				"---  -----\n" +
				"key  value\n",
		},
		{
			name: "multiple key-value pairs sorted alphabetically",
			data: map[string]string{"name": "Alice", "age": "30"},
			wantOut: "Key   Value\n" +
				"----  -----\n" +
				"age   30\n" +
				"name  Alice\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TableFormatter{}
			var buf bytes.Buffer
			err := f.Format(&buf, tt.data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if tt.wantNoop {
				if buf.Len() != 0 {
					t.Errorf("Format() wrote %q, want no output", buf.String())
				}
				return
			}
			if got := buf.String(); got != tt.wantOut {
				t.Errorf("Format() =\n%q\nwant\n%q", got, tt.wantOut)
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
			contains: []string{"key", "value", "Key", "Value"},
		},
		{
			name:     "integer value",
			data:     map[string]interface{}{"count": 42},
			contains: []string{"count", "42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TableFormatter{}
			var buf bytes.Buffer
			err := f.Format(&buf, tt.data)
			if err != nil {
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

func TestTableFormatter_Format_Struct(t *testing.T) {
	t.Run("struct with json tags", func(t *testing.T) {
		f := &TableFormatter{}
		var buf bytes.Buffer
		err := f.Format(&buf, testPerson{Name: "Alice", Age: 30})
		if err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		want := "Key   Value\n" +
			"----  -----\n" +
			"name  Alice\n" +
			"age   30\n"
		if got != want {
			t.Errorf("Format() =\n%q\nwant\n%q", got, want)
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		f := &TableFormatter{}
		var buf bytes.Buffer
		err := f.Format(&buf, &testPerson{Name: "Bob", Age: 25})
		if err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, "Bob") || !strings.Contains(got, "25") {
			t.Errorf("Format() = %q, want output containing 'Bob' and '25'", got)
		}
	})

	t.Run("struct without json tags uses field name", func(t *testing.T) {
		f := &TableFormatter{}
		var buf bytes.Buffer
		err := f.Format(&buf, testNoTag{ID: 1, Name: "Charlie"})
		if err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, "ID") || !strings.Contains(got, "Name") {
			t.Errorf("Format() = %q, want output containing field names 'ID' and 'Name'", got)
		}
		if !strings.Contains(got, "Charlie") || !strings.Contains(got, "1") {
			t.Errorf("Format() = %q, want output containing values '1' and 'Charlie'", got)
		}
	})

	t.Run("non-struct falls back to string representation", func(t *testing.T) {
		f := &TableFormatter{}
		var buf bytes.Buffer
		err := f.Format(&buf, 42)
		if err != nil {
			t.Fatalf("Format() error = %v", err)
		}
		got := buf.String()
		if got != "42\n" {
			t.Errorf("Format() = %q, want %q", got, "42\n")
		}
	})
}

func TestTableFormatter_ImplementsFormatter(t *testing.T) {
	var _ Formatter = &TableFormatter{}
}
