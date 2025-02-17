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
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type ErrReingest struct {
	Wrapped      error
	NodeID       string
	SchemaNodeID string
	Msg          string
}

func (e ErrReingest) Error() string {
	return fmt.Sprintf("NodeID: %s SchemaNodeID: %s Err: %s Msg: %s", e.NodeID, e.SchemaNodeID, e.Wrapped, e.Msg)
}

func (e ErrReingest) Unwrap() error { return e.Wrapped }

func Reingest(ctx *ls.Context, sourceNode graph.Node, target ls.GraphBuilder, variant *ls.Layer) error {
	nodeMap := make(map[graph.Node]graph.Node)
	graphPath := make([]graph.Node, 0, 16)
	err := reingestNode(ctx, sourceNode, target, variant, graphPath, nodeMap)
	return err
}

func reingestNode(ctx *ls.Context, sourceNode graph.Node, target ls.GraphBuilder, variant *ls.Layer, graphPath []graph.Node, nodeMap map[graph.Node]graph.Node) error {
	// Node processed already?
	if _, exists := nodeMap[sourceNode]; exists {
		return nil
	}

	schemaNodeID := ls.AsPropertyValue(sourceNode.GetProperty(ls.SchemaNodeIDTerm)).AsString()
	var schemaNode graph.Node
	if len(schemaNodeID) > 0 {
		schemaNode = variant.GetAttributeByID(schemaNodeID)
	}
	nodeID := ls.GetNodeID(sourceNode)
	var nodeValue interface{}
	var rawValue string
	var err error
	if sourceNode.HasLabel(ls.AttributeTypeValue) {
		rawValue, _ = ls.GetRawNodeValue(sourceNode)
		nodeValue, err = ls.GetNodeValue(sourceNode)
		if err != nil {
			return ErrReingest{Wrapped: err, NodeID: nodeID, SchemaNodeID: schemaNodeID, Msg: "Cannot get node value"}
		}
	}
	var parentNode graph.Node
	if len(graphPath) > 0 {
		parentNode = graphPath[len(graphPath)-1]
	}
	// Ingest the node
	switch {
	case sourceNode.HasLabel(ls.AttributeTypeValue):
		switch ls.GetIngestAs(schemaNode) {
		case "node":
			_, node, err := target.ValueAsNode(schemaNode, parentNode, "")
			if err != nil {
				return err
			}
			if err := ls.SetNodeValue(node, nodeValue); err != nil {
				return err
			}
			nodeMap[sourceNode] = node

		case "edge":
			edge, err := target.ValueAsEdge(schemaNode, parentNode, rawValue)
			if err != nil {
				return err
			}
			nodeMap[sourceNode] = edge.GetTo()

		case "property":
			err := target.ValueAsProperty(schemaNode, graphPath, rawValue)
			if err != nil {
				return err
			}
		}
		return nil

	case sourceNode.HasLabel(ls.AttributeTypeObject) || sourceNode.HasLabel(ls.AttributeTypeArray):
		var typeTerm string
		if sourceNode.HasLabel(ls.AttributeTypeObject) {
			typeTerm = ls.AttributeTypeObject
		} else {
			typeTerm = ls.AttributeTypeArray
		}
		var newNode graph.Node
		switch ls.GetIngestAs(schemaNode) {
		case "node":
			_, node, err := target.CollectionAsNode(schemaNode, parentNode, typeTerm)
			if err != nil {
				return err
			}
			nodeMap[sourceNode] = node
			newNode = node

		case "edge":
			edge, err := target.CollectionAsEdge(schemaNode, parentNode, typeTerm)
			if err != nil {
				return err
			}
			nodeMap[sourceNode] = edge.GetTo()
			newNode = edge.GetTo()
		}
		for edges := sourceNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			node := edge.GetTo()
			if !node.HasLabel(ls.DocumentNodeTerm) {
				continue
			}
			if err := reingestNode(ctx, node, target, variant, append(graphPath, newNode), nodeMap); err != nil {
				return err
			}
		}
	}
	return nil
}
