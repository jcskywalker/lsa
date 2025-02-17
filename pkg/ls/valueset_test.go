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
	"encoding/json"
	"testing"
	//	"github.com/cloudprivacylabs/opencypher/graph"
)

func TestBasicVS(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"valueType": "test",
"layer" :{
  "@type": "Object",
 "@id": "schroot",
  "attributes": {
    "src": {
      "@type": "Value",
      "attributeName": "src",
      "https://lschema.org/vs/context":"schroot",
      "https://lschema.org/vs/resultValues": "tgt"
    },
    "tgt": {
      "@type": "Value",
      "attributeName": "tgt"
    }
  }
}
}`
	var v interface{}
	if err := json.Unmarshal([]byte(schText), &v); err != nil {
		t.Error(err)
		return
	}
	layer, err := UnmarshalLayer(v, nil)
	if err != nil {
		t.Error(err)
		return
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	vsFunc := func(_ *Context, req ValuesetLookupRequest) (ValuesetLookupResponse, error) {
		ret := ValuesetLookupResponse{
			KeyValues: map[string]string{"": "X"},
		}
		return ret, nil
	}
	root := builder.NewNode(layer.GetAttributeByID("schroot"))
	builder.ValueAsNode(layer.GetAttributeByID("src"), root, "a")
	// Graph must have 2 nodes
	if builder.GetGraph().NumNodes() != 2 {
		t.Errorf("NumNodes: %d", builder.GetGraph().NumNodes())
	}

	processor := NewValuesetProcessor(layer, vsFunc)
	DefaultLogLevel = LogLevelDebug
	err = processor.ProcessGraph(DefaultContext(), builder)
	if err != nil {
		t.Error(err)
	}

	// Graph must have 3 nodes
	if builder.GetGraph().NumNodes() != 3 {
		t.Errorf("NumNodes: %d", builder.GetGraph().NumNodes())
	}

	nodes := FindChildInstanceOf(root, "tgt")
	if len(nodes) != 1 {
		t.Errorf("Child nodes: %v", nodes)
	}

}

func TestStructuredVS(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"valueType": "test",
"layer" :{
  "@type": "Object",
 "@id": "schroot",
  "attributes": {
    "src": {
      "@type": "Object",
      "attributeName": "src",
      "https://lschema.org/vs/context":"schroot",
      "https://lschema.org/vs/requestKeys": ["c","s"],
      "https://lschema.org/vs/requestValues": ["code","system"],
      "https://lschema.org/vs/resultKeys": ["tc","ts"],
      "https://lschema.org/vs/resultValues": ["tgtcode","tgtsystem"],
      "attributes": {
        "code": {
          "@type": "Value",
          "attributeName": "code"
        },
        "system": {
          "@type": "Value",
          "attributeName": "system"
        }
      }
    },
    "tgtcode": {
       "@type": "Value",
       "attributeName": "tgtcode"
    },
    "tgtsystem": {
      "@type": "Value",
      "attributeName": "tgtsystem"
    }
  }
}
}`
	var v interface{}
	if err := json.Unmarshal([]byte(schText), &v); err != nil {
		t.Error(err)
		return
	}
	layer, err := UnmarshalLayer(v, nil)
	if err != nil {
		t.Error(err)
		return
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	vsFunc := func(_ *Context, req ValuesetLookupRequest) (ValuesetLookupResponse, error) {
		ret := ValuesetLookupResponse{}
		if req.KeyValues["c"] == "a" && req.KeyValues["s"] == "b" {
			ret.KeyValues = map[string]string{"tc": "aa", "ts": "bb"}
		}
		return ret, nil
	}

	rootNode := builder.NewNode(layer.GetAttributeByID("schroot"))
	srcNode := layer.GetAttributeByID("src")
	codeNode := layer.GetAttributeByID("code")
	systemNode := layer.GetAttributeByID("system")

	_, src, _ := builder.ObjectAsNode(srcNode, rootNode)
	builder.ValueAsNode(codeNode, src, "a")
	builder.ValueAsNode(systemNode, src, "b")

	// Graph must have 4 nodes
	if builder.GetGraph().NumNodes() != 4 {
		t.Errorf("NumNodes: %d", builder.GetGraph().NumNodes())
	}
	processor := NewValuesetProcessor(layer, vsFunc)
	DefaultLogLevel = LogLevelDebug
	ctx := DefaultContext()
	err = processor.ProcessGraph(ctx, builder)
	if err != nil {
		t.Error(err)
	}

	// Graph must have 6 nodes
	if builder.GetGraph().NumNodes() != 6 {
		t.Errorf("NumNodes: %d", builder.GetGraph().NumNodes())
	}

	tgtCodeNodes := FindChildInstanceOf(rootNode, "tgtcode")
	if len(tgtCodeNodes) != 1 {
		t.Errorf("No tgtcode")
	}
	tgtSystemNodes := FindChildInstanceOf(rootNode, "tgtsystem")
	if len(tgtSystemNodes) != 1 {
		t.Errorf("No tgtsystem")
	}
}
