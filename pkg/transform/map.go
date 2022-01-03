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

package transform

import (
	"strings"

	"github.com/bserdar/digraph"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type MappingError struct {
	DocPath []ls.Node
	Msg     string
}

func (m MappingError) Error() string {
	path := make([]string, 0, len(m.DocPath))
	for _, x := range m.DocPath {
		path = append(path, x.GetID())
	}
	return strings.Join(path, ".") + ":" + m.Msg
}

// GraphMapper contains an ingester to control how data will be
// ingested
type GraphMapper struct {
	ls.Ingester
}

// Map converts the given document root to m.Schema
func (m GraphMapper) Map(docRoot ls.Node) (result ls.Node, err error) {
	return m.mapGraph(docRoot, m.Schema.GetSchemaRootNode(), []interface{}{m.Schema.GetID()})
}

func (m GraphMapper) getName(schemaNode ls.Node) string {
	s := schemaNode.GetProperties()[ls.AttributeNameTerm]
	return s.AsString()
}

// mapGraph converts the document at `sourceDocRoot` to `targetSchemaRoot` using
// `map` properties of the source doc nodes.
func (m GraphMapper) mapGraph(sourceDocRoot, targetSchemaRoot ls.Node, path []interface{}) (result ls.Node, err error) {
	schemaTypes := targetSchemaRoot.GetTypes()
	switch {
	case schemaTypes.Has(ls.AttributeTypes.Value):
		return m.mapValue(sourceDocRoot, targetSchemaRoot, path)

	case schemaTypes.Has(ls.AttributeTypes.Object):
		newNode := m.NewNode(path, targetSchemaRoot)
		return m.mapObject(sourceDocRoot, targetSchemaRoot, newNode, path)

	case schemaTypes.Has(ls.AttributeTypes.Array):
	case schemaTypes.Has(ls.AttributeTypes.Polymorphic):
		panic("Unsupported polymorphic mapping")
	}
	panic("Unsupported mapping")
}

func (m GraphMapper) findUnder(sourceDocRoot ls.Node, attrID string) ls.Node {
	var found ls.Node
	ls.IterateDescendants(sourceDocRoot, func(node ls.Node, path []ls.Node) bool {
		m := ls.GetSchemaProperty(node, ls.MapTerm)
		if m == nil || !m.IsString() || m.AsString() != attrID {
			return true
		}
		found = node
		return false
	}, func(edge ls.Edge, path []ls.Node) ls.EdgeFuncResult {
		if edge.GetLabel() == ls.HasTerm {
			return ls.FollowEdgeResult
		}
		return ls.SkipEdgeResult
	}, true)
	return found
}

func (m GraphMapper) mapValue(sourceDocNode, targetSchemaNode ls.Node, path []interface{}) (ls.Node, error) {
	newNode := m.NewNode(path, targetSchemaNode)
	value := ls.GetNodeFilteredValue(sourceDocNode)
	newNode.SetValue(value)
	return newNode, nil
}

func (m GraphMapper) mapObject(sourceDocRoot, targetSchemaRoot ls.Node, newNode ls.Node, path []interface{}) (ls.Node, error) {
	// sourceDocRoot will be used as the root for all searches
	// We will recurse into targetSchemaRoot
	// newNode is the node corresponding to the instance of targetSchemaNode
	attributes := ls.SortEdgesItr(targetSchemaRoot.OutWith(ls.LayerTerms.Attributes)).Targets().All()
	attributes = append(attributes, ls.SortEdgesItr(targetSchemaRoot.OutWith(ls.LayerTerms.AttributeList)).Targets().All()...)

	full := false
	for _, a := range attributes {
		attr := a.(ls.Node)
		found := m.findUnder(sourceDocRoot, attr.GetID())
		var childNode ls.Node
		var err error

		switch {
		case attr.GetTypes().Has(ls.AttributeTypes.Value):
			if found == nil {
				break
			}
			switch {
			case found.GetTypes().Has(ls.AttributeTypes.Value):
				childNode, err = m.mapValue(found, attr, append(path, m.getName(attr)))
			case found.GetTypes().Has(ls.AttributeTypes.Object):
				panic("Unhandled mapping: schema: value, doc: object")
			case found.GetTypes().Has(ls.AttributeTypes.Array):
				panic("Unhandled mapping: schema: value, doc: array")
			default:
				panic("Unhandled mapping")
			}

		case attr.GetTypes().Has(ls.AttributeTypes.Object):
			newName := append(path, m.getName(attr))
			if found == nil {
				childNode, err = m.mapObject(sourceDocRoot, attr, m.NewNode(newName, attr), newName)
				break
			}

			switch {
			case found.GetTypes().Has(ls.AttributeTypes.Value):
				panic("Unhandled mapping: schema: object, doc: value")
			case found.GetTypes().Has(ls.AttributeTypes.Object):
				childNode, err = m.mapObject(found, attr, m.NewNode(newName, attr), newName)
			case found.GetTypes().Has(ls.AttributeTypes.Array):
				panic("Unhandled mapping: schema: object, doc: array")
			default:
				panic("Unhandled mapping")
			}
		case attr.GetTypes().Has(ls.AttributeTypes.Array):
			if found == nil {
				break
			}
			switch {
			case found.GetTypes().Has(ls.AttributeTypes.Value):
				panic("Unhandled mapping: schema: array, doc: value")
			case found.GetTypes().Has(ls.AttributeTypes.Object):
				panic("Unhandled mapping: schema: array, doc: object")
			case found.GetTypes().Has(ls.AttributeTypes.Array):
				panic("Unhandled mapping: schema: array, doc: array")
			default:
				panic("Unhandled mapping")
			}

		default:
			panic("Unhandled mapping")
		}
		if err != nil {
			return nil, err
		}
		if childNode != nil {
			digraph.Connect(newNode, childNode, ls.NewEdge(ls.HasTerm))
			full = true
		}
	}

	if !full {
		return nil, nil
	}

	//for _,a:=range attributes {
	// fill defaults
	//}

	return newNode, nil
}
