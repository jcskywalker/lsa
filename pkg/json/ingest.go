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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/bserdar/jsonom"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Ingester converts a JSON object model into a graph using a schema
type Ingester struct {
	ls.Ingester

	Interner ls.Interner
}

// deferredProperty is an ingested value that is deferred to be stored
// as a property instead of a node
type deferredProperty struct {
	of    string
	name  string
	value string
}

// IngestBytes ingests JSON bytes
func IngestBytes(ingester *Ingester, baseID string, input []byte) (ls.Node, error) {
	return IngestStream(ingester, baseID, bytes.NewReader(input))
}

// IngestStream ingests JSON stream
func IngestStream(ingester *Ingester, baseID string, input io.Reader) (ls.Node, error) {
	node, err := jsonom.UnmarshalReader(input, ingester.Interner)
	if err != nil {
		return nil, err
	}
	return ingester.Ingest(baseID, node)
}

// Ingest a json document using the schema. The output will have all
// input nodes associated with schema nodes. The ingested object is a
// single instance of an entity.
//
// BaseID is the ID of the root object. All other attribute names are
// generated by appending the attribute path to baseID. BaseID is only
// used if the schema does not explicitly specify an ID
func (ingester *Ingester) Ingest(baseID string, input jsonom.Node) (ls.Node, error) {
	ingester.PreserveNodePaths = true
	path, root := ingester.Start(baseID)
	dn, dp, err := ingester.ingest(input, path, root)
	if err != nil {
		return nil, err
	}
	if len(dp) > 0 {
		return nil, ls.ErrNoParentNode{dp[0].of}
	}
	// Assign node IDs
	if ingester.Schema != nil {
		ls.AssignEntityIDs(dn, func(entity, ID string, node ls.Node, path []ls.Node) string {
			nodePath := ingester.NodePaths[node]
			eid := fmt.Sprintf("%s/%s", entity, ID)
			if len(nodePath) > 1 {
				eid += "/" + ls.NodePath(nodePath[1:]).String()
			}
			return eid
		})
	}

	return dn, err
}

func (ingester *Ingester) ingest(input jsonom.Node, path ls.NodePath, schemaNode ls.Node) (ls.Node, []deferredProperty, error) {

	validate := func(node ls.Node, dp []deferredProperty, err error) (ls.Node, []deferredProperty, error) {
		if err != nil {
			return nil, nil, err
		}
		if err := ingester.Validate(node, schemaNode); err != nil {
			return nil, nil, err
		}
		return node, dp, nil
	}
	// only ingest nodes that have a matching schema attribute
	if schemaNode != nil || !ingester.OnlySchemaAttributes {
		if schemaNode != nil && schemaNode.GetTypes().Has(ls.AttributeTypes.Polymorphic) {
			return validate(ingester.ingestPolymorphicNode(input, path, schemaNode))
		}
		switch next := input.(type) {
		case *jsonom.Object:
			return validate(ingester.ingestObject(next, path, schemaNode))
		case *jsonom.Array:
			return validate(ingester.ingestArray(next, path, schemaNode))
		}
		return validate(ingester.ingestValue(input.(*jsonom.Value), path, schemaNode))
	}
	return nil, nil, nil
}

func (ingester *Ingester) ingestPolymorphicNode(input jsonom.Node, path ls.NodePath, schemaNode ls.Node) (ls.Node, []deferredProperty, error) {
	var dp []deferredProperty
	node, err := ingester.Polymorphic(path, schemaNode, func(p ls.NodePath, optionNode ls.Node) (ls.Node, error) {
		n, x, err := ingester.ingest(input, p, optionNode)
		if err != nil {
			return nil, err
		}
		if x != nil {
			dp = x
		}
		return n, nil
	})
	return node, dp, err
}

func (ingester *Ingester) ingestObject(input *jsonom.Object, path ls.NodePath, schemaNode ls.Node) (ls.Node, []deferredProperty, error) {
	// An object node
	// There is a schema node for this node. It must be an object
	nextNodes, err := ingester.GetObjectAttributeNodes(schemaNode)
	if err != nil {
		return nil, nil, err
	}
	elements := make([]ls.Node, 0, input.Len())
	schemaNodes := make(map[ls.Node]ls.Node)
	dp := make([]deferredProperty, 0)
	for i := 0; i < input.Len(); i++ {
		keyValue := input.N(i)
		schNodes := nextNodes[keyValue.Key()]
		if len(schNodes) > 1 {
			return nil, nil, ls.ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", keyValue.Key()))
		}
		var schNode ls.Node
		if len(schNodes) == 1 {
			schNode = schNodes[0]
		}
		childNode, props, err := ingester.ingest(keyValue.Value(), append(path, keyValue.Key()), schNode)
		if err != nil {
			return nil, nil, ls.ErrDataIngestion{Key: keyValue.Key(), Err: err}
		}
		if childNode != nil {
			schemaNodes[childNode] = schNode
			childNode.GetProperties()[ls.AttributeNameTerm] = ls.StringPropertyValue(keyValue.Key())
			elements = append(elements, childNode)
		}
		if props != nil {
			dp = append(dp, props...)
		}
	}
	node, err := ingester.Object(path, schemaNode, elements, ObjectTypeTerm)
	if err != nil {
		return nil, nil, err
	}
	dp, err = processDeferredProperties(dp, node, func(ID string) (ls.Node, error) {
		done := false
		var ret ls.Node
		for nd, sch := range schemaNodes {
			if sch != nil && sch.GetID() == ID {
				if done {
					return nil, ls.ErrMultipleParentNodes{ID}
				}
				done = true
				ret = nd
			}
		}
		return ret, nil
	})
	return node, dp, err
}

func processDeferredProperties(dp []deferredProperty, node ls.Node, findNode func(string) (ls.Node, error)) ([]deferredProperty, error) {
	dpw := 0
	for i := range dp {
		if len(dp[i].of) == 0 {
			node.GetProperties()[dp[i].name] = ls.StringPropertyValue(dp[i].value)
		} else {
			nd, err := findNode(dp[i].of)
			if err != nil {
				return nil, err
			}
			if nd != nil {
				nd.GetProperties()[dp[i].name] = ls.StringPropertyValue(dp[i].value)
			} else {
				dp[dpw] = dp[i]
				dpw++
			}
		}
	}
	dp = dp[:dpw]
	return dp, nil
}

func (ingester *Ingester) ingestArray(input *jsonom.Array, path ls.NodePath, schemaNode ls.Node) (ls.Node, []deferredProperty, error) {
	elementsNode := ingester.GetArrayElementNode(schemaNode)
	elements := make([]ls.Node, 0, input.Len())

	dp := make([]deferredProperty, 0)
	for index := 0; index < input.Len(); index++ {
		childNode, prop, err := ingester.ingest(input.N(index), path.AppendInt(index), elementsNode)
		if err != nil {
			return nil, nil, ls.ErrDataIngestion{Key: fmt.Sprint(index), Err: err}
		}
		if prop != nil {
			dp = append(dp, prop...)
		}
		if childNode != nil {
			childNode.GetProperties()[ls.AttributeIndexTerm] = ls.StringPropertyValue(fmt.Sprint(index))
			elements = append(elements, childNode)
		}
	}
	node, err := ingester.Array(path, schemaNode, elements, ArrayTypeTerm)
	if err != nil {
		return nil, nil, err
	}
	dp, err = processDeferredProperties(dp, node, func(ID string) (ls.Node, error) {
		if schemaNode != nil && schemaNode.GetID() == ID {
			return node, nil
		}
		return nil, nil
	})

	return node, dp, err
}

func (ingester *Ingester) ingestValue(input *jsonom.Value, path ls.NodePath, schemaNode ls.Node) (ls.Node, []deferredProperty, error) {
	var value interface{}
	var typ string
	if input.Value() != nil {
		switch v := input.Value().(type) {
		case bool:
			value = fmt.Sprint(v)
			typ = BooleanTypeTerm
		case string:
			value = v
			typ = StringTypeTerm
		case uint8, uint16, uint32, uint64, int8, int16, int32, int64, int, uint, float32, float64:
			value = fmt.Sprint(input.Value())
			typ = NumberTypeTerm
		case json.Number:
			value = string(v)
			typ = NumberTypeTerm
		default:
			value = fmt.Sprint(v)
		}
	}
	propertyOf, propertyName := ls.GetAsProperty(schemaNode)
	if len(propertyName) != 0 || len(propertyOf) != 0 {
		if len(propertyName) == 0 {
			propertyName = findStringPath(path)
		}
		return nil, []deferredProperty{{
			of:    propertyOf,
			name:  propertyName,
			value: fmt.Sprint(value),
		}}, nil
	}

	node, err := ingester.Value(path, schemaNode, value, typ)
	return node, nil, err
}

// Locate the last string component in path
func findStringPath(path ls.NodePath) string {
	for i := len(path) - 1; i >= 0; i-- {
		_, err := strconv.Atoi(path[i])
		if err == nil {
			return path[i]
		}
	}
	if len(path) > 0 {
		return fmt.Sprint(path[len(path)-1])
	}
	return "0"
}
