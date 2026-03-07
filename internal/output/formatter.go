// Package output provides output formatting for CLI responses.
package output

import (
	"fmt"
	"io"
	"os"
)

// Formatter defines the interface for output formatters.
type Formatter interface {
	// Format formats the data and writes to the writer.
	Format(w io.Writer, data interface{}) error
}

// Format formats data according to the specified format type.
func Format(format string, data interface{}) error {
	return FormatTo(os.Stdout, format, data)
}

// FormatTo formats data and writes to the specified writer.
func FormatTo(w io.Writer, format string, data interface{}) error {
	var formatter Formatter

	switch format {
	case "json":
		formatter = &JSONFormatter{Indent: true}
	case "table":
		formatter = &TableFormatter{}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	return formatter.Format(w, data)
}

// Print prints data using the default format (json).
func Print(data interface{}) error {
	return Format("json", data)
}

// PrintWithFormat prints data using the specified format.
func PrintWithFormat(format string, data interface{}) error {
	return Format(format, data)
}
