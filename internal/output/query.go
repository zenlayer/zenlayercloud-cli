// Package output provides output formatting for CLI responses.
package output

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
)

// ApplyQuery applies a JMESPath query to the data. If query is empty, returns data unchanged.
// JMESPath is a query language for JSON - see https://jmespath.org/ for syntax.
func ApplyQuery(query string, data interface{}) (interface{}, error) {
	if query == "" {
		return data, nil
	}
	result, err := jmespath.Search(query, data)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return result, nil
}
