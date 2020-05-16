package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetJSONSchemaFromInterface(t *testing.T) {
	t.Run("uint8", func(t *testing.T) {
		var s = uint8(3)

		jsonSchema, err := GetJSONSchema(s)
		require.NoError(t, err)

		require.Equal(t, &JSONSchemaDescriptor{
			Type:    "integer",
			Maximum: 255,
		}, jsonSchema)
	})

	t.Run("string", func(t *testing.T) {
		var s = "my-string"

		jsonSchema, err := GetJSONSchema(s)
		require.NoError(t, err)

		require.Equal(t, &JSONSchemaDescriptor{
			Type: "string",
		}, jsonSchema)
	})

	t.Run("simple structure", func(t *testing.T) {
		type SimpleStructure struct {
			Foo string
			Bar uint8
		}
		var s = SimpleStructure{}

		jsonSchema, err := GetJSONSchema(s)
		require.NoError(t, err)

		require.Equal(t, &JSONSchemaDescriptor{
			Type: "object",
			Properties: map[string]*JSONSchemaDescriptor{
				"Foo": {
					Type: "string",
				},
				"Bar": {
					Type:    "integer",
					Maximum: 255,
				},
			},
		}, jsonSchema)
	})

	t.Run("neasting structure", func(t *testing.T) {
		type NestedStructure struct {
			Foo1    string
			Bar1    uint8
			ignored uint16
		}
		type SimpleStructure struct {
			Foo     string `json:"TheName"`
			Bar     NestedStructure
			ignored uint16
		}
		var s = SimpleStructure{}

		jsonSchema, err := GetJSONSchema(s)
		require.NoError(t, err)

		require.Equal(t, &JSONSchemaDescriptor{
			Type: "object",
			Properties: map[string]*JSONSchemaDescriptor{
				"TheName": {
					Type: "string",
				},
				"Bar": {
					Type: "object",
					Properties: map[string]*JSONSchemaDescriptor{
						"Foo1": {
							Type: "string",
						},
						"Bar1": {
							Type:    "integer",
							Maximum: 255,
						},
					},
				},
			},
		}, jsonSchema)
	})
}

func TestAddJSONEndpoint(t *testing.T) {
	type SimpleStructure struct {
		Foo string
		Bar uint8
	}

	t.Run("ok", func(t *testing.T) {
		oas := New("myTitle", "2.2.2")
		ret, err := GetJSONSchema(SimpleStructure{})
		require.NoError(t, err)

		oas.AddJSONEndpoint(http.MethodGet, "/foo", nil).
			AddResponse(http.StatusOK, ret)

		require.Equal(t, &OpenAPISpec{
			OpenAPI: "3.0.0",
			Info: &InfoSpec{
				Title:   "myTitle",
				Version: "2.2.2",
			},
			Paths: &OpenAPIPaths{
				"/foo": {
					"get": {
						Responses: map[int]*OpenAPIResponse{
							200: {
								Content: map[string]*OpenAPIResponseContent{
									"application/json": {
										Schema: ret,
									},
								},
							},
						},
					},
				},
			},
		}, oas)

		definitionAsBytes, err := json.Marshal(oas)
		require.NoError(t, err)
		definitionAsString := string(definitionAsBytes)

		require.Equal(t, `{"openApi":"","info":{},"paths":{"/foo":{"GET":{"responses":{"200":{"content":{"application/json":{"schema":{"type":"object","properties":{"Bar":{"type":"integer","maximum":255},"Foo":{"type":"string"}}}}}}}}}}}`, definitionAsString)
	})
}
