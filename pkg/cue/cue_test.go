package cue

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"testing"
)

func TestUnstructured(t *testing.T) {
	c := cuecontext.New()

	// 准备测试数据
	cueStr1 := `{ "kind": "TestKind", "apiVersion": "v1", "metadata": {"name": "test"}, "spec": {} }`
	validValue1 := c.CompileString(cueStr1)

	cueStr2 := `{ "kind": "TestKind", "apiVersion": "v1", "metadata": {"name2": "test2"}, "spec": {} }`
	validValue2 := c.CompileString(cueStr2)

	cuestr3 := `{"name": "name", "value": "value" }`
	invalidValue := c.CompileString(cuestr3)

	type args struct {
		v cue.Value
	}
	tests := []struct {
		name    string
		args    args
		want    *unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "valid input",
			args: args{
				v: validValue1,
			},
			want:    &unstructured.Unstructured{Object: map[string]interface{}{"kind": "TestKind", "apiVersion": "v1", "metadata": map[string]interface{}{"name": "test"}, "spec": map[string]interface{}{}}},
			wantErr: false,
		},
		{
			name: "valid input2",
			args: args{
				v: validValue2,
			},
			want:    &unstructured.Unstructured{Object: map[string]interface{}{"kind": "TestKind", "apiVersion": "v1", "metadata": map[string]interface{}{"name2": "test2"}, "spec": map[string]interface{}{}}},
			wantErr: false,
		},
		{
			name: "invalid input",
			args: args{
				v: invalidValue,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unstructured(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unstructured() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unstructured() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndCompile(t *testing.T) {

	validCueString := `{"name": "valid", "value": 100}`
	invalidCueString := `{"name": "invalid", "value": _|_ }` // 这是一个无效的 CUE 值

	// 从字符串创建有效和无效的cue.Value
	validCueValue, _ := cueValueFromString(validCueString)
	invalidCueValue, _ := cueValueFromString(invalidCueString)

	type args struct {
		v cue.Value
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ValidCue",
			args: args{
				v: validCueValue,
			},
			want:    []byte(`{"name":"valid","value":100}`),
			wantErr: false,
		},
		{
			name: "InvalidCue",
			args: args{
				v: invalidCueValue,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndCompile(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndCompile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateAndCompile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func cueValueFromString(s string) (cue.Value, error) {
	ctx := cuecontext.New()
	v := ctx.CompileString(s)
	return v, v.Err()
}
