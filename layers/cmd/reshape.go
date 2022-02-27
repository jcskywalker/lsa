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
	"os"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
	"github.com/cloudprivacylabs/lsa/pkg/transform"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reshapeCmd)
	reshapeCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	reshapeCmd.Flags().String("repo", "", "Schema repository directory")
	reshapeCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	reshapeCmd.PersistentFlags().String("output", "json", "Output format, json, jsonld, or dot")
	reshapeCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
}

var reshapeCmd = &cobra.Command{
	Use:   "reshape",
	Short: "Reshape a graph using a target schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interner := ls.NewInterner()
		input, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, interner, input)
		if err != nil {
			failErr(err)
		}
		compiledSchema, _ := cmd.Flags().GetString("compiledschema")
		repoDir, _ := cmd.Flags().GetString("repo")
		schemaName, _ := cmd.Flags().GetString("schema")
		layer, err := LoadSchemaFromFileOrRepo(compiledSchema, repoDir, schemaName, interner)
		if err != nil {
			failErr(err)
		}

		reshaper := transform.Reshaper{}
		reshaper.Schema = layer
		reshaper.EmbedSchemaNodes = true
		target := graph.NewOCGraph()
		err = reshaper.Reshape(ls.DefaultContext(), g, target)
		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("format")
		err = OutputIngestedGraph(outFormat, target, os.Stdout, false)
		if err != nil {
			failErr(err)
		}
	},
}
