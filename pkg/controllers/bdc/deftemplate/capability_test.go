package deftemplate

import (
	"kdp-oam-operator/api/bdc/common"
	"kdp-oam-operator/api/bdc/v1alpha1"
	"reflect"
	"testing"
)

var testTemplateCap = `output: {
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

func TestCapabilityDefinition_GetOpenAPIAndUischemaSchema(t *testing.T) {
	type fields struct {
		Name                     string
		XDefinition              v1alpha1.XDefinition
		CapabilityBaseDefinition CapabilityBaseDefinition
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		want1   []byte
		wantErr bool
	}{
		{
			name: "Test_GetOpenAPIAndUischemaSchema",
			fields: fields{
				Name: "Test_GetOpenAPIAndUischemaSchema",
				XDefinition: v1alpha1.XDefinition{
					Spec: v1alpha1.XDefinitionSpec{
						Schematic: &common.Schematic{
							CUE: &common.CUE{
								Template: testTemplateCap,
							},
						},
					},
				},
			},
			args: args{
				name: "Test_GetOpenAPIAndUischemaSchema",
			},
			want:    []byte(`{"properties":{"host":{"type":"string"},"hostname":{"type":"string"},"port":{"type":"string"}},"required":["host","hostname","port"],"type":"object"}`),
			want1:   []byte(`{"host":{},"hostname":{},"port":{}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &CapabilityDefinition{
				Name:                     tt.fields.Name,
				XDefinition:              tt.fields.XDefinition,
				CapabilityBaseDefinition: tt.fields.CapabilityBaseDefinition,
			}
			got, got1, err := def.GetOpenAPIAndUischemaSchema(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOpenAPIAndUischemaSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOpenAPIAndUischemaSchema() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetOpenAPIAndUischemaSchema() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
