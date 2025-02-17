package validators

import (
	"github.com/santhosh-tekuri/jsonschema/v3"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// JsonFormatTerm validates if the value matches one of the json format implementations
var JsonFormatTerm = ls.NewTerm(ls.LS, "validation/json/format", false, false, ls.OverrideComposition, struct {
	JsonFormatValidator
}{
	JsonFormatValidator{},
})

// JsonFormatValidator checks if the input value matches a given format
type JsonFormatValidator struct{}

const compiledJsonFormatTerm = "$compiledJsonFormat"

// ValidateValue checks if the value matches the format
func (validator JsonFormatValidator) validateValue(value, format string) error {
	f := jsonschema.Formats[format]
	if f == nil {
		return ls.ErrValidation{Validator: JsonFormatTerm, Msg: "Unknown format: " + format}
	}
	if !f(value) {
		return ls.ErrValidation{Validator: JsonFormatTerm, Msg: "Invalid value for " + format}
	}
	return nil
}

// ValidateValue checks if the value matches the format
func (validator JsonFormatValidator) ValidateValue(value *string, schemaNode graph.Node) error {
	if value == nil {
		return nil
	}
	c, _ := schemaNode.GetProperty(compiledJsonFormatTerm)
	return validator.validateValue(*value, c.(string))
}

// ValidateNode validates the node value if it is non-nil
func (validator JsonFormatValidator) ValidateNode(docNode, schemaNode graph.Node) error {
	if docNode == nil {
		return nil
	}
	value, ok := ls.GetRawNodeValue(docNode)
	if !ok {
		return nil
	}
	return validator.ValidateValue(&value, schemaNode)
}

func (validator JsonFormatValidator) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if !value.IsString() {
		return ls.ErrValidatorCompile{Validator: JsonFormatTerm, Msg: "Invalid format value"}
	}
	if jsonschema.Formats[value.AsString()] == nil {
		return ls.ErrValidatorCompile{Validator: JsonFormatTerm, Msg: "Invalid format value"}
	}
	target.SetProperty(compiledJsonFormatTerm, value.AsString())
	return nil
}
