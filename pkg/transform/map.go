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
	return m.mapGraph(docRoot, m.Schema.GetSchemaRootNode(), nil)
}

// mapGraph converts the document at `sourceDocRoot` to `targetSchemaRoot` using
// `map` properties of the source doc nodes.
func (m GraphMapper) mapGraph(sourceDocRoot, targetSchemaRoot ls.Node, path []interface{}) (result ls.Node, err error) {
	var name interface{}
	attributeName := targetSchemaRoot.GetProperties()[ls.AttributeNameTerm]
	if attributeName != nil && attributeName.IsString() {
		name = attributeName.AsString()
	} else {
		name = targetSchemaRoot.GetID()
	}
	schemaTypes := targetSchemaRoot.GetTypes()
	switch {
	case schemaTypes.Has(ls.AttributeTypes.Value):
		newNode := m.NewNode(append(path, name), targetSchemaRoot)
		value := ls.GetNodeFilteredValue(sourceDocRoot)
		newNode.SetValue(value)
		return newNode, nil

	case schemaTypes.Has(ls.AttributeTypes.Object):
		newNode := m.NewNode(append(path, name), targetSchemaRoot)
		return m.mapObject(sourceDocRoot, targetSchemaRoot, newNode, path)

	case schemaTypes.Has(ls.AttributeTypes.Array):
	case schemaTypes.Has(ls.AttributeTypes.Polymorphic):
		panic("Unsupported polymorphic mapping")
	}
	panic("Unsupported mapping")
}

func (m GraphMapper) mapObject(sourceDocRoot, targetSchemaRoot ls.Node, newNode ls.Node, path []interface{}) (ls.Node, error) {
	// sourceDocRoot will be used as the root for all searches
	// We will recurse into targetSchemaRoot
	// newNode is the node corresponding to the instance of targetSchemaNode
	attributes := ls.SortEdgesItr(targetSchemaRoot.OutWith(ls.LayerTerms.Attributes)).Targets().All()
	attributes = append(attributes, ls.SortEdgesItr(targetSchemaRoot.OutWith(ls.LayerTerms.AttributeList)).Targets().All()...)

	for _, a := range attributes {
		attr := a.(ls.Node)

	}
}
