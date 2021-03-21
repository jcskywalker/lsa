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
	"errors"
	"fmt"
)

// ErrInvalidInput is used for invalid input to an api
var ErrInvalidInput = errors.New("Invalid input")

// ErrAttributeWithoutID is returned when an attribute is parsed without an id
var ErrAttributeWithoutID = errors.New("Attribute without ID")

// ErrDuplicateAttribute is returned when a duplicate attribute ID is
// detected in a layer
type ErrDuplicateAttribute string

// Error returns a message containing the duplicate ID
func (e ErrDuplicateAttribute) Error() string {
	return fmt.Sprintf("Duplicate attribute: %s", string(e))
}

// ErrInvalidObject is returned for objects that cannot be
// interpreted, like objects containing both reference and attributes,
// etc.
type ErrInvalidObject string

func (e ErrInvalidObject) Error() string {
	return fmt.Sprintf("Invalid object: %s", string(e))
}

// ErrIncompatibleComposition is returned when two trees cannot be combined
type ErrIncompatibleComposition struct {
	ID  string
	Msg string
}

func (e ErrIncompatibleComposition) Error() string {
	return fmt.Sprintf("IncompatibleComposition: %s %s", e.ID, e.Msg)
}

// ErrInvalidLayerType is retured if a layer is not a schema base or overlay
type ErrInvalidLayerType string

func (e ErrInvalidLayerType) Error() string {
	return fmt.Sprintf("Invalid layer type: %s", string(e))
}

// ErrNotASchema is retured if a jsonld object is not a schema during schema parsing
type ErrNotASchema string

func (e ErrNotASchema) Error() string {
	return fmt.Sprintf("Not a schema: %s", string(e))
}

// ErrValidation is returned for validation errors
type ErrValidation string

func (e ErrValidation) Error() string {
	return string(e)
}
