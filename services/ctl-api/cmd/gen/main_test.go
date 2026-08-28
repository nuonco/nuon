package main

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestNormalizeBinaryRequestBodies(t *testing.T) {
	doc := &openapi3.T{Paths: openapi3.NewPaths()}
	doc.Paths.Set("/upload", &openapi3.PathItem{Post: &openapi3.Operation{
		RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{Content: openapi3.Content{
			"application/octet-stream": &openapi3.MediaType{Schema: &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}},
		}}},
	}})

	normalizeBinaryRequestBodies(doc)

	media := doc.Paths.Find("/upload").Post.RequestBody.Value.Content.Get("application/octet-stream")
	require.Equal(t, "binary", media.Schema.Value.Format)
}
