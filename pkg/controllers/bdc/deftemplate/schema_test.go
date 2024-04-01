package deftemplate

import (
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"strconv"
	"testing"
)

func TestFixOpenAPISchema(t *testing.T) {
	tenFloat64, _ := strconv.ParseFloat("10", 64)
	tenUnit64, _ := strconv.ParseUint("10", 10, 64)
	type args struct {
		schema *openapi3.Schema
	}
	tests := []struct {
		name string
		args args
		want *openapi3.Schema
	}{
		{
			name: "Test description and title correct",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+title=Example Title\n+description=Example Description\n+minimum=10\n+maximum=10\n+minLength=10\n+maxLength=10\n+pattern=^\\d{3}$",
						},
					},
				},
			}},
			want: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Title:       "Example Title",
							Description: "Example Description",
							Min:         &tenFloat64,
							Max:         &tenFloat64,
							MinLength:   tenUnit64,
							MaxLength:   &tenUnit64,
							Pattern:     "^\\d{3}$",
						},
					},
				},
			},
		},
		{
			name: "Test description and title with wrong min",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+title=Example Title\n+description=Example Description\n+minimum=x\n+maximum=10\n+minLength=10\n+maxLength=10\n+pattern=^\\d{3}$",
						},
					},
				},
			}},
			want: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Title:       "Example Title",
							Description: "Example Description",
							//Min:         &tenFloat64,
							Max:       &tenFloat64,
							MinLength: tenUnit64,
							MaxLength: &tenUnit64,
							Pattern:   "^\\d{3}$",
						},
					},
				},
			},
		},
		{
			name: "Test description and title with wrong max",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+title=Example Title\n+description=Example Description\n+minimum=10\n+maximum=x\n+minLength=10\n+maxLength=10\n+pattern=^\\d{3}$",
						},
					},
				},
			}},
			want: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Title:       "Example Title",
							Description: "Example Description",
							Min:         &tenFloat64,
							//Max:       &tenFloat64,
							MinLength: tenUnit64,
							MaxLength: &tenUnit64,
							Pattern:   "^\\d{3}$",
						},
					},
				},
			},
		},
		{
			name: "Test description and title with wrong minLength",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+title=Example Title\n+description=Example Description\n+minimum=10\n+maximum=10\n+minLength=x\n+maxLength=10\n+pattern=^\\d{3}$",
						},
					},
				},
			}},
			want: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Title:       "Example Title",
							Description: "Example Description",
							Min:         &tenFloat64,
							Max:         &tenFloat64,
							//MinLength: tenUnit64,
							MaxLength: &tenUnit64,
							Pattern:   "^\\d{3}$",
						},
					},
				},
			},
		},
		{
			name: "Test description and title with wrong maxLength",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+title=Example Title\n+description=Example Description\n+minimum=10\n+maximum=10\n+minLength=10\n+maxLength=x\n+pattern=^\\d{3}$",
						},
					},
				},
			}},
			want: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Title:       "Example Title",
							Description: "Example Description",
							Min:         &tenFloat64,
							Max:         &tenFloat64,
							MinLength:   tenUnit64,
							//MaxLength: &tenUnit64,
							Pattern: "^\\d{3}$",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if FixOpenAPISchema(tt.args.schema); !reflect.DeepEqual(tt.args.schema, tt.want) {
				t.Errorf("FixOpenAPISchema() = %v, want %v", tt.args.schema, tt.want)
			}
		})
	}
}

func TestGetAdditionUiSchema(t *testing.T) {
	type args struct {
		schema *openapi3.Schema
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Test ui order",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+ui:order=1",
						},
					},
				},
			}},
			want: []byte(`{"exampleProperty":{},"ui:order":["exampleProperty","*"]}`),
		},
		{
			name: "Test ui order",
			args: args{schema: &openapi3.Schema{
				Type: "object",
				Properties: map[string]*openapi3.SchemaRef{
					"exampleProperty": {
						Value: &openapi3.Schema{
							Description: "+ui:order=1\n+ui:title=example tile\n+ui:description=example description\n+ui:widget=example widget\n+ui:options=example options\n+ui:placeholder=example placeholder\n+ui:description=example description\n+ui:hidden={{rootFormData.hive.enable == false}}\n+err:options={\"required\":\"example err options\"}",
						},
					},
				},
			}},
			want: []byte(`{"exampleProperty":{"err:options":{"required":"example err options"},"ui:description":"example description","ui:hidden":"{{rootFormData.hive.enable == false}}","ui:title":"example tile"},"ui:order":["exampleProperty","*"]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAdditionUiSchema(tt.args.schema); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAdditionUiSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPairList_Len(t *testing.T) {
	tests := []struct {
		name string
		p    PairList
		want int
	}{
		{
			name: "Test len",
			p:    PairList{{"exampleProperty", 1}, {"exampleProperty1", 2}},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPairList_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		p    PairList
		args args
		want bool
	}{
		{
			name: "Test less",
			p:    PairList{{"exampleProperty", 1}, {"exampleProperty1", 2}},
			args: args{i: 0, j: 1},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPairList_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		p    PairList
		args args
	}{
		{
			name: "Test swap",
			p:    PairList{{"exampleProperty", 1}, {"exampleProperty1", 2}},
			args: args{i: 0, j: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Swap(tt.args.i, tt.args.j)
		})
	}
}

func Test_boolAnnotationToUiSchema(t *testing.T) {
	type args struct {
		annotation string
		prefix     string
		uiSchema   map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test bool annotation to ui schema",
			args: args{
				annotation: "+ui:hidden=false",
				prefix:     UiHidden,
				uiSchema: map[string]interface{}{
					"err:options":    map[string]interface{}{"required": "example err options"},
					"ui:title":       "example tile",
					"ui:description": "example description",
				},
			},
			wantErr: false,
		},
		{
			name: "Test bool annotation to ui schema with error",
			args: args{
				annotation: "+ui:hidden=stringtest",
				prefix:     UiHidden,
				uiSchema: map[string]interface{}{
					"err:options":    map[string]interface{}{"required": "example err options"},
					"ui:title":       "example tile",
					"ui:description": "example description",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := boolAnnotationToUiSchema(tt.args.annotation, tt.args.prefix, tt.args.uiSchema); (err != nil) != tt.wantErr {
				t.Errorf("boolAnnotationToUiSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dealNode(t *testing.T) {
	type args struct {
		schema   *openapi3.Schema
		uiSchema map[string]interface{}
		order    map[string]int
		name     string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test deal node",
			args: args{
				schema: &openapi3.Schema{
					Type: "object",
					Properties: map[string]*openapi3.SchemaRef{
						"exampleProperty": {
							Value: &openapi3.Schema{
								Description: "+title=Example Title\n+description=Example Description\n+minimum=10\n+maximum=10\n+minLength=10\n+maxLength=10\n+pattern=^\\d{3}$",
							},
						},
					},
				},
				uiSchema: map[string]interface{}{
					"exampleProperty": map[string]interface{}{
						"err:options":    map[string]interface{}{"required": "example err options"},
						"ui:title":       "example tile",
						"ui:description": "example description",
						"ui:hidden":      "{{rootFormData.hive.enable == false}}",
					},
					"ui:order": []string{"exampleProperty", "*"},
				},
				order: map[string]int{"exampleProperty": 1},
				name:  "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dealNode(tt.args.schema, tt.args.uiSchema, tt.args.order, tt.args.name)
		})
	}
}

func Test_generateAdditionUiSchema(t *testing.T) {
	type args struct {
		schema      *openapi3.Schema
		name        string
		uiSchema    map[string]interface{}
		parentOrder map[string]int
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "Test generate addition ui schema",
			args: args{
				schema: &openapi3.Schema{
					Type: "object",
					Properties: map[string]*openapi3.SchemaRef{
						"exampleProperty": {
							Value: &openapi3.Schema{
								Description: "+ui:order=1\n+ui:title=example tile\n+ui:description=example description\n+ui:widget=example widget\n+ui:options=example options\n+ui:placeholder=example placeholder\n+ui:description=example description\n+ui:hidden={{rootFormData.hive.enable == false}}\n+err:options={\"required\":\"example err options\"}",
							},
						},
					}},
				name:        "test",
				uiSchema:    nil,
				parentOrder: nil,
			},
			want: map[string]interface{}{
				"exampleProperty": map[string]interface{}{
					"err:options":    map[string]interface{}{"required": "example err options"},
					"ui:title":       "example tile",
					"ui:description": "example description",
					"ui:hidden":      "{{rootFormData.hive.enable == false}}",
				},
				"ui:order": []string{"exampleProperty", "*"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAdditionUiSchema(tt.args.schema, tt.args.name, tt.args.uiSchema, tt.args.parentOrder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateAdditionUiSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonAnnotationToUiSchema(t *testing.T) {
	type args struct {
		annotation string
		prefix     string
		uiSchema   map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test json annotation to ui schema",
			args: args{
				annotation: "+ui:options={\"required\":\"example err options\"}",
				prefix:     UiOptions,
				uiSchema: map[string]interface{}{
					"err:options":    map[string]interface{}{"required": "example err options"},
					"ui:title":       "example tile",
					"ui:description": "example description",
					"ui:hidden":      "{{rootFormData.hive.enable == false}}",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := jsonAnnotationToUiSchema(tt.args.annotation, tt.args.prefix, tt.args.uiSchema); (err != nil) != tt.wantErr {
				t.Errorf("jsonAnnotationToUiSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_orderToList(t *testing.T) {
	type args struct {
		order map[string]int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test order to list",
			args: args{order: map[string]int{"exampleProperty1": 1, "exampleProperty2": 2}},
			want: []string{"exampleProperty1", "exampleProperty2", "*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orderToList(tt.args.order); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderToList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringAnnotationToUiSchema(t *testing.T) {
	type args struct {
		annotation string
		prefix     string
		uiSchema   map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test string annotation to ui schema",
			args: args{
				annotation: "+ui:title=example tile",
				prefix:     UiTitle,
				uiSchema: map[string]interface{}{
					"err:options":    map[string]interface{}{"required": "example err options"},
					"ui:description": "example description",
					"ui:hidden":      "{{rootFormData.hive.enable == false}}",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stringAnnotationToUiSchema(tt.args.annotation, tt.args.prefix, tt.args.uiSchema)
		})
	}
}
