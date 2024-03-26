package deftemplate

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang-collections/collections/queue"
	"k8s.io/klog"
	"sort"
	"strconv"
	"strings"
)

const (
	// OpenApiTitle is the openapi schema title
	OpenApiTitle = "+title="
	// OpenApiDescription is the openapi schema description
	OpenApiDescription = "+description="
	// OpenApiMinimum is the openapi schema minimum
	OpenApiMinimum = "+minimum="
	// OpenApiMaximum is the openapi schema maximum
	OpenApiMaximum = "+maximum="
	// OpenApiMinLength is the openapi schema minLength
	OpenApiMinLength = "+minLength="
	// OpenApiMaxLength is the openapi schema maxLength
	OpenApiMaxLength = "+maxLength="
	// OpenApiPattern is the openapi schema pattern
	OpenApiPattern = "+pattern="

	// UiOrder is the ui schema order for sort fields
	UiOrder = "+ui:order="
	// UiDescription is the ui schema description
	UiDescription = "+ui:description="
	// UiTitle is the ui schema title
	UiTitle = "+ui:title="
	// UiHidden defines whether the field should be hidden
	UiHidden = "+ui:hidden="
	// UiOptions is the ui schema options
	UiOptions = "+ui:options="
	// ErrOptions is the error schema options, put it into ui schema now
	ErrOptions = "+err:options="
)

var (
	// UiSchemaAnnotationToKey is the map of ui schema annotation to key
	UiSchemaAnnotationToKey = map[string]string{
		UiOrder:       "ui:order",
		UiDescription: "ui:description",
		UiTitle:       "ui:title",
		UiHidden:      "ui:hidden",
		UiOptions:     "ui:options",
		ErrOptions:    "err:options",
	}
)

// FixOpenAPISchema replaces openapi3.Schema "description" and "title" fields with user defined annotation
func FixOpenAPISchema(schema *openapi3.Schema) {
	t := schema.Type
	switch t {
	case "object":
		for _, v := range schema.Properties {
			s := v.Value
			FixOpenAPISchema(s)
		}
	case "array":
		if schema.Items != nil {
			FixOpenAPISchema(schema.Items.Value)
		}
	}

	if schema.Description == "" {
		return
	}

	tagList := strings.Split(schema.Description, "\n")
	// set default description, for annotation may not contain description
	schema.Description = ""
	// these tags are added by the XDefinition parameter fields annotation, like: // +title=example title
	// only care about annotations start with "+title=" or  "+description=" which are the standard fields in openapi3.Schema
	for _, tag := range tagList {
		if strings.Contains(tag, OpenApiDescription) {
			schema.Description = strings.TrimSpace(strings.Split(tag, OpenApiDescription)[1])
		} else if strings.Contains(tag, OpenApiTitle) {
			title := strings.TrimSpace(strings.Split(tag, OpenApiTitle)[1])
			if title != "" {
				schema.Title = title
			}
		} else if strings.Contains(tag, OpenApiMinimum) {
			minum := strings.TrimSpace(strings.Split(tag, OpenApiMinimum)[1])
			float, err := strconv.ParseFloat(minum, 64)
			if err != nil {
				klog.Error(err)
				continue
			}

			schema.Min = &float
		} else if strings.Contains(tag, OpenApiMaximum) {
			maximum := strings.TrimSpace(strings.Split(tag, OpenApiMaximum)[1])
			float, err := strconv.ParseFloat(maximum, 64)
			if err != nil {
				klog.Error(err)
				continue
			}

			schema.Max = &float
		} else if strings.Contains(tag, OpenApiMinLength) {
			minLength := strings.TrimSpace(strings.Split(tag, OpenApiMinLength)[1])
			integer, err := strconv.ParseUint(minLength, 10, 64)
			if err != nil {
				klog.Error(err)
				continue
			}

			schema.MinLength = integer
		} else if strings.Contains(tag, OpenApiMaxLength) {
			maxLength := strings.TrimSpace(strings.Split(tag, OpenApiMaxLength)[1])
			integer, err := strconv.ParseUint(maxLength, 10, 64)
			if err != nil {
				klog.Error(err)
				continue
			}

			schema.MaxLength = &integer
		} else if strings.Contains(tag, OpenApiPattern) {
			pattern := strings.TrimSpace(strings.Split(tag, OpenApiPattern)[1])
			if pattern != "" {
				schema.Pattern = pattern
			}
		}

	}
}

func GetAdditionUiSchema(schema *openapi3.Schema) []byte {
	uiSchema := generateAdditionUiSchema(schema, "", nil, nil)

	marshal, err := json.Marshal(uiSchema)
	if err != nil {
		return []byte("{}")
	}

	return marshal
}

// generateAdditionUiSchema generate ui schema form user defined annotations, it should process before FixOpenAPISchema
// for FixOpenAPISchema will replace schema.description with user defined annotation
func generateAdditionUiSchema(schema *openapi3.Schema, name string, uiSchema map[string]interface{}, parentOrder map[string]int) map[string]interface{} {
	if uiSchema == nil {
		uiSchema = make(map[string]interface{})
	}
	queue := queue.New()
	if parentOrder == nil {
		parentOrder = make(map[string]int)
	}
	order := make(map[string]int)

	// BFS traverse the schema tree by branch and level

	// prepare
	if schema.Type == "object" {
		for k, v := range schema.Properties {
			uiSchema[k] = make(map[string]interface{})
			s := v.Value
			queue.Enqueue([]interface{}{k, s})
		}
	}
	// node handle
	dealNode(schema, uiSchema, parentOrder, name)

	// traverse
	for queue.Len() > 0 {
		node := queue.Dequeue().([]interface{})
		generateAdditionUiSchema(node[1].(*openapi3.Schema), node[0].(string), uiSchema[node[0].(string)].(map[string]interface{}), order)
	}

	if len(order) > 0 {
		uiSchema[UiSchemaAnnotationToKey[UiOrder]] = orderToList(order)
	}

	return uiSchema
}

func dealNode(schema *openapi3.Schema, uiSchema map[string]interface{}, order map[string]int, name string) {
	if schema.Description == "" {
		return
	}

	// annotations must be added in line style
	annotationList := strings.Split(schema.Description, "\n")

	for _, annotation := range annotationList {
		if strings.Contains(annotation, UiOrder) {
			orderStr := strings.TrimSpace(strings.Split(annotation, UiOrder)[1])
			orderInt, err := strconv.Atoi(orderStr)
			if err != nil {
				klog.Error(err)
			}
			order[name] = orderInt
		} else if strings.Contains(annotation, UiTitle) {
			stringAnnotationToUiSchema(annotation, UiTitle, uiSchema)
		} else if strings.Contains(annotation, UiDescription) {
			stringAnnotationToUiSchema(annotation, UiDescription, uiSchema)
		} else if strings.Contains(annotation, UiHidden) {
			err := boolAnnotationToUiSchema(annotation, UiHidden, uiSchema)
			if err == nil {
				continue
			}
			klog.Infof("boolAnnotationToUiSchema error: %v. Setting uiSchema hidden field with raw annotation string", err)
			hiddenStr := strings.TrimSpace(strings.Split(annotation, UiHidden)[1])
			if strings.HasSuffix(hiddenStr, "}}") && strings.HasPrefix(hiddenStr, "{{") {
				uiSchema[UiSchemaAnnotationToKey[UiHidden]] = hiddenStr
			}
		} else if strings.Contains(annotation, UiOptions) {
			jsonAnnotationToUiSchema(annotation, UiOptions, uiSchema)
		} else if strings.Contains(annotation, ErrOptions) {
			jsonAnnotationToUiSchema(annotation, ErrOptions, uiSchema)
		}
	}
}

func stringAnnotationToUiSchema(annotation string, prefix string, uiSchema map[string]interface{}) {
	uiSchema[UiSchemaAnnotationToKey[prefix]] = strings.TrimSpace(strings.Split(annotation, prefix)[1])
}

func boolAnnotationToUiSchema(annotation string, prefix string, uiSchema map[string]interface{}) error {
	value := strings.TrimSpace(strings.Split(annotation, prefix)[1])
	valueBool, err := strconv.ParseBool(value)
	// bool type conversion error, ignore this annotation and go on
	if err != nil {
		return err
	}
	uiSchema[UiSchemaAnnotationToKey[prefix]] = valueBool

	return nil
}

func jsonAnnotationToUiSchema(annotation string, prefix string, uiSchema map[string]interface{}) error {
	var uiOptionMap map[string]interface{}
	optionJson := strings.TrimSpace(strings.Split(annotation, prefix)[1])
	err := json.Unmarshal([]byte(optionJson), &uiOptionMap)
	// json type conversion error, ignore this annotation and go on
	if err != nil {
		klog.Error(err)
		return err
	}

	uiSchema[UiSchemaAnnotationToKey[prefix]] = uiOptionMap

	return nil
}

// orderToList convert order map to list
func orderToList(order map[string]int) []string {
	var result []string

	// 将map的key和value存储到PairList中
	pairs := make(PairList, 0, len(order))
	for k, v := range order {
		pairs = append(pairs, Pair{k, v})
	}

	// 对PairList进行排序
	sort.Sort(pairs)
	for _, pair := range pairs {
		result = append(result, pair.Key)
	}

	result = append(result, "*")
	return result
}

type Pair struct {
	Key   string
	Value int
}
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
