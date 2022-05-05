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
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Bundle defines type names for variants so references can be resolved
type Bundle struct {
	Base      string                   `json:"base" yaml:"base"`
	TypeNames map[string]BundleVariant `json:"typeNames" yaml:"typeNames"`
}

// Merge bundle into b
func (b *Bundle) Merge(bundle Bundle) {
	if b.TypeNames == nil {
		b.TypeNames = make(map[string]BundleVariant)
	}
	for typeName, variant := range bundle.TypeNames {
		existingVariant, ok := b.TypeNames[typeName]
		if !ok {
			b.TypeNames[typeName] = variant
			continue
		}
		existingVariant.Merge(variant)
		b.TypeNames[typeName] = existingVariant
	}
}

func (b *Bundle) ResolveFilenames(dir string) {
	for k, v := range b.TypeNames {
		v.ResolveFilenames(dir)
		b.TypeNames[k] = v
	}
	if len(b.Base) > 0 {
		b.Base = getRelativeFileName(dir, b.Base)
	}
}

type BundleSchemaRef struct {
	Schema     string               `json:"schema,omitempty" yaml:"schema,omitempty"`
	JSONSchema *JSONSchemaReference `json:"jsonSchema" yaml:"jsonSchema"`
}

// Merge resets b with ref if ref is nonempty
func (b *BundleSchemaRef) Merge(ref BundleSchemaRef) {
	if len(ref.Schema) > 0 || ref.JSONSchema != nil {
		*b = ref
	}
}

// GetLayerID returns the layer id for the schema reference
func (ref BundleSchemaRef) GetLayerID() string {
	if ref.JSONSchema != nil {
		return ref.JSONSchema.LayerID
	}
	return ref.Schema
}

func (ref *BundleSchemaRef) ResolveFilenames(dir string) {
	if len(ref.Schema) > 0 {
		ref.Schema = getRelativeFileName(dir, ref.Schema)
	}
	if ref.JSONSchema != nil {
		if len(ref.JSONSchema.Ref) > 0 {
			ref.JSONSchema.Ref = getRelativeFileName(dir, ref.JSONSchema.Ref)
		}
	}
}

type JSONSchemaReference struct {
	// This is here for compatibility with commercial version
	LayerID string `json:"layerId" yaml:"layerId" bson:"layerId"`
	// This is the filename
	Ref       string `json:"ref" yaml:"ref" bson:"ref"`
	Namespace string `json:"namespace" yaml:"namespace" bson:"namespace"`
}

// BundleVariant combines a schema and overlays
type BundleVariant struct {
	BundleSchemaRef
	Overlays []BundleSchemaRef `json:"overlays" yaml:"overlays"`
}

func (b *BundleVariant) ResolveFilenames(dir string) {
	b.BundleSchemaRef.ResolveFilenames(dir)
	for i := range b.Overlays {
		b.Overlays[i].ResolveFilenames(dir)
	}
}

// Merge variant into b
func (b *BundleVariant) Merge(variant BundleVariant) {
	b.BundleSchemaRef.Merge(variant.BundleSchemaRef)
	b.Overlays = append(b.Overlays, variant.Overlays...)
}

// ParseBundle parses a bundle from JSON
func ParseBundle(text string, contentType string) (*Bundle, error) {
	var ret Bundle
	if err := cmdutil.ReadJSON(contentType, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

var DefaultFileLoader = func(s string) (io.ReadCloser, error) {
	obj, err := cmdutil.ReadURL(s)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(obj)), nil
}

func (bundle *Bundle) importJSONSchema(ctx *ls.Context, typeTerm string, importEntities []jsonsch.Entity, fileLoader func(s string) (io.ReadCloser, error)) (map[string]*ls.Layer, error) {
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = fileLoader
	// Import all JSON schemas into a graph
	jsonLayers := make(map[string]*ls.Layer)
	if len(importEntities) > 0 {
		compiled, err := jsonsch.CompileEntitiesWith(compiler, importEntities...)
		if err != nil {
			return nil, err
		}
		g := ls.NewLayerGraph()
		layers, err := jsonsch.BuildEntityGraph(g, typeTerm, jsonsch.LinkRefsByValueType, compiled...)
		if err != nil {
			return nil, err
		}
		for _, layer := range layers {
			jsonLayers[layer.Layer.GetID()] = layer.Layer
		}
	}
	return jsonLayers, nil
}

// GetLayers returns the layers of the bundle keyed by variant type
func (bundle *Bundle) GetLayers(ctx *ls.Context, loader func(s string) (*ls.Layer, error), fileLoader func(string) (io.ReadCloser, error)) (map[string]*ls.Layer, error) {
	// layers keyed by layer id
	layers := make(map[string]*ls.Layer)

	// For JSON-LS schemas, layerId refers to the filename. We use this map to map that filename to loaded layerid
	layerIDMap := make(map[string]string)
	// entities keyed by layer id
	schemaEntities := make(map[string]jsonsch.Entity)
	ovlEntities := make(map[string]jsonsch.Entity)

	processRef := func(variantType string, ref BundleSchemaRef, entitiesMap map[string]jsonsch.Entity) error {
		switch {
		case len(ref.Schema) > 0:
			// ref.LayerID refers to a file
			// Load the layer if not loaded before
			fileName := ref.Schema
			_, loaded := layerIDMap[fileName]
			if loaded {
				break
			}
			layer, err := loader(fileName)
			if err != nil {
				return err
			}
			layerID := layer.GetID()
			if _, exists := layers[layerID]; exists {
				return fmt.Errorf("Duplicate id %s in %s", layerID, ref.Schema)
			}
			layers[layerID] = layer
			layerIDMap[fileName] = layerID

		case ref.JSONSchema != nil:
			_, loaded := entitiesMap[ref.JSONSchema.LayerID]
			if loaded {
				break
			}
			entity := jsonsch.Entity{
				LayerID:       ref.JSONSchema.LayerID,
				ValueType:     variantType,
				Ref:           ref.JSONSchema.Ref,
				AttrNamespace: ref.JSONSchema.Namespace,
			}
			entitiesMap[entity.LayerID] = entity
			layerIDMap[entity.LayerID] = entity.LayerID
		}
		return nil
	}

	// Load all layers, construct entities
	for variantType, variant := range bundle.TypeNames {
		if err := processRef(variantType, variant.BundleSchemaRef, schemaEntities); err != nil {
			return nil, err
		}
		for _, ovl := range variant.Overlays {
			if err := processRef(variantType, ovl, ovlEntities); err != nil {
				return nil, err
			}
		}
	}

	// If there are entities, import them
	// First import schemas, then overlays
	importJson := func(input map[string]jsonsch.Entity, typeTerm string) error {
		importEntities := make([]jsonsch.Entity, 0, len(input))
		for _, entity := range input {
			importEntities = append(importEntities, entity)
		}
		if len(importEntities) > 0 {
			jlayers, err := bundle.importJSONSchema(ctx, typeTerm, importEntities, fileLoader)
			if err != nil {
				return err
			}
			for k, v := range jlayers {
				if _, exists := layers[k]; exists {
					return fmt.Errorf("Multiple definitions for layer %s", k)
				}
				layers[k] = v
			}
		}
		return nil
	}

	if err := importJson(schemaEntities, ls.SchemaTerm); err != nil {
		return nil, err
	}
	if err := importJson(ovlEntities, ls.OverlayTerm); err != nil {
		return nil, err
	}

	resultBundle := ls.BundleByType{}
	for variantType, variant := range bundle.TypeNames {
		sch := layers[layerIDMap[variant.GetLayerID()]]
		ovl := make([]*ls.Layer, 0, len(variant.Overlays))
		for _, o := range variant.Overlays {
			ovl = append(ovl, layers[layerIDMap[o.GetLayerID()]])
		}
		if len(ovl) > 0 {
			sch = sch.Clone()
		}
		_, err := resultBundle.Add(ctx, variantType, sch, ovl...)
		if err != nil {
			return nil, err
		}
	}
	return resultBundle.Variants, nil
}

func loadBundleChain(ctx *ls.Context, file string) (Bundle, error) {
	var b Bundle
	if err := cmdutil.ReadJSONOrYAML(file, &b); err != nil {
		return Bundle{}, fmt.Errorf("While reading %s: %w", file, err)
	}
	b.ResolveFilenames(filepath.Dir(file))
	if len(b.Base) > 0 {
		base, err := loadBundleChain(ctx, b.Base)
		if err != nil {
			return Bundle{}, err
		}
		b.Merge(base)
	}
	return b, nil
}

func LoadBundle(ctx *ls.Context, file []string) (ls.SchemaLoader, error) {
	var bundle Bundle
	for _, f := range file {
		b, err := loadBundleChain(ctx, f)
		if err != nil {
			return nil, fmt.Errorf("While reading %s: %w", f, err)
		}
		bundle.Merge(b)
	}
	items, err := bundle.GetLayers(ctx, func(fname string) (*ls.Layer, error) {
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			return nil, err
		}
		var input interface{}
		err = json.Unmarshal([]byte(data), &input)
		if err != nil {
			return nil, err
		}
		return ls.UnmarshalLayer(input, nil)
	}, DefaultFileLoader)
	if err != nil {
		return nil, err
	}
	b := ls.BundleByType{}
	for k, v := range items {
		b.Add(ctx, k, v)
	}
	return &b, nil
}

// ReadLayers reads layer(s) from jsongraph, jsonld
func ReadLayers(input []byte, interner ls.Interner) ([]*ls.Layer, error) {
	var v interface{}
	err := json.Unmarshal(input, &v)
	if err != nil {
		return nil, err
	}
	// Input is JSON or JSON-LD
	// If input is []interface{}, it must be JSON-LD
	if _, arr := v.([]interface{}); arr {
		l, err := ls.UnmarshalLayer(v, interner)
		if err != nil {
			return nil, err
		}
		return []*ls.Layer{l}, nil
	}
	// If input has "nodes", it is a JSON graph
	if m, ok := v.(map[string]interface{}); ok {
		if _, exists := m["nodes"]; exists {
			target := ls.NewLayerGraph()
			if err := ls.NewJSONMarshaler(interner).Unmarshal(input, target); err != nil {
				return nil, err
			}
			layers := ls.LayersFromGraph(target)
			if len(layers) == 0 {
				return nil, fmt.Errorf("No layers in input")
			}
			return layers, nil
		}
	}
	// Try json-ld
	l, err := ls.UnmarshalLayer(v, interner)
	if err != nil {
		return nil, err
	}
	return []*ls.Layer{l}, nil
}

func getRelativeFileName(dir, fname string) string {
	if filepath.IsAbs(fname) {
		return fname
	}
	return filepath.Join(dir, fname)
}
