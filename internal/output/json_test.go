package output

import (
	"bytes"
	"testing"
)

func TestJSONFormatter_Format(t *testing.T) {
	tests := []struct {
		name    string
		indent  bool
		data    interface{}
		want    string
		wantErr bool
	}{
		{
			name:   "map with indent",
			indent: true,
			data:   map[string]string{"key": "value"},
			want:   "{\n  \"key\": \"value\"\n}\n",
		},
		{
			name:   "map without indent",
			indent: false,
			data:   map[string]string{"key": "value"},
			want:   "{\"key\":\"value\"}\n",
		},
		{
			name:   "nil data",
			indent: false,
			data:   nil,
			want:   "null\n",
		},
		{
			name:   "string data",
			indent: false,
			data:   "hello",
			want:   "\"hello\"\n",
		},
		{
			name:   "integer data",
			indent: false,
			data:   42,
			want:   "42\n",
		},
		{
			name:   "boolean true",
			indent: false,
			data:   true,
			want:   "true\n",
		},
		{
			name:   "boolean false",
			indent: false,
			data:   false,
			want:   "false\n",
		},
		{
			name:   "slice data",
			indent: false,
			data:   []int{1, 2, 3},
			want:   "[1,2,3]\n",
		},
		{
			name:   "nested map with indent",
			indent: true,
			data:   map[string]interface{}{"a": 1, "b": "two"},
			want:   "{\n  \"a\": 1,\n  \"b\": \"two\"\n}\n",
		},
		{
			name:   "empty struct",
			indent: false,
			data:   struct{}{},
			want:   "{}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &JSONFormatter{Indent: tt.indent}
			var buf bytes.Buffer
			err := f.Format(&buf, tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Format() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got := buf.String(); got != tt.want {
					t.Errorf("Format() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestJSONFormatter_ImplementsFormatter(t *testing.T) {
	var _ Formatter = &JSONFormatter{}
}
