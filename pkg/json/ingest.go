// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package json

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// ErrValidation is returned for validation errors
type ErrValidation struct {
	NodeID  string
	Message string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("Validation error at '%s': %s", e.NodeID, e.Message)
}

// ErrNotAnObject is returned if input is not a JSON object
var ErrNotAnObject = errors.New("Not a JSON Object")

// Node is the common interface for ingested data nodes. It
// contains the attribute ID and name, and provides a way to convert the
// underlying data to a map
type Node interface {
	GetID() string
	ToMap() interface{}
	GetAttributeName() string
	SetAttributeName(string)
}

// NullNode is a null JSON node
type NullNode struct {
	ID            string
	AttributeName string
	SchemaNode    *ls.Attribute
}

func (n NullNode) GetID() string { return n.ID }

func (n NullNode) GetAttributeName() string { return n.AttributeName }

func (n *NullNode) SetAttributeName(name string) { n.AttributeName = name }

func (n NullNode) ToMap() interface{} {
	ret := map[string]interface{}{ls.DocTerms.Value.ID: []interface{}{map[string]interface{}{"@value": nil}}}
	embedSchema(ret, n.SchemaNode, n.AttributeName)
	return []interface{}{ret}
}

// ValueNode is a JSON node with number, string, or boolean value
type ValueNode struct {
	// Node ID
	ID            string
	AttributeName string
	// SchemaNode is the associated schema node
	SchemaNode *ls.Attribute
	// Value of the node
	Value interface{}
}

func (v ValueNode) GetID() string { return v.ID }

func (v ValueNode) GetAttributeName() string { return v.AttributeName }

func (v *ValueNode) SetAttributeName(name string) { v.AttributeName = name }

func (v ValueNode) ToMap() interface{} {
	ret := map[string]interface{}{ls.DocTerms.Value.ID: []interface{}{map[string]interface{}{"@value": v.Value}}}
	embedSchema(ret, v.SchemaNode, v.AttributeName)
	return []interface{}{ret}
}

// ObjectNode is a JSON object node
type ObjectNode struct {
	// Node ID
	ID            string
	AttributeName string
	// Schema node associated with this node
	SchemaNode *ls.Attribute
	// Children nodes
	Children []Node
}

func (o ObjectNode) GetID() string { return o.ID }

func (o ObjectNode) GetAttributeName() string { return o.AttributeName }

func (o *ObjectNode) SetAttributeName(name string) { o.AttributeName = name }

func (o ObjectNode) ToMap() interface{} {
	children := make(map[string]interface{})
	for _, ch := range o.Children {
		children[ch.GetID()] = ch.ToMap()
	}
	ret := map[string]interface{}{
		ls.DocTerms.Attributes.ID: []interface{}{children},
	}
	embedSchema(ret, o.SchemaNode, o.AttributeName)
	return []interface{}{ret}
}

type ArrayNode struct {
	// NodeID
	ID            string
	AttributeName string
	SchemaNode    *ls.Attribute
	Elements      []Node
}

func (a ArrayNode) GetID() string { return a.ID }

func (a ArrayNode) GetAttributeName() string { return a.AttributeName }

func (a *ArrayNode) SetAttributeName(name string) { a.AttributeName = name }

func (a ArrayNode) ToMap() interface{} {
	el := make([]interface{}, 0, len(a.Elements))
	for _, x := range a.Elements {
		el = append(el, x.ToMap())
	}
	ret := map[string]interface{}{
		ls.DocTerms.ArrayElements.ID: []interface{}{map[string]interface{}{"@list": el}},
	}
	embedSchema(ret, a.SchemaNode, a.AttributeName)
	return []interface{}{ret}
}

func validate(docValue interface{}, schemaTerms map[string]interface{}) error {
	for k, termValue := range schemaTerms {
		if len(k) > 0 && k[0] != '@' {
			termMeta := ls.Terms[k]
			if termMeta.Validate != nil {
				if err := termMeta.Validate(termValue, docValue); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func appendID(base, prefix string) string {
	if len(base) == 0 {
		return prefix
	}
	return base + "." + prefix
}

// Ingest a json document using the schema. The output will have all
// input nodes associated with schema nodes.
//
// BaseID is the ID of the root object. All other attribute names are
// generated by appending the attribute path to baseID
func Ingest(baseID string, input interface{}, schema *ls.SchemaLayer) (Node, error) {
	if schema == nil {
		return merge(baseID, input, nil)
	}
	if input == nil {
		return nil, nil
	}
	m, ok := input.(map[string]interface{})
	if !ok {
		return nil, ErrNotAnObject
	}
	output := ObjectNode{ID: baseID}
	err := mergeAttributes(baseID, m, &schema.Attributes, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

// merge schema information into the input document and return a tree of *ObjectNode, *ValueNode, *ArrayNode, *NullNode objects
func merge(baseID string, input interface{}, schema *ls.Attribute) (Node, error) {
	if input == nil {
		return &NullNode{ID: baseID, SchemaNode: schema}, nil
	}
	if m, ok := input.(map[string]interface{}); ok {
		return mergeObject(baseID, m, schema)
	}
	if a, ok := input.([]interface{}); ok {
		return mergeArray(baseID, a, schema)
	}
	return mergeValue(baseID, input, schema)
}

func mergeAttributes(baseID string, input map[string]interface{}, attributes *ls.Attributes, output *ObjectNode) error {
	schemaNodes := make(map[string]*ls.Attribute)
	for i := 0; i < attributes.Len(); i++ {
		attribute := attributes.Get(i)
		name := ls.GetStringValue("@value", attribute.Values[ls.AttributeAnnotations.Name.ID])
		if len(name) == 0 {
			name = attribute.ID
		}
		if len(name) > 0 {
			schemaNodes[name] = attribute
		}
	}
	// Validate schema elements that are not in the document to catch possible required elements
	for k, v := range schemaNodes {
		if _, ok := input[k]; !ok {
			if err := validate(nil, v.Values); err != nil {
				return err
			}
		}
	}
	output.Children = make([]Node, 0, len(input))
	for k, v := range input {
		nodeId := appendID(baseID, k)
		schemaNode := schemaNodes[k]
		child, err := merge(nodeId, v, schemaNode)
		if err != nil {
			return err
		}
		child.SetAttributeName(k)
		output.Children = append(output.Children, child)
	}
	return nil
}

func mergeObject(baseID string, input map[string]interface{}, schema *ls.Attribute) (Node, error) {
	if schema == nil {
		ret := ObjectNode{ID: baseID}
		for k, v := range input {
			newID := appendID(baseID, k)
			child, err := merge(newID, v, nil)
			if err != nil {
				return nil, err
			}
			child.SetAttributeName(k)
			ret.Children = append(ret.Children, child)
		}
		return &ret, nil
	}
	if schema.IsArray() {
		return nil, ErrValidation{NodeID: baseID, Message: "Schema expects array, but document has object"}
	}
	if schema.IsValue() {
		return nil, ErrValidation{NodeID: baseID, Message: "Schema expects value, but document has object"}
	}
	if attributes := schema.GetAttributes(); attributes != nil {
		if err := validate(input, schema.Values); err != nil {
			return nil, err
		}
		ret := ObjectNode{ID: baseID, SchemaNode: schema}
		if err := mergeAttributes(baseID, input, attributes, &ret); err != nil {
			return nil, err
		}
		return &ret, nil
	}

	// OneOf
	var validOption interface{}
	var validRet Node
	for _, option := range schema.GetPolymorphicOptions() {
		if v, err := mergeObject(baseID, input, option); err == nil {
			if validOption != nil {
				return nil, ErrValidation{NodeID: baseID, Message: "Multiple options match"}
			}
			validOption = option
			validRet = v
		}
	}
	if validOption == nil {
		return nil, ErrValidation{NodeID: baseID, Message: "No options match"}
	}
	return validRet, nil
}

func mergeArray(baseID string, input []interface{}, schema *ls.Attribute) (Node, error) {
	if schema == nil {
		ret := ArrayNode{ID: baseID, Elements: make([]Node, 0, len(input))}
		for i := range input {
			newID := appendID(baseID, strconv.Itoa(i))
			el, err := merge(newID, input[i], nil)
			if err != nil {
				return nil, err
			}
			ret.Elements = append(ret.Elements, el)
		}
		return &ret, nil
	}

	if schema.IsObject() {
		return nil, ErrValidation{NodeID: baseID, Message: "Schema expects object, but document has array"}
	}

	if schema.IsValue() {
		return nil, ErrValidation{NodeID: baseID, Message: "Schema expects value, but document has array"}
	}
	if items := schema.GetArrayItems(); items != nil {
		if err := validate(input, schema.Values); err != nil {
			return nil, err
		}
		ret := ArrayNode{ID: baseID, SchemaNode: schema, Elements: make([]Node, 0, len(input))}
		for i := range input {
			nodeId := appendID(baseID, strconv.Itoa(i))
			child, err := merge(nodeId, input[i], items)
			if err != nil {
				return nil, err
			}
			ret.Elements = append(ret.Elements, child)
		}
		return &ret, nil
	}

	// OneOf
	var validOption interface{}
	var validRet Node
	for _, option := range schema.GetPolymorphicOptions() {
		if v, err := mergeArray(baseID, input, option); err == nil {
			if validOption != nil {
				return nil, ErrValidation{NodeID: baseID, Message: "Multiple options match"}
			}
			validOption = option
			validRet = v
		}
	}
	if validOption == nil {
		return nil, ErrValidation{NodeID: baseID, Message: "No options match"}
	}
	return validRet, nil
}

func mergeValue(baseID string, input interface{}, schema *ls.Attribute) (Node, error) {
	if schema == nil {
		if input == nil {
			return &NullNode{ID: baseID}, nil
		}
		return &ValueNode{ID: baseID, Value: input}, nil
	}

	if schema.IsObject() {
		return nil, ErrValidation{NodeID: baseID, Message: "Schema expects object but document has value"}
	}
	if schema.IsArray() {
		return nil, ErrValidation{NodeID: baseID, Message: "Schema expects array, but document has value"}
	}
	if err := validate(input, schema.Values); err != nil {
		return nil, err
	}
	if !schema.IsPolymorphic() {
		if input == nil {
			return &NullNode{ID: baseID, SchemaNode: schema}, nil
		}
		return &ValueNode{ID: baseID, SchemaNode: schema, Value: input}, nil
	}
	// A value wrt a oneOf schema. One must validate
	var validOption interface{}
	var validRet Node
	for _, option := range schema.GetPolymorphicOptions() {
		if v, err := mergeValue(baseID, input, option); err == nil {
			if validOption != nil {
				return nil, ErrValidation{NodeID: baseID, Message: "Multiple options match"}
			}
			validOption = option
			validRet = v
		}
	}
	if validOption == nil {
		return nil, ErrValidation{NodeID: baseID, Message: "No options match"}
	}
	return validRet, nil
}

func embedSchema(target map[string]interface{}, schema *ls.Attribute, attributeName string) {
	if schema == nil {
		if len(attributeName) > 0 {
			target[ls.AttributeAnnotations.Name.ID] = []interface{}{map[string]interface{}{"@value": attributeName}}
		}
		return
	}
	if len(schema.ID) > 0 {
		target[ls.DocTerms.SchemaAttributeID.ID] = []interface{}{map[string]interface{}{"@id": schema.ID}}
	}
	for k, v := range schema.Values {
		target[k] = v
	}
}
