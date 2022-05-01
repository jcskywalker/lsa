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
	"io"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type JSONIngester struct {
	BaseIngestParams
	ID          string
	initialized bool
}

func (ji *JSONIngester) Run(pipeline *PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !ji.initialized {
		layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, ji.CompiledSchema, ji.Repo, ji.Schema, ji.Type, ji.Bundle)
		if err != nil {
			return err
		}
		ji.initialized = true
	}
	var input io.Reader
	if layer != nil {
		enc, err := layer.GetEncoding()
		if err != nil {
			return err
		}
		input, err = cmdutil.StreamFileOrStdin(pipeline.InputFiles, enc)
		if err != nil {
			return err
		}
	} else {
		input, err = cmdutil.StreamFileOrStdin(pipeline.InputFiles)
		if err != nil {
			return err
		}
	}

	parser := jsoningest.Parser{}

	parser.OnlySchemaAttributes = ji.OnlySchemaAttributes
	parser.SchemaNode = layer.GetSchemaRootNode()
	embedSchemaNodes := ji.EmbedSchemaNodes

	builder := ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     embedSchemaNodes,
		OnlySchemaAttributes: parser.OnlySchemaAttributes,
	})
	baseID := ji.ID

	_, err = jsoningest.IngestStream(pipeline.Context, baseID, input, parser, builder)
	if err != nil {
		return err
	}

	if err := pipeline.Next(); err != nil {
		return err
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	operations["ingest/json"] = func() Step { return &JSONIngester{} }
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ing := JSONIngester{}
		ing.fromCmd(cmd)
		ing.ID, _ = cmd.Flags().GetString("id")
		p := []Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, initialGraph, args)
		return err
	},
}
