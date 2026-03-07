package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter formats output as JSON.
type JSONFormatter struct {
	// Indent enables pretty-printing with indentation.
	Indent bool
}

// Format implements Formatter interface for JSON output.
func (f *JSONFormatter) Format(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}
