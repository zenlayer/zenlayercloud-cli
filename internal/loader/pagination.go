package loader

import (
	"fmt"

	"github.com/zenlayer/zenlayercloud-cli/internal/apiclient"
)

// metadataFields are response fields that are never the paginated data array.
var metadataFields = map[string]bool{
	"requestId":  true,
	"totalCount": true,
	"pageSize":   true,
	"pageNum":    true,
}

// defaultPageSize is used when the caller did not specify --page-size.
const defaultPageSize = 100

// isPaginatedDef reports whether def has both page-num and page-size parameters,
// indicating the API supports offset pagination.
func isPaginatedDef(def *APIDefinition) bool {
	hasPageNum, hasPageSize := false, false
	for i := range def.Parameters {
		p := &def.Parameters[i]
		if p.Name == "page-num" || p.SDKField == "pageNum" {
			hasPageNum = true
		}
		if p.Name == "page-size" || p.SDKField == "pageSize" {
			hasPageSize = true
		}
	}
	return hasPageNum && hasPageSize
}

// findDataArrayField returns the name and value of the first array field in resp
// that is not a known metadata field. Returns ("", nil) when none is found.
func findDataArrayField(resp map[string]interface{}) (string, []interface{}) {
	for k, v := range resp {
		if metadataFields[k] {
			continue
		}
		if arr, ok := v.([]interface{}); ok {
			return k, arr
		}
	}
	return "", nil
}

// fetchAllPages repeatedly calls the API, incrementing pageNum each time, and
// merges the data array from every page into a single response map.
//
// Rules:
//   - pageNum always starts at 1 (ignores any value in params).
//   - pageSize is taken from params["pageSize"] if set; otherwise defaultPageSize.
//   - Stops when accumulated count >= totalCount OR when the page returns fewer
//     items than pageSize (signals the last page).
//   - If the response has no array field the first page is returned as-is.
func fetchAllPages(client *apiclient.CommonClient, def *APIDefinition, params map[string]interface{}) (map[string]interface{}, error) {
	// Determine batch size.
	pageSize := defaultPageSize
	if ps, ok := params["pageSize"]; ok {
		switch v := ps.(type) {
		case int:
			if v > 0 {
				pageSize = v
			}
		case float64:
			if v > 0 {
				pageSize = int(v)
			}
		}
	}

	params["pageSize"] = pageSize

	var (
		allItems   []interface{}
		firstResp  map[string]interface{}
		arrayField string
		totalCount int
	)

	for page := 1; ; page++ {
		params["pageNum"] = page

		resp, err := client.Call(def.SDK.Service, def.SDK.Version, def.SDK.Action, params)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}

		if page == 1 {
			firstResp = resp

			// Extract total count from first response.
			if tc, ok := resp["totalCount"]; ok {
				switch v := tc.(type) {
				case float64:
					totalCount = int(v)
				case int:
					totalCount = v
				}
			}

			// Locate the data array field.
			field, items := findDataArrayField(resp)
			if field == "" {
				// No array field found; return single-page response unchanged.
				return resp, nil
			}
			arrayField = field
			allItems = append(allItems, items...)
		} else {
			arr, _ := resp[arrayField].([]interface{})
			allItems = append(allItems, arr...)

			// Stop if this page was short (last page).
			if len(arr) < pageSize {
				break
			}
		}

		// Stop once we have collected everything.
		if totalCount > 0 && len(allItems) >= totalCount {
			break
		}

		// First-page short-circuit (totalCount unknown or zero).
		if page == 1 {
			items := firstResp[arrayField].([]interface{})
			if len(items) < pageSize {
				break
			}
		}
	}

	// Build merged response: copy first response and replace the array field.
	merged := make(map[string]interface{}, len(firstResp))
	for k, v := range firstResp {
		merged[k] = v
	}
	merged[arrayField] = allItems
	return merged, nil
}
