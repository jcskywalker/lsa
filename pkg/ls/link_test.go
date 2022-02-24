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
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

func TestBasicLink(t *testing.T) {
	schemas := make([]*Layer, 3)
	for i, x := range []string{"testdata/basic_link_test_1.json", "testdata/basic_link_test_2.json", "testdata/basic_link_test_arr.json"} {
		var err error
		schemas[i], err = ReadLayerFromFile(x)
		if err != nil {
			t.Error(err)
			return
		}
	}
	compiler := Compiler{
		Loader: func(ref string) (*Layer, error) {
			for i := range schemas {
				if ref == schemas[i].GetID() {
					return schemas[i], nil
				}
			}
			return nil, fmt.Errorf("Not found: %s", ref)
		},
	}
	layer, err := compiler.Compile(DefaultContext(), schemas[2].GetID())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(layer)

	ingester := Ingester{
		Schema:           layer,
		EmbedSchemaNodes: true,
	}

	path, _ := ingester.Start(DefaultContext(), "root")
	g := graph.NewOCGraph()

	var docRoot1 graph.Node
	{
		attr, _ := layer.FindAttributeByID("https://test_root.id")
		node, _ := ingester.Value(DefaultContext(), g, path.AppendString("id2"), attr, "abc")
		rootNode, _ := layer.FindAttributeByID("http://ref1")
		docRoot1, _ = ingester.Object(DefaultContext(), g, path.AppendString("second"), rootNode, []graph.Node{node})
	}
	var docRoot2 graph.Node
	{
		attr, _ := layer.FindAttributeByID("https://test_ref")
		node, _ := ingester.Value(DefaultContext(), g, path.AppendString("id1"), attr, "abc")
		rootNode, _ := layer.FindAttributeByID("http://ref2")
		docRoot2, _ = ingester.Object(DefaultContext(), g, path.AppendString("first"), rootNode, []graph.Node{node})
	}
	var arr1, arr2 graph.Node
	{
		attr, _ := layer.FindAttributeByID("https://type1")
		arr1, _ = ingester.Array(DefaultContext(), g, path.AppendString("arr1"), attr, []graph.Node{docRoot1})
	}
	{
		attr, _ := layer.FindAttributeByID("https://type2")
		arr2, _ = ingester.Array(DefaultContext(), g, path.AppendString("arr2"), attr, []graph.Node{docRoot2})
	}

	root, _ := ingester.Object(DefaultContext(), g, path.AppendString("rt"), layer.GetSchemaRootNode(), []graph.Node{arr1, arr2})
	ingester.Finish(DefaultContext(), root, nil)

	linkSpecs := GetAllLinkSpecs(root)
	// There must be one
	if len(linkSpecs) != 1 {
		t.Errorf("Expecting 1 linkspec, got %d", len(linkSpecs))
	}
	info := GetDocumentEntityInfo(root)
	// There must be 3
	if len(info) != 3 {
		t.Errorf("Expecting 3 entity info, got %d", len(info))
	}
	for node, spec := range linkSpecs {
		if err := spec.Link(node, info); err != nil {
			t.Error(err)
		}
	}
	// Check if linkNode is linked to the entity
	next := graph.TargetNodes(docRoot2.GetEdgesWithLabel(graph.OutgoingEdge, HasTerm))
	if len(next) == 0 {
		t.Errorf("No link")
	}
	// One of the next nodes must be docRoot1
	found := false
	for _, x := range next {
		if x == docRoot1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Not found")
	}

}
