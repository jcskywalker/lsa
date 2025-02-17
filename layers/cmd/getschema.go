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
	"github.com/spf13/cobra"
)

// Following structs are used to define a Layered Schema
type Attribute struct {
	ID            string `json:"@id"`
	AttributeName string `json:"attributeName"`
	Types         string `json:"@type"`
}

type Layer struct {
	AttributeList []Attribute `json:"attributeList"`
}

type LS struct {
	Context string `json:"@context"`
	ID      string `json:"@id"`
	Type    string `json:"@type"`
	Layer   Layer  `json:"layer"`
}

func init() {
	rootCmd.AddCommand(getschemaCmd)
}

var getschemaCmd = &cobra.Command{
	Use:   "getschema",
	Short: "Write layered schema based on input file",
}
