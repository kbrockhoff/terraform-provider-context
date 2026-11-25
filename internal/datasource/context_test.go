package datasource

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringList(t *testing.T, values ...string) types.List {
	t.Helper()
	elements := make([]attr.Value, len(values))
	for i, v := range values {
		elements[i] = types.StringValue(v)
	}
	list, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("ListValue: %v", diags)
	}
	return list
}

func stringMap(t *testing.T, values map[string]string) types.Map {
	t.Helper()
	elements := make(map[string]attr.Value, len(values))
	for k, v := range values {
		elements[k] = types.StringValue(v)
	}
	m, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("MapValue: %v", diags)
	}
	return m
}

func TestMergeStringValue(t *testing.T) {
	tests := []struct {
		name   string
		child  types.String
		parent types.String
		want   string
	}{
		{name: "child set", child: types.StringValue("child"), parent: types.StringValue("parent"), want: "child"},
		{name: "parent only", child: types.StringNull(), parent: types.StringValue("parent"), want: "parent"},
		{name: "neither", child: types.StringNull(), parent: types.StringNull(), want: ""},
		{name: "child empty string wins", child: types.StringValue(""), parent: types.StringValue("parent"), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeStringValue(tt.child, tt.parent); got != tt.want {
				t.Errorf("mergeStringValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMergeBoolValue(t *testing.T) {
	tests := []struct {
		name   string
		child  types.Bool
		parent types.Bool
		def    bool
		want   bool
	}{
		{name: "child set", child: types.BoolValue(false), parent: types.BoolValue(true), def: true, want: false},
		{name: "parent only", child: types.BoolNull(), parent: types.BoolValue(false), def: true, want: false},
		{name: "neither uses default", child: types.BoolNull(), parent: types.BoolNull(), def: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeBoolValue(tt.child, tt.parent, tt.def); got != tt.want {
				t.Errorf("mergeBoolValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeListValue(t *testing.T) {
	ctx := context.Background()
	nullList := types.ListNull(types.StringType)

	tests := []struct {
		name   string
		child  types.List
		parent types.List
		want   []string
	}{
		{name: "child set", child: stringList(t, "c"), parent: stringList(t, "p"), want: []string{"c"}},
		{name: "parent only", child: nullList, parent: stringList(t, "p1", "p2"), want: []string{"p1", "p2"}},
		{name: "neither", child: nullList, parent: nullList, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeListValue(ctx, tt.child, tt.parent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeListValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeMapValue(t *testing.T) {
	ctx := context.Background()
	nullMap := types.MapNull(types.StringType)

	tests := []struct {
		name   string
		child  types.Map
		parent types.Map
		want   map[string]string
	}{
		{
			name:   "child key wins over parent",
			child:  stringMap(t, map[string]string{"team": "api", "tier": "1"}),
			parent: stringMap(t, map[string]string{"team": "platform", "owner": "ops"}),
			want:   map[string]string{"team": "api", "tier": "1", "owner": "ops"},
		},
		{
			name:   "parent only",
			child:  nullMap,
			parent: stringMap(t, map[string]string{"owner": "ops"}),
			want:   map[string]string{"owner": "ops"},
		},
		{name: "neither", child: nullMap, parent: nullMap, want: map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeMapValue(ctx, tt.child, tt.parent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeMapValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextDataSourceSchema(t *testing.T) {
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	NewContextDataSource().Schema(ctx, datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics: %v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema implementation invalid: %v", diags)
	}

	for _, name := range []string{"data_sovereignty", "environment_type", "deletion_date", "parent_context", "context_output"} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}
