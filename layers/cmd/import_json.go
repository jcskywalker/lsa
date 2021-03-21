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
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/spf13/cobra"

	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	importCmd.AddCommand(importJSONCmd)
}

const LayersContextURL = "http://schemas.cloudprivacylabs.com/layers.jsonld"
const SchemaContextURL = "http://schemas.cloudprivacylabs.com/schema.jsonld"

var importJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Import a JSON schema and slice into its layers",
	Long: `Input a JSON file of the format:

{
  "entities": [
     {
       "ref": "<reference to schema>",
       "name": "entity name"
     },
    ...
   ],
  "schema": "schema output file",
  "schemaId": "schema id",
  "objectType": "object type",
  "schemaBase": "schema base output file",
  "schemaBaseId": "schema base id",
  "overlays": [
     {
         "@id": "output object id",
         "terms": [ terms to include in the overlay ],
         "file": "output file"
     },
     ...
   ]
}

Each element in the input will be compiled, and sliced into a schema base plus
overlays, and saved as a schema+schema base+overlays. The file names, @ids, and terms
are Go templates, you can reference entity names and references using {{.name}} and {{.ref}}.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputData, err := ioutil.ReadFile(args[0])
		if err != nil {
			failErr(err)
		}

		type overlay struct {
			Terms  []string `json:"terms"`
			termst []*template.Template
			File   string `json:"file"`
			filet  *template.Template
			ID     string `json:"@id"`
			idt    *template.Template
		}

		type request struct {
			Entities        []jsonsch.Entity `json:"entities"`
			SchemaID        string           `json:"schemaId"`
			schemaIDt       *template.Template
			SchemaBaseID    string `json:"schemaBaseId"`
			schemaBaseIDt   *template.Template
			Schema          string `json:"schema"`
			schemat         *template.Template
			ObjectType      string `json:"objectType"`
			objectTypet     *template.Template
			SchemaBase      string `json:"schemaBase"`
			schemaBaset     *template.Template
			Overlays        []overlay `json:"overlays"`
			SchemaBaseTerms []string  `json:"schemaBaseTerms"`
		}

		var req request
		if err := json.Unmarshal(inputData, &req); err != nil {
			failErr(err)
		}
		if len(req.Entities) == 0 {
			return
		}

		req.schemaIDt = template.Must(template.New("schemaID").Parse(req.SchemaID))
		req.schemaBaseIDt = template.Must(template.New("schemaBaseID").Parse(req.SchemaBaseID))
		req.schemaBaset = template.Must(template.New("schemaBase").Parse(req.SchemaBase))
		req.schemat = template.Must(template.New("schema").Parse(req.Schema))
		req.objectTypet = template.Must(template.New("objectType").Parse(req.ObjectType))
		for i := range req.Overlays {
			for _, x := range req.Overlays[i].Terms {
				req.Overlays[i].termst = append(req.Overlays[i].termst, template.Must(template.New(fmt.Sprintf("terms-%d", i)).Parse(x)))
			}
			req.Overlays[i].filet = template.Must(template.New(fmt.Sprintf("file-%d", i)).Parse(req.Overlays[i].File))
			req.Overlays[i].idt = template.Must(template.New(fmt.Sprintf("id-%d", i)).Parse(req.Overlays[i].ID))
		}

		exec := func(t *template.Template, entity jsonsch.Entity) string {
			tdata := map[string]interface{}{"name": entity.Name, "ref": entity.Ref}
			out := bytes.Buffer{}
			if err := t.Execute(&out, tdata); err != nil {
				panic(fmt.Errorf("During %s: %w", entity.Name, err))
			}
			return out.String()
		}
		for i := range req.Entities {
			req.Entities[i].SchemaName = exec(req.schemaIDt, req.Entities[i])
		}

		compiled, err := jsonsch.Compile(req.Entities)
		if err != nil {
			failErr(err)
		}
		results, err := jsonsch.Import(compiled)
		if err != nil {
			failErr(err)
		}

		for i, item := range results {
			if len(req.SchemaBase) > 0 {
				layer := ls.NewEmptySchemaLayer()
				a := item.BaseAttributes.Slice(func(term string, attribute *ls.Attribute) bool {
					if len(term) == 0 {
						return true
					}
					for _, x := range req.SchemaBaseTerms {
						if x == term {
							return true
						}
					}
					return false
				})
				if a != nil {
					layer.Attributes = *a
				}
				layer.ID = exec(req.schemaBaseIDt, req.Entities[i])
				layer.ObjectType = exec(req.objectTypet, req.Entities[i])
				layer.Type = ls.TermSchemaBaseType
				data, err := json.MarshalIndent(layer.MarshalExpanded(), "", "  ")
				if err != nil {
					failErr(err)
				}
				ioutil.WriteFile(exec(req.schemaBaset, req.Entities[i]), data, 0664)
			}

			overlayIDs := make([]string, 0)
			for _, ovl := range req.Overlays {
				inclterms := make(map[string]struct{})
				for _, x := range ovl.termst {
					inclterms[exec(x, req.Entities[i])] = struct{}{}
				}
				layer := ls.NewEmptySchemaLayer()
				a := item.BaseAttributes.Slice(func(term string, attribute *ls.Attribute) bool {
					_, ok := inclterms[term]
					return ok
				})
				if a != nil {
					layer.Attributes = *a
				}
				layer.ID = exec(ovl.idt, req.Entities[i])
				layer.Type = ls.TermOverlayType
				layer.ObjectType = exec(req.objectTypet, req.Entities[i])
				overlayIDs = append(overlayIDs, layer.ID)
				data, err := json.MarshalIndent(layer.MarshalExpanded(), "", "  ")
				if err != nil {
					failErr(err)
				}
				ioutil.WriteFile(exec(ovl.filet, req.Entities[i]), data, 0664)
			}

			if len(req.Schema) > 0 {
				sch := ls.Schema{
					ID:         exec(req.schemaIDt, req.Entities[i]),
					ObjectType: exec(req.objectTypet, req.Entities[i]),
					SchemaBase: exec(req.schemaBaseIDt, req.Entities[i]),
					Overlays:   overlayIDs,
				}
				data, _ := json.MarshalIndent(sch.MarshalExpanded(), "", "  ")
				ioutil.WriteFile(exec(req.schemat, req.Entities[i]), data, 0664)
			}
		}
	},
}
