package deftemplate

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"kdp-oam-operator/api/bdc/common"
	"kdp-oam-operator/pkg/controllers/bdc/defcontext"
	"reflect"
	"testing"
)

var testTemplateTemp = `output: {
  apiVersion: "v1"
  kind: "ConfigMap"
  metadata: {
    name: context.name
    namespace: context.namespace
    annotations: context.bdcAnnotations
  }
  data: {
    "host": parameter.host
    "hostname": parameter.hostname
    "port": parameter.port
  }
}
parameter: {
  host: string
  hostname: string
  port: string
}
`

func TestBigDataClusterDef_RenderCUETemplate(t *testing.T) {
	contextData := defcontext.ContextData{
		Namespace:      "kdp-test",
		Name:           "app-config",
		Cluster:        "",
		BDCName:        "kdp-bdc-name",
		BDCLabels:      map[string]string{},
		BDCAnnotations: map[string]string{},
		Ctx:            nil,
	}
	contextData = defcontext.NewBDCContext(contextData)

	type fields struct {
		def def
	}
	type args struct {
		ctx              defcontext.ContextData
		abstractTemplate string
		params           interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "TestRenderCUETemplate",
			fields: fields{
				def: def{
					name: "test",
				},
			},
			args: args{
				ctx:              contextData,
				abstractTemplate: testTemplateTemp,
				params: map[string]interface{}{
					"host":     "zookeeper.kdp-test.svc.cluster.local:2181",
					"hostname": "zookeeper.kdp-test.svc.cluster.local",
					"port":     "2181",
				},
			},
			want: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":        "app-config",
							"namespace":   "kdp-test",
							"annotations": map[string]interface{}{},
						},
						"data": map[string]interface{}{
							"host":     "zookeeper.kdp-test.svc.cluster.local:2181",
							"hostname": "zookeeper.kdp-test.svc.cluster.local",
							"port":     "2181",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wd := &BigDataClusterDef{
				def: tt.fields.def,
			}
			got, err := wd.RenderCUETemplate(tt.args.ctx, tt.args.abstractTemplate, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderCUETemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RenderCUETemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var bdcTemplateTest = `output: {
	apiVersion: "v1"
	kind:       "Namespace"
	metadata: {
		name: parameter.namespaces[0].name
		annotations: context.bdcAnnotations
        labels: context.bdcLabels

	}
}
outputs: {
	for i, v in parameter.namespaces {
		if i > 0 {
			"objects-\(i)": {
				apiVersion: "v1"
				kind:       "Namespace"
				metadata: {
					name: v.name
					annotations: "bdc.kdp.io/name": context.bdcName
				}
			}
		}
	}
}

parameter: {
	frozen?:   *false | bool
	disabled?: *false | bool
	namespaces: [...{
		name:      string
		isDefault: bool
	},
	]
}
`

func TestConvertTemplateJSON2Object(t *testing.T) {
	type args struct {
		capabilityName string
		in             *runtime.RawExtension
		schematic      *common.Schematic
	}
	tests := []struct {
		name    string
		args    args
		want    common.Capability
		wantErr bool
	}{
		{
			name: "TestConvertTemplateJSON2Object",
			args: args{
				capabilityName: "test",
				in:             nil,
				schematic: &common.Schematic{
					CUE: &common.CUE{
						Template: bdcTemplateTest,
					},
				},
			},
			want: common.Capability{
				Name:        "test",
				CueTemplate: bdcTemplateTest,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertTemplateJSON2Object(tt.args.capabilityName, tt.args.in, tt.args.schematic)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTemplateJSON2Object() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertTemplateJSON2Object() got = %v, want %v", got, tt.want)
			}
		})
	}
}
