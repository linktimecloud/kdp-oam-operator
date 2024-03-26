package defcontext

import (
	"context"
	"reflect"
	"testing"
)

func TestContextData_BaseContextFile(t *testing.T) {
	type fields struct {
		Namespace      string
		Name           string
		Cluster        string
		BDCName        string
		BDCLabels      map[string]string
		BDCAnnotations map[string]string
		Ctx            context.Context
		data           map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "BaseContextFile1",
			fields: fields{
				Namespace:      "namespace",
				Name:           "name",
				Cluster:        "cluster",
				BDCName:        "bdcname",
				BDCLabels:      map[string]string{"key_label": "value_label"},
				BDCAnnotations: map[string]string{"key_anno": "value_anno"},
				Ctx:            context.Background(),
				data: map[string]interface{}{
					"key_data":  "value_data",
					"key1_data": "value1_data",
				},
			},
			want: "context: {\"key1_data\":\"value1_data\",\"key_data\":\"value_data\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextData{
				Namespace:      tt.fields.Namespace,
				Name:           tt.fields.Name,
				Cluster:        tt.fields.Cluster,
				BDCName:        tt.fields.BDCName,
				BDCLabels:      tt.fields.BDCLabels,
				BDCAnnotations: tt.fields.BDCAnnotations,
				Ctx:            tt.fields.Ctx,
				data:           tt.fields.data,
			}
			got, err := ctx.BaseContextFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseContextFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BaseContextFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextData_BaseContextLabels(t *testing.T) {
	type fields struct {
		Namespace      string
		Name           string
		Cluster        string
		BDCName        string
		BDCLabels      map[string]string
		BDCAnnotations map[string]string
		Ctx            context.Context
		data           map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "BaseContextLabels1",
			fields: fields{
				Namespace:      "namespace",
				Name:           "name",
				Cluster:        "cluster",
				BDCName:        "bdcname",
				BDCLabels:      map[string]string{"key_label": "value_label"},
				BDCAnnotations: map[string]string{"key_anno": "value_anno"},
				Ctx:            context.Background(),
				data: map[string]interface{}{
					"key_data":  "value_data",
					"key1_data": "value1_data",
					"name":      "name",
				},
			},
			want: map[string]string{"name": "name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextData{
				Namespace:      tt.fields.Namespace,
				Name:           tt.fields.Name,
				Cluster:        tt.fields.Cluster,
				BDCName:        tt.fields.BDCName,
				BDCLabels:      tt.fields.BDCLabels,
				BDCAnnotations: tt.fields.BDCAnnotations,
				Ctx:            tt.fields.Ctx,
				data:           tt.fields.data,
			}
			if got := ctx.BaseContextLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseContextLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextData_GetData(t *testing.T) {
	type fields struct {
		Namespace      string
		Name           string
		Cluster        string
		BDCName        string
		BDCLabels      map[string]string
		BDCAnnotations map[string]string
		Ctx            context.Context
		data           map[string]interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "GetData1",
			fields: fields{
				Namespace:      "namespace",
				Name:           "name",
				Cluster:        "cluster",
				BDCName:        "bdcname",
				BDCLabels:      map[string]string{"key_label": "value_label"},
				BDCAnnotations: map[string]string{"key_anno": "value_anno"},
				Ctx:            context.Background(),
				data: map[string]interface{}{
					"key_data":  "value_data",
					"key1_data": "value1_data",
				},
			},
			args: args{key: "key1_data"},
			want: "value1_data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextData{
				Namespace:      tt.fields.Namespace,
				Name:           tt.fields.Name,
				Cluster:        tt.fields.Cluster,
				BDCName:        tt.fields.BDCName,
				BDCLabels:      tt.fields.BDCLabels,
				BDCAnnotations: tt.fields.BDCAnnotations,
				Ctx:            tt.fields.Ctx,
				data:           tt.fields.data,
			}
			if got := ctx.GetData(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextData_PushData(t *testing.T) {
	type fields struct {
		Namespace      string
		Name           string
		Cluster        string
		BDCName        string
		BDCLabels      map[string]string
		BDCAnnotations map[string]string
		Ctx            context.Context
		data           map[string]interface{}
	}
	type args struct {
		key  string
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "PushData1",
			fields: fields{
				Namespace:      "namespace",
				Name:           "name",
				Cluster:        "cluster",
				BDCName:        "bdcname",
				BDCLabels:      map[string]string{"key_label": "value_label"},
				BDCAnnotations: map[string]string{"key_anno": "value_anno"},
				Ctx:            context.Background(),
				data: map[string]interface{}{
					"key_data":  "value_data",
					"key1_data": "value1_data",
				},
			},
			args: args{
				key:  "test",
				data: "test_data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextData{
				Namespace:      tt.fields.Namespace,
				Name:           tt.fields.Name,
				Cluster:        tt.fields.Cluster,
				BDCName:        tt.fields.BDCName,
				BDCLabels:      tt.fields.BDCLabels,
				BDCAnnotations: tt.fields.BDCAnnotations,
				Ctx:            tt.fields.Ctx,
				data:           tt.fields.data,
			}
			ctx.PushData(tt.args.key, tt.args.data)
		})
	}
}

func TestNewBDCContext(t *testing.T) {
	type args struct {
		cd ContextData
	}
	tests := []struct {
		name string
		args args
		want ContextData
	}{
		{
			name: "NewBDCContext1",
			args: args{cd: ContextData{
				Namespace:      "namespace",
				Name:           "name",
				Cluster:        "cluster",
				BDCName:        "bdcname",
				BDCLabels:      map[string]string{"key_label": "value_label"},
				BDCAnnotations: map[string]string{"key_anno": "value_anno"},
				Ctx:            context.Background(),
				data: map[string]interface{}{
					"key_data": "value_data",
				}},
			},
			want: ContextData{
				Namespace:      "namespace",
				Name:           "name",
				Cluster:        "cluster",
				BDCName:        "bdcname",
				BDCLabels:      map[string]string{"key_label": "value_label"},
				BDCAnnotations: map[string]string{"key_anno": "value_anno"},
				Ctx:            context.Background(),
				data: map[string]interface{}{
					"key_data":       "value_data",
					"name":           "name",
					"bdcName":        "bdcname",
					"bdcLabels":      map[string]string{"key_label": "value_label"},
					"bdcAnnotations": map[string]string{"key_anno": "value_anno"},
					"namespace":      "namespace",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBDCContext(tt.args.cd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBDCContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
