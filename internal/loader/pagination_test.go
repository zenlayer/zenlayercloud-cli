package loader

import (
	"testing"
)

func TestIsPaginatedDef(t *testing.T) {
	paginatedDef := &APIDefinition{
		Parameters: []Parameter{
			{Name: "page-num", SDKField: "pageNum", Type: "integer"},
			{Name: "page-size", SDKField: "pageSize", Type: "integer"},
			{Name: "zone-id", SDKField: "zoneId", Type: "string"},
		},
	}
	nonPaginatedDef := &APIDefinition{
		Parameters: []Parameter{
			{Name: "zone-id", SDKField: "zoneId", Type: "string"},
		},
	}
	pageNumOnlyDef := &APIDefinition{
		Parameters: []Parameter{
			{Name: "page-num", SDKField: "pageNum", Type: "integer"},
		},
	}

	if !isPaginatedDef(paginatedDef) {
		t.Error("expected def with page-num and page-size to be paginated")
	}
	if isPaginatedDef(nonPaginatedDef) {
		t.Error("expected def without pagination params to be non-paginated")
	}
	if isPaginatedDef(pageNumOnlyDef) {
		t.Error("expected def with only page-num (no page-size) to be non-paginated")
	}
}

func TestFindDataArrayField(t *testing.T) {
	tests := []struct {
		name      string
		resp      map[string]interface{}
		wantField string
		wantLen   int
	}{
		{
			name: "instance set",
			resp: map[string]interface{}{
				"requestId":   "req-1",
				"totalCount":  float64(2),
				"instanceSet": []interface{}{"a", "b"},
			},
			wantField: "instanceSet",
			wantLen:   2,
		},
		{
			name: "data set",
			resp: map[string]interface{}{
				"requestId":  "req-2",
				"totalCount": float64(1),
				"dataSet":    []interface{}{"x"},
			},
			wantField: "dataSet",
			wantLen:   1,
		},
		{
			name: "no array field",
			resp: map[string]interface{}{
				"requestId":  "req-3",
				"totalCount": float64(0),
			},
			wantField: "",
			wantLen:   0,
		},
		{
			name: "metadata arrays skipped",
			resp: map[string]interface{}{
				"requestId":   "req-4",
				"totalCount":  float64(1),
				"bandwidthSet": []interface{}{"b1"},
			},
			wantField: "bandwidthSet",
			wantLen:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field, arr := findDataArrayField(tc.resp)
			if field != tc.wantField {
				t.Errorf("field = %q, want %q", field, tc.wantField)
			}
			if len(arr) != tc.wantLen {
				t.Errorf("len(arr) = %d, want %d", len(arr), tc.wantLen)
			}
		})
	}
}
