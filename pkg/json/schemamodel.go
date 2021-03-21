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
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type schemaProperty struct {
	key         string
	reference   string
	object      *objectSchema
	array       *arraySchema
	oneOf       []schemaProperty
	allOf       []schemaProperty
	typ         string
	format      string
	enum        []interface{}
	required    bool
	pattern     string
	description string
}

type arraySchema struct {
	items schemaProperty
}

type objectSchema struct {
	properties map[string]schemaProperty
}

func (a arraySchema) itr(entityId string, name []string, out *ls.Attribute) {
	elem := ls.NewAttribute(nil)
	schemaAttrs(entityId, append(name, "*"), a.items, elem)
	out.MakeArray(elem)
}

func (obj objectSchema) itr(entityId string, name []string, out *ls.Attributes) {
	for k, v := range obj.properties {
		attr := ls.NewAttribute(nil)
		nm := append(name, k)
		schemaAttrs(entityId, nm, v, attr)
		attr.ID = entityId + "." + strings.Join(nm, ".")
		out.Add(attr)
	}
}

func schemaAttrs(entityId string, name []string, attr schemaProperty, out *ls.Attribute) {
	if len(attr.format) > 0 {
		out.Values[ls.AttributeAnnotations.Format.ID] = []interface{}{map[string]interface{}{"@value": attr.format}}
	}
	if len(attr.typ) > 0 {
		out.Values[ls.AttributeAnnotations.Type.ID] = []interface{}{map[string]interface{}{"@value": attr.typ}}
	}
	if len(attr.key) > 0 {
		out.Values[ls.AttributeAnnotations.Name.ID] = []interface{}{map[string]interface{}{"@value": attr.key}}
	}
	if len(attr.enum) > 0 {
		elements := make([]interface{}, 0, len(attr.enum))
		for _, v := range attr.enum {
			elements = append(elements, map[string]interface{}{"@value": v})
		}
		out.Values[ls.AttributeAnnotations.Enumeration.ID] = []interface{}{map[string]interface{}{"@list": elements}}
	}
	if len(attr.pattern) > 0 {
		out.Values[ls.AttributeAnnotations.Pattern.ID] = []interface{}{map[string]interface{}{"@value": attr.pattern}}
	}
	if len(attr.description) > 0 {
		out.Values[ls.AttributeAnnotations.Information.ID] = []interface{}{map[string]interface{}{"@value": attr.description}}
	}
	if attr.required {
		out.Values[ls.AttributeAnnotations.Required.ID] = []interface{}{map[string]interface{}{"@value": true}}
	}
	if len(attr.reference) > 0 {
		out.MakeReference(attr.reference)
		return
	}
	if attr.object != nil {
		attrs := ls.NewAttributes(nil)
		attr.object.itr(entityId, name, attrs)
		out.MakeObject(attrs)
		return
	}
	if attr.array != nil {
		attr.array.itr(entityId, name, out)
		return
	}
	buildChoices := func(arr []schemaProperty) []*ls.Attribute {
		elements := make([]*ls.Attribute, 0, len(arr))
		for i, x := range arr {
			out := ls.NewAttribute(nil)
			newName := append(name, fmt.Sprint(i))
			schemaAttrs(entityId, newName, x, out)
			if out.ID == "" {
				out.ID = entityId + "." + strings.Join(newName, ".")
			}
			elements = append(elements, out)
		}
		return elements
	}
	if len(attr.oneOf) > 0 {
		out.MakePolymorphic(buildChoices(attr.oneOf))
		return
	}
	if len(attr.allOf) > 0 {
		out.MakeComposition(buildChoices(attr.allOf))
	}
}
