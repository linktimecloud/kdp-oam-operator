package definition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/reference/pkg/types"
	"reflect"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/encoding/gocode/gocodec"
	"cuelang.org/go/tools/fix"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Definition the general struct for handling all kinds of definitions like ComponentDefinition or TraitDefinition
type Definition struct {
	unstructured.Unstructured
}

// the names for different type of definition
const (
	applicationDefType = "application"
	xdefinitionDefType = "xdefinition"
)

var (
	// DefinitionTemplateKeys the keys for accessing definition template
	DefinitionTemplateKeys = []string{"spec", "schematic", "cue", "template"}
	DefinitionTypeToKind   = map[string]string{
		xdefinitionDefType: reflect.TypeOf(v1alpha1.XDefinition{}).Name(),
		applicationDefType: reflect.TypeOf(v1alpha1.Application{}).Name(),
	}
)

const (
	// DescriptionKey the key for accessing definition description
	DescriptionKey = "bdc.kdp.io/description"
	// AliasKey the key for accessing definition alias
	AliasKey = "bdc.kdp.io/alias"
	// UserPrefix defines the prefix of user customized label or annotation
	UserPrefix = "custom.bdc.kdp.io/"
)

// FromCUEString converts cue string into Definition
func (def *Definition) FromCUEString(cueString string) error {
	cuectx := cuecontext.New()
	f, err := parser.ParseFile("-", cueString, parser.ParseComments)
	if err != nil {
		return err
	}
	n := fix.File(f)
	var importDecls, metadataDecls, templateDecls []ast.Decl
	for _, decl := range n.Decls {
		if importDecl, ok := decl.(*ast.ImportDecl); ok {
			importDecls = append(importDecls, importDecl)
		} else if field, ok := decl.(*ast.Field); ok {
			label := ""
			switch l := field.Label.(type) {
			case *ast.Ident:
				label = l.Name
			case *ast.BasicLit:
				label = l.Value
			}
			if label == "" {
				return errors.Errorf("found unexpected decl when parsing cue: %v", label)
			}
			if label == "template" {
				if v, ok := field.Value.(*ast.StructLit); ok {
					templateDecls = append(templateDecls, v.Elts...)
				} else {
					return errors.Errorf("unexpected decl found in template: %v", decl)
				}
			} else {
				metadataDecls = append(metadataDecls, field)
			}
		}
	}
	if len(metadataDecls) == 0 {
		return errors.Errorf("no metadata found, invalid")
	}
	if len(templateDecls) == 0 {
		return errors.Errorf("no template found, invalid")
	}
	var importString, metadataString, templateString string
	if importString, err = encodeDeclsToString(importDecls); err != nil {
		return errors.Wrapf(err, "failed to encode import decls to string")
	}
	if metadataString, err = encodeDeclsToString(metadataDecls); err != nil {
		return errors.Wrapf(err, "failed to encode metadata decls to string")
	}
	// notice that current template decls are concatenated without any blank lines which might be inconsistent with original cue file, but it would not affect the syntax
	if templateString, err = encodeDeclsToString(templateDecls); err != nil {
		return errors.Wrapf(err, "failed to encode template decls to string")
	}

	inst := cuectx.CompileString(metadataString)
	if inst.Err() != nil {
		return inst.Err()
	}
	templateString, err = formatCUEString(importString + templateString)
	if err != nil {
		return err
	}

	return def.FromCUE(&inst, templateString)
}

// FromCUE converts CUE value (predefined Definition's cue format) to Definition
// nolint:gocyclo,staticcheck
func (def *Definition) FromCUE(val *cue.Value, templateString string) error {
	if def.Object == nil {
		def.Object = map[string]interface{}{}
	}
	annotations := map[string]string{}
	for k, v := range def.GetAnnotations() {
		if !strings.HasPrefix(k, UserPrefix) && k != DescriptionKey {
			annotations[k] = v
		}
	}
	labels := map[string]string{}
	for k, v := range def.GetLabels() {
		if !strings.HasPrefix(k, UserPrefix) {
			labels[k] = v
		}
	}
	spec, ok := def.Object["spec"].(map[string]interface{})
	if !ok {
		spec = map[string]interface{}{}
	}
	codec := gocodec.New(&cue.Runtime{}, &gocodec.Config{})
	nameFlag := false
	fields, err := val.Fields()
	if err != nil {
		return err
	}
	for fields.Next() {
		definitionName := fields.Label()
		v := fields.Value()
		if nameFlag {
			return fmt.Errorf("duplicated definition name found, %s and %s", def.GetName(), definitionName)
		}
		nameFlag = true
		def.SetName(definitionName)
		_fields, err := v.Fields()
		if err != nil {
			return err
		}
		for _fields.Next() {
			_key := _fields.Label()
			_value := _fields.Value()
			switch _key {
			case "type":
				_type, err := _value.String()
				if err != nil {
					return err
				}
				if err = def.SetType(_type); err != nil {
					return err
				}
			case "alias":
				alias, err := _value.String()
				if err != nil {
					return err
				}
				annotations[AliasKey] = alias
			case "description":
				desc, err := _value.String()
				if err != nil {
					return err
				}
				annotations[DescriptionKey] = desc
			case "annotations":
				var _annotations map[string]string
				if err := codec.Encode(_value, &_annotations); err != nil {
					return err
				}
				for _k, _v := range _annotations {
					annotations[_k] = _v
				}
			case "labels":
				var _labels map[string]string
				if err := codec.Encode(_value, &_labels); err != nil {
					return err
				}
				for _k, _v := range _labels {
					if strings.Contains(_k, "bdc.dev") {
						labels[_k] = _v
					} else {
						labels[UserPrefix+_k] = _v
					}
				}
			case "attributes":
				if err := codec.Encode(_value, &spec); err != nil {
					return err
				}
			}
		}
	}
	def.SetAnnotations(annotations)
	def.SetLabels(labels)
	if err := unstructured.SetNestedField(spec, templateString, DefinitionTemplateKeys[1:]...); err != nil {
		return err
	}
	if err = validateSpec(spec, def.GetType()); err != nil {
		return fmt.Errorf("invalid definition spec: %w", err)
	}
	def.Object["spec"] = spec

	// set default apiResource
	if spec["apiResource"] == nil {
		spec["apiResource"] = apiResource(def.GetName())
	}

	definitionMap := spec["apiResource"].(map[string]interface{})["definition"]
	if definitionMap != nil {
		relatedResourceKind := definitionMap.(map[string]interface{})["kind"]
		if relatedResourceKind != nil {
			prefix := types.ApiResourceTypePrefix[relatedResourceKind.(string)]
			if prefix != "" && !strings.Contains(def.GetName(), prefix) {
				def.SetName(prefix + def.GetName())
			}
		}
	}

	return nil
}

func validateSpec(spec map[string]interface{}, t string) error {
	bs, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	var tpl interface{}
	switch t {
	case applicationDefType:
		tpl = &v1alpha1.ApplicationSpec{}
	case xdefinitionDefType:
		tpl = &v1alpha1.XDefinitionSpec{}
	default:
	}
	if tpl != nil {
		return StrictUnmarshal(bs, tpl)
	}
	return nil
}

func encodeDeclsToString(decls []ast.Decl) (string, error) {
	bs, err := format.Node(&ast.File{Decls: decls}, format.Simplify())
	if err != nil {
		return "", fmt.Errorf("failed to encode cue: %w", err)
	}
	return strings.TrimSpace(string(bs)) + "\n", nil
}

func formatCUEString(cueString string) (string, error) {
	f, err := parser.ParseFile("-", cueString, parser.ParseComments)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse file during format cue string")
	}
	n := fix.File(f)
	b, err := format.Node(n, format.Simplify())
	if err != nil {
		return "", errors.Wrapf(err, "failed to format node during formating cue string")
	}
	return string(b), nil
}

// SetGVK set the GroupVersionKind of Definition
func (def *Definition) SetGVK(kind string) {
	def.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "bdc.kdp.io",
		Version: "v1alpha1",
		Kind:    kind,
	})
}

// GetType gets the type of Definition
func (def *Definition) GetType() string {
	kind := def.GetKind()
	for k, v := range DefinitionTypeToKind {
		if v == kind {
			return k
		}
	}
	return "Application"
}

// SetType sets the type of Definition
func (def *Definition) SetType(t string) error {
	kind, ok := DefinitionTypeToKind[t]
	if !ok {
		return fmt.Errorf("invalid type %s", t)
	}
	def.SetGVK(kind)
	return nil
}

// StrictUnmarshal unmarshal target structure and disallow unknown fields
func StrictUnmarshal(bs []byte, dest interface{}) error {
	d := json.NewDecoder(bytes.NewReader(bs))
	d.DisallowUnknownFields()
	return d.Decode(dest)
}

func apiResource(typeName string) map[string]interface{} {
	resource := map[string]interface{}{
		"definition": map[string]interface{}{
			"apiVersion": "bdc.kdp.io/v1alpha1",
			"kind":       "Application",
			"type":       typeName,
		},
	}
	return resource
}
