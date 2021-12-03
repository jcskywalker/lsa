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

package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

const X_LS = "x-ls"

type ErrCyclicSchema struct {
	Loop []*jsonschema.Schema
}

func (e ErrCyclicSchema) Error() string {
	items := make([]string, 0)
	for _, x := range e.Loop {
		items = append(items, x.Location)
	}
	return "Cycle:" + strings.Join(items, " ")
}

// Entity gives an entity name to a location in a schema
type Entity struct {
	Name       string `json:"name"`
	Ref        string `json:"ref"`
	ID         string `json:"id"`
	SchemaName string `json:"schema"`
}

// GetEntityName returns the ID if name is empty
func (e Entity) GetEntityName() string {
	if len(e.Name) == 0 {
		return e.ID
	}
	return e.Name
}

// CompiledEntity contains the JSON schema for the entity
type CompiledEntity struct {
	Entity
	Schema *jsonschema.Schema
}

// CompileAndImport compiles the given entities and imports them as layers
func CompileAndImport(entities []Entity, typeTerm string) ([]ImportedEntity, error) {
	compiled, err := Compile(entities)
	if err != nil {
		return nil, err
	}
	return Import(compiled, typeTerm)
}

// Compile all entities as a single unit.
func Compile(entities []Entity) ([]CompiledEntity, error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	return CompileWith(compiler, entities)
}

// The meta-schema for annotations
var annotationsMeta = jsonschema.MustCompileString("annotations.json", `{}`)

type annotationsCompiler struct{}

type annotationExtSchema map[string]*ls.PropertyValue

func (annotationExtSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	return nil
}

func (annotationsCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if ext, ok := m[X_LS]; ok {
		if extMap, ok := ext.(map[string]interface{}); ok {
			propertyMap := make(map[string]*ls.PropertyValue)
			for k, v := range extMap {
				switch value := v.(type) {
				case string, json.Number, float64, bool:
					propertyMap[k] = ls.StringPropertyValue(fmt.Sprint(v))
				case []interface{}:
					arr := make([]string, 0, len(value))
					for _, val := range value {
						switch val.(type) {
						case string, json.Number, float64, bool:
							arr = append(arr, fmt.Sprint(val))
						default:
							return nil, fmt.Errorf("Invalid "+X_LS+" value: %s=%v", k, v)
						}
					}
					propertyMap[k] = ls.StringSlicePropertyValue(arr)
				default:
					return nil, fmt.Errorf("Invalid "+X_LS+" value: %s=%v", k, v)
				}
			}
			return annotationExtSchema(propertyMap), nil
		} else if ext != nil {
			return nil, fmt.Errorf(X_LS + " is not an object")
		}
	}
	// nothing to compile, return nil
	return nil, nil
}

// CompileWith compiles all entities as a single unit using the given compiler
func CompileWith(compiler *jsonschema.Compiler, entities []Entity) ([]CompiledEntity, error) {
	ret := make([]CompiledEntity, 0, len(entities))
	compiler.RegisterExtension(X_LS, annotationsMeta, annotationsCompiler{})
	for _, e := range entities {
		sch, err := compiler.Compile(e.Ref)
		if err != nil {
			return nil, fmt.Errorf("During %s: %w", e.GetEntityName(), err)
		}
		ret = append(ret, CompiledEntity{Entity: e, Schema: sch})
	}
	return ret, nil
}

type importContext struct {
	entities      []CompiledEntity
	loop          []*jsonschema.Schema
	currentEntity *CompiledEntity
}

func (ctx *importContext) checkLoopAndPush(sch *jsonschema.Schema) bool {
	for _, x := range ctx.loop {
		if sch == x {
			return true
		}
	}
	ctx.loop = append(ctx.loop, sch)
	return false
}

func (ctx *importContext) pop() {
	ctx.loop = ctx.loop[:len(ctx.loop)-1]
}

func (ctx *importContext) findEntity(sch *jsonschema.Schema) *CompiledEntity {
	for i := range ctx.entities {
		if ctx.entities[i].Schema == sch {
			return &ctx.entities[i]
		}
	}
	return nil
}

type ImportedEntity struct {
	Entity CompiledEntity

	Layer *ls.Layer `json:"-"`
}

// Import a JSON schema or overlay
//
// A JSON schema may include many object definitions. This import
// algorithm creates a schema for each entity.
//
// typeTerm should be either ls.SchemaTerm or ls.OverlayTerm
func Import(entities []CompiledEntity, typeTerm string) ([]ImportedEntity, error) {
	ctx := importContext{entities: entities}
	ret := make([]ImportedEntity, 0, len(ctx.entities))
	for i := range ctx.entities {
		ctx.currentEntity = &ctx.entities[i]

		s := schemaProperty{}
		ctx.loop = make([]*jsonschema.Schema, 0)
		if err := importSchema(&ctx, &s, ctx.currentEntity.Schema); err != nil {
			return nil, err
		}

		imported := ImportedEntity{}
		imported.Entity = ctx.entities[i]
		imported.Layer = ls.NewLayer()
		imported.Layer.SetLayerType(typeTerm)
		imported.Layer.SetID(ctx.currentEntity.ID)
		rootNode := imported.Layer.NewNode(ctx.currentEntity.ID)
		ls.Connect(imported.Layer.GetLayerInfoNode(), rootNode, ls.LayerRootTerm)
		if s.object != nil {
			rootNode.GetTypes().Add(ls.AttributeTypes.Object)
			nodes := s.object.itr(ctx.currentEntity.ID, nil, imported.Layer)
			for _, node := range nodes {
				ls.Connect(rootNode, node, ls.LayerTerms.Attributes)
			}
		} else if s.array != nil {
			rootNode.GetTypes().Add(ls.AttributeTypes.Array)
			node := s.array.itr(ctx.currentEntity.ID, nil, imported.Layer)
			ls.Connect(rootNode, node, ls.LayerTerms.ArrayItems)
		} else {
			rootNode.GetTypes().Add(ls.AttributeTypes.Value)
			buildSchemaAttrs(ctx.currentEntity.ID, nil, s, imported.Layer, rootNode)
		}

		ret = append(ret, imported)
	}
	return ret, nil
}

func importSchema(ctx *importContext, target *schemaProperty, sch *jsonschema.Schema) error {
	if ctx.checkLoopAndPush(sch) {
		return ErrCyclicSchema{Loop: ctx.loop}
	}
	defer ctx.pop()

	switch {
	case sch.Ref != nil:
		ref := ctx.findEntity(sch.Ref)
		if ref != nil {
			target.reference = ref.ID
			return nil
		}
		return importSchema(ctx, target, sch.Ref)

	case len(sch.AllOf) > 0:
		for _, x := range sch.AllOf {
			prop := schemaProperty{}
			if err := importSchema(ctx, &prop, x); err != nil {
				return err
			}
			target.allOf = append(target.allOf, prop)
		}

	case len(sch.AnyOf) > 0:
		for _, x := range sch.AnyOf {
			prop := schemaProperty{}
			if err := importSchema(ctx, &prop, x); err != nil {
				return err
			}
			target.allOf = append(target.allOf, prop)
		}

	case len(sch.OneOf) > 0:
		for _, x := range sch.OneOf {
			prop := schemaProperty{}
			if err := importSchema(ctx, &prop, x); err != nil {
				return err
			}
			target.oneOf = append(target.oneOf, prop)
		}

	case len(sch.Properties) > 0:
		target.object = &objectSchema{properties: make(map[string]schemaProperty), required: sch.Required}
		for k, v := range sch.Properties {
			val := schemaProperty{key: k}
			err := importSchema(ctx, &val, v)
			if err != nil {
				return err
			}
			target.object.properties[k] = val
		}
		// TODO: patternProperties, etc
	case sch.Items != nil:
		target.array = &arraySchema{}
		if itemSchema, ok := sch.Items.(*jsonschema.Schema); ok {
			err := importSchema(ctx, &target.array.items, itemSchema)
			if err != nil {
				return err
			}
		} else {
			panic("Multiple item schemas not supported")
		}
		// TODO: additionalItems, etc
	default:
		if len(sch.Types) > 0 {
			target.typ = sch.Types
		}
		target.format = sch.Format
		if len(sch.Enum) > 0 {
			target.enum = sch.Enum
		}
		if len(sch.Constant) > 0 {
			target.enum = sch.Constant
		}
		if sch.Pattern != nil {
			target.pattern = sch.Pattern.String()
		}
		if len(sch.Description) > 0 {
			target.description = sch.Description
		}
		if sch.Default != nil {
			s := fmt.Sprint(sch.Default)
			target.defaultValue = &s
		}
		if ext, ok := sch.Extensions[X_LS]; ok {
			mext, _ := ext.(annotationExtSchema)
			if len(mext) > 0 {
				target.annotations = map[string]*ls.PropertyValue(mext)
			}
		}
	}

	return nil
}
