package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatTo(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		data    interface{}
		wantErr bool
		contains []string
	}{
		{
			name:     "json format outputs valid JSON",
			format:   "json",
			data:     map[string]string{"key": "value"},
			contains: []string{"key", "value"},
		},
		{
			name:     "table format outputs key-value table",
			format:   "table",
			data:     map[string]string{"key": "value"},
			contains: []string{"key", "value"},
		},
		{
			name:    "unsupported format returns error",
			format:  "xml",
			data:    "data",
			wantErr: true,
		},
		{
			name:    "empty format string returns error",
			format:  "",
			data:    "data",
			wantErr: true,
		},
		{
			name:     "json format with struct",
			format:   "json",
			data:     struct{ Name string }{Name: "Alice"},
			contains: []string{"Name", "Alice"},
		},
		{
			name:     "table format with nil data falls back to string",
			format:   "table",
			data:     nil,
			contains: []string{"<nil>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := FormatTo(&buf, tt.format, tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			got := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("FormatTo() output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestFormatTo_UnsupportedErrorMessage(t *testing.T) {
	var buf bytes.Buffer
	err := FormatTo(&buf, "yaml", nil)
	if err == nil {
		t.Fatal("FormatTo() expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "yaml") {
		t.Errorf("error message %q should contain the format name 'yaml'", err.Error())
	}
}

func TestPrintWithFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		data    interface{}
		wantErr bool
	}{
		{
			name:   "valid json format",
			format: "json",
			data:   map[string]string{"hello": "world"},
		},
		{
			name:   "valid table format",
			format: "table",
			data:   map[string]string{"hello": "world"},
		},
		{
			name:    "invalid format returns error",
			format:  "csv",
			data:    "data",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrintWithFormat(tt.format, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintWithFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	err := Print(map[string]string{"test": "value"})
	if err != nil {
		t.Errorf("Print() error = %v", err)
	}
}
