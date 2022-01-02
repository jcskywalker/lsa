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
	"encoding/json"
	"testing"

	"github.com/bserdar/digraph"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type mapTestCase struct {
	Name          string      `json:"name"`
	DocGraph      interface{} `json:"doc"`
	TargetSchema  interface{} `json:"targetSchema"`
	ExpectedGraph interface{} `json:"expected"`
}

func (tc mapTestCase) GetName() string { return tc.Name }

func (tc mapTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)

	d, _ := json.Marshal(tc.DocGraph)
	docGraph := digraph.New()
	err := ls.UnmarshalGraphJSON(d, docGraph, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal docgraph: %v", tc.Name, err)
		return
	}

	targetSchema, err := ls.UnmarshalLayer(tc.TargetSchema, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal target layer: %v", tc.Name, err)
		return
	}

	mapper := GraphMapper{
		Ingester: ls.Ingester{
			Schema:           targetSchema,
			EmbedSchemaNodes: true,
		},
	}

	result, err := mapper.Map(digraph.Sources(docGraph.GetIndex())[0].(ls.Node))
	if err != nil {
		t.Errorf("Test case: %s Cannot map: %v", tc.Name, err)
		return
	}

	d, _ = json.Marshal(tc.ExpectedGraph)
	expectedGraph := digraph.New()
	err = ls.UnmarshalGraphJSON(d, expectedGraph, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal expectedGraph: %v", tc.Name, err)
		return
	}
	resultGraph := digraph.New()
	resultGraph.AddNode(result)
	eq := digraph.CheckIsomorphism(resultGraph.GetIndex(), expectedGraph.GetIndex(), func(n1, n2 digraph.Node) bool { return true }, func(e1, e2 digraph.Edge) bool { return true })

	if !eq {
		r, _ := ls.MarshalGraphJSON(resultGraph)
		t.Errorf("Test case: %s Result is different from the expected: Result: %v Expected: %v", tc.Name, string(r), ls.ToMap(tc.ExpectedGraph))
	}
}

func TestMapping(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c mapTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/map.json", run)
}
