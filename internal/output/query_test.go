package output

import (
	"reflect"
	"testing"
)

func TestApplyQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		data    interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:  "empty query returns data unchanged",
			query: "",
			data:  map[string]interface{}{"a": 1, "b": "two"},
			want:  map[string]interface{}{"a": 1, "b": "two"},
		},
		{
			name:  "extract single field",
			query: "a",
			data:  map[string]interface{}{"a": 1, "b": "two"},
			want:  1.0,
		},
		{
			name:  "nested path",
			query: "foo.bar",
			data:  map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
			want:  "baz",
		},
		{
			name:  "array projection",
			query: "dataSet[*].id",
			data: map[string]interface{}{
				"dataSet": []interface{}{
					map[string]interface{}{"id": "i1", "name": "n1"},
					map[string]interface{}{"id": "i2", "name": "n2"},
				},
			},
			want: []interface{}{"i1", "i2"},
		},
		{
			name:  "filter expression",
			query: "dataSet[?state=='RUNNING']",
			data: map[string]interface{}{
				"dataSet": []interface{}{
					map[string]interface{}{"id": "i1", "state": "RUNNING"},
					map[string]interface{}{"id": "i2", "state": "STOPPED"},
				},
			},
			want: []interface{}{
				map[string]interface{}{"id": "i1", "state": "RUNNING"},
			},
		},
		{
			name:    "invalid query returns error",
			query:   "foo[",
			data:    map[string]interface{}{"foo": 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyQuery(tt.query, tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ApplyQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
