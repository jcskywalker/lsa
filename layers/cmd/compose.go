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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	rootCmd.AddCommand(composeCmd)
	composeCmd.Flags().StringP("output", "o", "", "Output file")
	composeCmd.Flags().String("repo", "", "Schema repository directory. If a repository is given, all layers are resolved using that repository. Otherwise, all layers are read as files.")
}

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a schema from components",
	Long:  `Compose a schema from components and output the resulting schema layer.`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoDir, _ := cmd.Flags().GetString("repo")
		interner := ls.NewInterner()
		var output *ls.Layer
		if len(repoDir) == 0 {
			inputs, err := cmdutil.ReadJSONMultiple(args)
			if err != nil {
				failErr(err)
			}
			for i, input := range inputs {
				layer, err := ls.UnmarshalLayer(input, interner)
				if err != nil {
					fail(fmt.Sprintf("Cannot unmarshal %s: %v", args[i], err))
				}
				if output == nil {
					output = layer
				} else {
					if err := output.Compose(ls.DefaultContext(), layer); err != nil {
						fail(fmt.Sprintf("Cannot compose %s: %s", args[i], err))
					}
				}
			}
		} else {
			repo, err := getRepo(repoDir, interner)
			if err != nil {
				failErr(err)
			}
			output, err = repo.GetComposedSchema(ls.DefaultContext(), args[0])
			if err != nil {
				failErr(err)
			}
		}
		if output != nil {
			out, _ := ls.MarshalLayer(output)
			d, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(d))
		}
	},
}
