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

package ls

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher/graph"
)

type GraphBuilderOptions struct {
	// If true, schema node properties are embedded into document
	// nodes. If false, schema nodes are preserved as separate nodes,
	// with an instanceOf edge between the document node to the schema
	// node.
	EmbedSchemaNodes bool
	// If OnlySchemaAttributes is true, only ingest data points if there is a schema for it.
	// If OnlySchemaAttributes is false, ingest whether or not there is a schema for it.
	OnlySchemaAttributes bool
}

// GraphBuilder contains the methods to ingest a graph
type GraphBuilder struct {
	options *GraphBuilderOptions
	// SchemaNodeMap keeps the map of schema nodes copied into the target graph
	schemaNodeMap map[graph.Node]graph.Node
	targetGraph   graph.Graph
}

type ErrCannotInstantiateSchemaNode struct {
	SchemaNodeID string
	Reason       string
}

func (e ErrCannotInstantiateSchemaNode) Error() string {
	return fmt.Sprintf("Cannot instantiate schema node %s because: %s", e.SchemaNodeID, e.Reason)
}

// NewGraphBuilder returns a new builder with an optional graph. If g
// is nil, a new graph is initialized
func NewGraphBuilder(g graph.Graph, options GraphBuilderOptions) GraphBuilder {
	if g == nil {
		g = NewDocumentGraph()
	}
	ret := GraphBuilder{
		options:       &options,
		targetGraph:   g,
		schemaNodeMap: make(map[graph.Node]graph.Node),
	}
	return ret
}

func (gb GraphBuilder) GetOptions() GraphBuilderOptions {
	return *gb.options
}

func (gb GraphBuilder) GetGraph() graph.Graph {
	return gb.targetGraph
}

func determineEdgeLabel(schemaNode graph.Node) string {
	if x, ok := schemaNode.GetProperty(EdgeLabelTerm); ok {
		if label := x.(*PropertyValue).AsString(); len(label) > 0 {
			return label
		}
	}
	if x, ok := schemaNode.GetProperty(AttributeNameTerm); ok {
		if label := x.(*PropertyValue).AsString(); len(label) > 0 {
			return label
		}
	}
	return ""
}

// NewNode creates a new graph node as an instance of SchemaNode. Then
// it either merges schema properties into the new node, or creates an
// instanceOf edge to the schema node.
func (gb GraphBuilder) NewNode(schemaNode graph.Node) graph.Node {
	types := []string{DocumentNodeTerm}
	if schemaNode != nil {
		for l := range schemaNode.GetLabels() {
			if l != AttributeNodeTerm {
				types = append(types, l)
			}
		}
	}
	newNode := gb.targetGraph.NewNode(types, nil)
	if schemaNode == nil {
		return newNode
	}
	newNode.SetProperty(SchemaNodeIDTerm, StringPropertyValue(GetNodeID(schemaNode)))
	// If this is an entity boundary, mark it
	if pv, rootNode := schemaNode.GetProperty(EntitySchemaTerm); rootNode {
		newNode.SetProperty(EntitySchemaTerm, pv)
	}

	copyNodesAttachedToSchema := func(targetNode graph.Node) {
		for edges := schemaNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			if IsAttributeTreeEdge(edge) {
				continue
			}
			graph.CopySubgraph(edge.GetTo(), gb.targetGraph, ClonePropertyValueFunc, gb.schemaNodeMap)
			gb.targetGraph.NewEdge(targetNode, gb.schemaNodeMap[edge.GetTo()], edge.GetLabel(), nil)
		}
	}

	if gb.options.EmbedSchemaNodes {
		schemaNode.ForEachProperty(func(k string, v interface{}) bool {
			if k == NodeIDTerm {
				return true
			}
			if _, ok := newNode.GetProperty(k); !ok {
				if pv, ok := v.(*PropertyValue); ok {
					newNode.SetProperty(k, pv.Clone())
				} else {
					newNode.SetProperty(k, v)
				}
			}
			return true
		})
		copyNodesAttachedToSchema(newNode)
		return newNode
	}
	pat := graph.Pattern{{
		Labels:     graph.NewStringSet(AttributeNodeTerm),
		Properties: map[string]interface{}{NodeIDTerm: GetNodeID(schemaNode)},
	}}
	nodes, _ := pat.FindNodes(gb.targetGraph, nil)
	// Copy the schema node into this
	// If the schema node already exists in the target graph, use it
	if len(nodes) != 0 {
		gb.targetGraph.NewEdge(newNode, nodes[0], InstanceOfTerm, nil)
	} else {
		// Copy the node
		newSchemaNode := graph.CopyNode(schemaNode, gb.targetGraph, ClonePropertyValueFunc)
		gb.schemaNodeMap[schemaNode] = newSchemaNode
		gb.targetGraph.NewEdge(newNode, newSchemaNode, InstanceOfTerm, nil)
		copyNodesAttachedToSchema(newSchemaNode)
	}
	return newNode
}

func (gb GraphBuilder) setEntityID(value string, parentDocumentNode, schemaNode graph.Node) error {
	entityRootNode := GetEntityRootNode(parentDocumentNode)
	if entityRootNode == nil {
		return nil
	}
	idFieldsProp := GetEntityIDFields(entityRootNode)
	idFields := idFieldsProp.MustStringSlice()
	if len(idFields) == 0 {
		return nil
	}
	schemaNodeID := GetNodeID(schemaNode)
	idIndex := -1
	for i, idField := range idFields {
		if schemaNodeID == idField {
			idIndex = i
			break
		}
	}
	// Is this an ID field?
	if idIndex == -1 {
		return nil
	}

	// Get existing ID
	entityID := AsPropertyValue(entityRootNode.GetProperty(EntityIDTerm))

	existingEntityIDSlice := entityID.MustStringSlice()
	for len(existingEntityIDSlice) <= idIndex {
		existingEntityIDSlice = append(existingEntityIDSlice, "")
	}
	existingEntityIDSlice[idIndex] = value

	if idFieldsProp.IsString() {
		entityRootNode.SetProperty(EntityIDTerm, StringPropertyValue(value))
		return nil
	}
	entityRootNode.SetProperty(EntityIDTerm, StringSlicePropertyValue(existingEntityIDSlice))
	return nil
}

//  ValueAsEdge ingests a value using the following scheme:
//
//  input: (name: value)
//  output: --(label)-->(value:value, attributeName:name)
//
// where label=attributeName (in this case "name") if edgeLabel is not
// specified in schema.
func (gb GraphBuilder) ValueAsEdge(schemaNode, parentDocumentNode graph.Node, value string, types ...string) (graph.Edge, error) {
	var edgeLabel string
	if schemaNode != nil {
		gb.setEntityID(value, parentDocumentNode, schemaNode)
		if !schemaNode.HasLabel(AttributeTypeValue) {
			return nil, ErrSchemaValidation{Msg: "A value is expected here"}
		}
		edgeLabel = determineEdgeLabel(schemaNode)
		if len(edgeLabel) == 0 {
			return nil, ErrCannotDetermineEdgeLabel{SchemaNodeID: GetNodeID(schemaNode)}
		}
	} else if gb.options.OnlySchemaAttributes {
		return nil, nil
	}
	node := gb.NewNode(schemaNode)
	SetRawNodeValue(node, value)
	t := node.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue)
	node.SetLabels(t)
	edge := gb.targetGraph.NewEdge(parentDocumentNode, node, edgeLabel, nil)
	return edge, nil
}

// ValueAsNode creates a new value node. The new node has the given value
// and the types
func (gb GraphBuilder) ValueAsNode(schemaNode, parentDocumentNode graph.Node, value string, types ...string) (graph.Edge, graph.Node, error) {
	if schemaNode != nil {
		gb.setEntityID(value, parentDocumentNode, schemaNode)
		if !schemaNode.HasLabel(AttributeTypeValue) {
			return nil, nil, ErrSchemaValidation{Msg: "A value expected here"}
		}
	} else {
		if gb.options.OnlySchemaAttributes {
			return nil, nil, nil
		}
	}
	newNode := gb.NewNode(schemaNode)
	SetRawNodeValue(newNode, value)
	t := newNode.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue)
	newNode.SetLabels(t)
	var edge graph.Edge
	if parentDocumentNode != nil {
		edge = gb.targetGraph.NewEdge(parentDocumentNode, newNode, HasTerm, nil)
	}
	return edge, newNode, nil
}

// ValueAsProperty ingests a value as a property of an ancestor node. The ancestor
func (gb GraphBuilder) ValueAsProperty(schemaNode graph.Node, graphPath []graph.Node, value string) error {
	// Schema node cannot be nil here
	if schemaNode == nil {
		return ErrInvalidInput{Msg: "Missing schema node"}
	}
	asPropertyOf := AsPropertyValue(schemaNode.GetProperty(AsPropertyOfTerm)).AsString()
	propertyName := AsPropertyValue(schemaNode.GetProperty(PropertyNameTerm)).AsString()
	if len(propertyName) == 0 {
		propertyName = AsPropertyValue(schemaNode.GetProperty(AttributeNameTerm)).AsString()
	}
	if len(propertyName) == 0 {
		return ErrCannotDeterminePropertyName{SchemaNodeID: GetNodeID(schemaNode)}
	}
	var targetNode graph.Node
	if len(asPropertyOf) == 0 {
		targetNode = graphPath[len(graphPath)-1]
	} else {
		// Find ancestor that is instance of asPropertyOf
		for i := len(graphPath) - 1; i >= 0; i-- {
			if AsPropertyValue(graphPath[i].GetProperty(SchemaNodeIDTerm)).AsString() == asPropertyOf {
				targetNode = graphPath[i]
				break
			}
		}
	}
	if targetNode == nil {
		return ErrCannotFindAncestor{SchemaNodeID: GetNodeID(schemaNode)}
	}
	targetNode.SetProperty(propertyName, StringPropertyValue(value))
	return nil
}

func (gb GraphBuilder) CollectionAsNode(schemaNode, parentNode graph.Node, typeTerm string, types ...string) (graph.Edge, graph.Node, error) {
	if schemaNode != nil {
		if !schemaNode.HasLabel(typeTerm) {
			return nil, nil, ErrSchemaValidation{Msg: fmt.Sprintf("A %s is expected here but found %s", typeTerm, schemaNode.GetLabels())}
		}
	}
	if schemaNode == nil && gb.options.OnlySchemaAttributes {
		return nil, nil, nil
	}
	ret := gb.NewNode(schemaNode)
	t := ret.GetLabels()
	t.Add(types...)
	t.Add(typeTerm)
	ret.SetLabels(t)
	var edge graph.Edge
	if parentNode != nil {
		edge = gb.targetGraph.NewEdge(parentNode, ret, HasTerm, nil)
	}
	return edge, ret, nil
}

func (gb GraphBuilder) CollectionAsEdge(schemaNode, parentNode graph.Node, typeTerm string, types ...string) (graph.Edge, error) {
	if schemaNode != nil {
		if !schemaNode.HasLabel(typeTerm) {
			return nil, ErrSchemaValidation{Msg: fmt.Sprintf("A %s is expected here but found %s", typeTerm, schemaNode.GetLabels())}
		}
	}
	if schemaNode == nil && gb.options.OnlySchemaAttributes {
		return nil, nil
	}
	if parentNode == nil {
		return nil, ErrDataIngestion{Err: fmt.Errorf("Document root object cannot be an edge")}
	}
	blankNode := gb.NewNode(schemaNode)
	edgeLabel := determineEdgeLabel(schemaNode)
	if len(edgeLabel) == 0 {
		return nil, ErrCannotDetermineEdgeLabel{SchemaNodeID: GetNodeID(schemaNode)}
	}
	t := blankNode.GetLabels()
	t.Add(types...)
	t.Add(typeTerm)
	blankNode.SetLabels(t)
	edge := gb.targetGraph.NewEdge(parentNode, blankNode, edgeLabel, nil)
	return edge, nil
}

// ObjectAsNode creates a new object node
func (gb GraphBuilder) ObjectAsNode(schemaNode, parentNode graph.Node, types ...string) (graph.Edge, graph.Node, error) {
	return gb.CollectionAsNode(schemaNode, parentNode, AttributeTypeObject, types...)
}

func (gb GraphBuilder) ArrayAsNode(schemaNode, parentNode graph.Node, types ...string) (graph.Edge, graph.Node, error) {
	return gb.CollectionAsNode(schemaNode, parentNode, AttributeTypeArray, types...)
}

// ObjectAsEdge creates an object node as an edge using the following scheme:
//
//  parent --object--> _blankNode --...
func (gb GraphBuilder) ObjectAsEdge(schemaNode, parentNode graph.Node, types ...string) (graph.Edge, error) {
	return gb.CollectionAsEdge(schemaNode, parentNode, AttributeTypeObject, types...)
}

func (gb GraphBuilder) ArrayAsEdge(schemaNode, parentNode graph.Node, types ...string) (graph.Edge, error) {
	return gb.CollectionAsEdge(schemaNode, parentNode, AttributeTypeArray, types...)
}

// Link the given node, or create a link from the parent node.
//
// `spec` is the link spec. `docNode` contains the ingested document
// node that will be linked. It can be nil. `parentNode` is the
// document node containing the docNode.
func (gb GraphBuilder) LinkNode(spec *LinkSpec, docNode, parentNode graph.Node, entityInfo map[graph.Node]EntityInfo) error {
	entityRoot := GetEntityRoot(parentNode)
	if entityRoot == nil {
		return ErrCannotResolveLink(*spec)
	}
	foreignKeyNodes := make([][]graph.Node, len(spec.FK))
	IterateDescendants(entityRoot, func(n graph.Node) bool {
		attrId := AsPropertyValue(n.GetProperty(SchemaNodeIDTerm)).AsString()
		if len(attrId) == 0 {
			return true
		}
		for i := range spec.FK {
			if spec.FK[i] == attrId {
				foreignKeyNodes[i] = append(foreignKeyNodes[i], n)
			}
		}
		return true
	}, OnlyDocumentNodes, false)
	// All foreign key elements must have the same number of elements, and no index must be skipped
	var numKeys int
	for index := 0; index < len(foreignKeyNodes); index++ {
		if index == 0 {
			numKeys = len(foreignKeyNodes[index])
		} else {
			if len(foreignKeyNodes[index]) != numKeys {
				return ErrInvalidForeignKeys{Spec: *spec, Msg: "Inconsistent foreign keys"}
			}
		}
	}
	if numKeys == 0 {
		// Nothing to link
		return nil
	}
	if numKeys > 1 && !spec.Multi {
		return ErrInvalidForeignKeys{Spec: *spec, Msg: "Multiple foreign key values not allowed"}
	}

	g := parentNode.GetGraph()
	var nodeProperties map[string]interface{}
	if spec.IngestAs == IngestAsEdge && docNode != nil {
		// This document node is removed and a link from the parent to the target is created
		nodeProperties = CloneProperties(docNode)
		docNode.DetachAndRemove()
	}

	// Find remote references
	for i := 0; i < numKeys; i++ {
		fk := make([]string, len(foreignKeyNodes))
		for k, v := range foreignKeyNodes {
			fk[k], _ = GetRawNodeValue(v[i])
		}
		ref, err := spec.FindReference(entityInfo, fk)
		if err != nil {
			return err
		}
		if len(ref) == 0 {
			continue
		}
		for _, linkRef := range ref {
			if spec.IngestAs == IngestAsEdge {
				// Node is already removed. Make an edge
				if spec.Forward {
					g.NewEdge(parentNode, linkRef, spec.Label, nodeProperties)
				} else {
					g.NewEdge(linkRef, parentNode, spec.Label, nodeProperties)
				}
			} else {
				if docNode == nil {
					docNode = gb.NewNode(spec.SchemaNode)
					gb.targetGraph.NewEdge(parentNode, docNode, HasTerm, nil)
				}
				// A link from this document node to target is created
				if spec.Forward {
					gb.targetGraph.NewEdge(docNode, linkRef, spec.Label, nil)
				} else {
					gb.targetGraph.NewEdge(linkRef, docNode, spec.Label, nil)
				}
			}
		}
	}

	return nil
}

func (gb GraphBuilder) LinkNodes(schema *Layer, entityInfo map[graph.Node]EntityInfo) error {
	for nodes := schema.Graph.GetNodes(); nodes.Next(); {
		attrNode := nodes.Node()
		ls, err := GetLinkSpec(attrNode)
		if err != nil {
			return err
		}
		if ls == nil {
			continue
		}
		attrId := GetNodeID(attrNode)
		// Found a link spec. Find corresponding parent nodes in the document
		parentSchemaNode := GetParentAttribute(attrNode)
		// Find nodes that are instance of this node
		parentDocNodes := GetNodesInstanceOf(gb.targetGraph, GetNodeID(parentSchemaNode))
		for _, parent := range parentDocNodes {
			// Each parent node has at least one reference node child
			childFound := false
			for edges := parent.GetEdges(graph.OutgoingEdge); edges.Next(); {
				childNode := edges.Edge().GetTo()
				if !IsDocumentNode(childNode) {
					continue
				}
				if AsPropertyValue(childNode.GetProperty(SchemaNodeIDTerm)).AsString() != attrId {
					continue
				}
				// childNode is an instance of attrNode, which is a link
				childFound = true
				if err := gb.LinkNode(ls, childNode, parent, entityInfo); err != nil {
					return err
				}
			}
			if !childFound {
				if err := gb.LinkNode(ls, nil, parent, entityInfo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
