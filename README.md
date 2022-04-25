[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/lsa?status.svg)](https://godoc.org/github.com/cloudprivacylabs/lsa)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/lsa)](https://goreportcard.com/report/github.com/cloudprivacylabs/lsa)
[![Build Status](https://github.com/cloudprivacylabs/lsa/actions/workflows/CI.yml/badge.svg?branch=main)](https://github.com/cloudprivacylabs/lsa/actions/workflows/CI.yml)
# Layered Schemas

Layered Schema Architecture (LSA) enables semantic interoperability
between heterogeneous systems. LSA uses a schema (such as FHIR schemas
for health data) to define data structures, and overlays to add
semantics, contextual metadata, and processing directives. A schema
variant is the composition of the schema base and use-case specific
overlays. Different schema variants can be used to ingest data from or
export data to disparate systems. Each variant encodes source specific
metadata and rules to ingest data into a knowledge graph, or target
specific metadata and rules to translate the knowledge graph into
another format.

A layered schema defines a mapping from structured data (in JSON, XML,
tabular format) to a labeled property graph. For instance, when a JSON
document is ingested using a layered schema, a labeled property graph
is constructed by combining schema information with the input data
elements. The resulting graph is self-contained, and includes both the
input data and the schema. This allows processing data based on its
semantic attributes instead of its attributes names or other
structural properties. For example, it is possible to use layered
schema to ingest health data, mark certain attributes as
"identifiable", and then, remove all attributes that are marked as
such. The algorithm to remove "identifiable" elements can be written
without any knowledge of the input structure. Same algorithm can be
used to de-identify health data as well as marketing data.

The main documentation site for layered schemas is:

https://layeredschemas.org

This Go module contains the reference implementation of the layered
schema specification. It contains a layered schema processor to
slice/compose schemas, import JSON schemas, and ingest data from
JSON/CSV/XML sources.


## Overview

A traditional schema (a JSON schema, an XML schema, etc.) defines one
or more data types. Such schemas are primarily used for documentation,
data validation, and automated code generation. A schema includes
definitions about the shape of the data object such as the list of
fields, nested fields, or constraints on field contents. A layered
schema adds additional interchangeable layers to a traditional schema
so the same schema can be adjusted for different use cases, or can be
processed with different processing directives to handle data
variations. The layered schema architecture provides the framework to
extend traditional schemas using such layers.

### Bundles, schemas, overlays

A *schema* defines a data type by specifying its attributes. The name
of the data type is the *valueType* for the schema.

Below is an example schema for a `Person` data type containing
`firstName` and `lastName`.

```
{
  "@context": "https://lschema.org/ls.json",
  "@id": "https://example.org/Person/schema",
  "@type": "Schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://example.org/Person",
    "attributes": {
      "firstName": {
        "@type": "Value"
      },
      "lastName": {
        "@type": "Value"
     }
   }
 }
}
```

The `Person` object is defined as an `Object` that contains two
`Value` attributes: `firstName` and `lastName`. These are the unique
attribute IDs. An overlay can be defined to add new semantics to these
attributes. The following overlay adds personal data category `Name`
to the `lastName` attribute:

```
{
  "@context": "https://lschema.org/ls.json",
  "@id": "https://example.org/Person/dpv-ovl",
  "@type": "Overlay",
  "valueType": "Person",
  "layer": {
    "lastName": {
      "@type": "Value",
      "http://www.w3.org/ns/dpv#hasPersonalDataCategory": "http://www.w3.org/dpv#Name"
    }
  }
}
```

## Building

Once you clone the repository, you can build the schema compiler using
the Go build system.

```
cd layers
go build
```

This will build the `layers` binary in the current directory.

## Examples

The `examples/` directory contains some example data processing
scenarios.

## Commercial Support

Commercial support for this library is available from Cloud Privacy Labs: support@cloudprivacylabs.com


