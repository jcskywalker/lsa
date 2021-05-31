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

// import (
// 	"encoding/csv"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"os"

// 	"github.com/spf13/cobra"

// 	csvingest "github.com/cloudprivacylabs/lsa/pkg/csv"
// 	"github.com/cloudprivacylabs/lsa/pkg/ls"
// 	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
// )

// func init() {
// 	ingestCmd.AddCommand(ingestCSVCmd)
// 	ingestCSVCmd.Flags().String("schema", "", "Schema id to use")
// 	ingestCSVCmd.Flags().String("profile", "", "CSV profile")
// 	ingestCSVCmd.Flags().String("id", "http://example.org/data", "Base ID to use for ingested nodes")
// 	ingestCSVCmd.Flags().Int("skip", 1, "Number of rows to skip (default 1)")
// 	ingestCSVCmd.MarkFlagRequired("schema")
// }

// var ingestCSVCmd = &cobra.Command{
// 	Use:   "csv",
// 	Short: "Ingest a CSV document and enrich it with a schema",
// 	Args:  cobra.MaximumNArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		f, err := os.Open(args[0])
// 		if err != nil {
// 			failErr(err)
// 		}

// 		var profile csvingest.IngestionProfile
// 		if s, _ := cmd.Flags().GetString("profile"); len(s) > 0 {
// 			if err := readJSON(s, &profile); err != nil {
// 				failErr(err)
// 			}
// 		}

// 		repoDir, _ := cmd.Flags().GetString("repo")
// 		if len(repoDir) == 0 {
// 			fail("Specify a repository directory using --repo")
// 		}
// 		repo := fs.New(repoDir, ls.Terms, func(fname string, err error) {
// 			fmt.Printf("%s: %s\n", fname, err)
// 		})
// 		if err := repo.Load(true); err != nil {
// 			failErr(err)
// 		}
// 		schemaId, _ := cmd.Flags().GetString("schema")
// 		schema := repo.GetSchemaManifest(schemaId)
// 		if schema == nil {
// 			fail(fmt.Sprintf("Schema not found: %s", schemaId))
// 		}
// 		compiler := ls.Compiler{Resolver: func(x string) (string, error) {
// 			if manifest := repo.GetSchemaManifestByObjectType(x); manifest != nil {
// 				return manifest.ID, nil
// 			}
// 			return x, nil
// 		},
// 			Loader: repo.LoadAndCompose,
// 		}
// 		resolved, err := compiler.Compile(schemaId)
// 		if err != nil {
// 			failErr(err)
// 		}

// 		reader := csv.NewReader(f)
// 		skip, _ := cmd.Flags().GetInt("skip")
// 		for i := 0; i < skip; i++ {
// 			row, err := reader.Read()
// 			if err == io.EOF {
// 				return
// 			}
// 			if err != nil {
// 				failErr(err)
// 			}
// 			if i == 0 && len(profile.Columns) == 0 {
// 				profile.Columns, err = csvingest.DefaultProfile(row)
// 				if err != nil {
// 					failErr(err)
// 				}
// 			}
// 		}
// 		if len(profile.Columns) == 0 {
// 			fail("No CSV profile")
// 		}

// 		for {
// 			row, err := reader.Read()
// 			if err == io.EOF {
// 				break
// 			}
// 			if err != nil {
// 				failErr(err)
// 			}
// 			ID, _ := cmd.Flags().GetString("id")
// 			data, err := csvingest.Ingest(ID, row, profile, resolved)
// 			if err != nil {
// 				failErr(err)
// 			}
// 			out, _ := json.MarshalIndent(ls.DataModelToMap(data, true), "", "  ")
// 			fmt.Println(string(out))
// 		}
// 	},
// }
