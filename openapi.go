package main

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// JSONSchemaDescriptor struct.
type JSONSchemaDescriptor struct {
	Type       string                           `json:"type"`
	Enum       []string                         `json:"enum,omitempty"`
	Minimum    int64                            `json:"minimum,omitempty"`
	Maximum    int64                            `json:"maximum,omitempty"`
	Properties map[string]*JSONSchemaDescriptor `json:"properties,omitempty"`
}

func getJSONSchemaFromType(st reflect.Type) (*JSONSchemaDescriptor, error) {
	switch st.Kind() {
	case reflect.Bool:
		return &JSONSchemaDescriptor{
			Type: "bool",
		}, nil
	case reflect.Int8:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: -128,
			Maximum: 127,
		}, nil
	case reflect.Int16:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: -32768,
			Maximum: 32767,
		}, nil
	case reflect.Int, reflect.Int32:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: -2147483648,
			Maximum: 2147483647,
		}, nil
	case reflect.Int64:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: -9223372036854775808,
			Maximum: 9223372036854775807,
		}, nil
	case reflect.Uint8:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: 0,
			Maximum: 255,
		}, nil
	case reflect.Uint16:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: 0,
			Maximum: 65535,
		}, nil
	case reflect.Uint, reflect.Uint32:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: 0,
			Maximum: 4294967295,
		}, nil
	case reflect.Uint64:
		return &JSONSchemaDescriptor{
			Type:    "integer",
			Minimum: 0,
		}, nil
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.String:
		return &JSONSchemaDescriptor{
			Type: "string",
		}, nil
	case reflect.Struct:
		props := map[string]*JSONSchemaDescriptor{}
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)

			tag := field.Tag
			jsonTagProperties := strings.SplitN(tag.Get("json"), ",", 2)

			fieldName := field.Name
			if len(jsonTagProperties) > 0 && jsonTagProperties[0] != "" {
				fieldName = jsonTagProperties[0]
			}

			if !unicode.IsUpper([]rune(fieldName)[0]) {
				continue
			}

			f, err := getJSONSchemaFromType(field.Type)
			if err != nil {
				return nil, err
			}
			props[fieldName] = f
		}

		aa := JSONSchemaDescriptor{
			Type:       "object",
			Properties: props,
		}
		return &aa, nil
	default:
		// OK
	}
	return nil, fmt.Errorf("implement me! %s", st.Kind())
}

// GetJSONSchema returns the jsonschema description of a interface.
func GetJSONSchema(s interface{}) (*JSONSchemaDescriptor, error) {
	st := reflect.TypeOf(s)
	return getJSONSchemaFromType(st)
}

// OpenAPISpec struct.
type OpenAPISpec struct {
	OpenAPI string           `json:"openapi"`
	Info    *InfoSpec        `json:"info"`
	Servers []*OpenAPIServer `json:"servers,omitempty"`
	Paths   *OpenAPIPaths    `json:"paths,omitempty"`
}

// New returns a new OpenApiSpec
func New(title, version string) *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &InfoSpec{
			Title:   title,
			Version: version,
		},
	}
}

// InfoSpec struct.
type InfoSpec struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPIServer struct.
type OpenAPIServer struct {
	URL         string `json:"url,omitempty"`
	Description string `json:"desciption,omitempty"`
}

// OpenAPIPaths struct.
type OpenAPIPaths (map[string]map[string]*OpenAPIAPI)

// OpenAPIAPI struct.
type OpenAPIAPI struct {
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description"`
	Responses   map[int]*OpenAPIResponse `json:"responses,omitempty"`
}

// OpenAPIResponse struct.
type OpenAPIResponse struct {
	Description string                             `json:"description"`
	Content     map[string]*OpenAPIResponseContent `json:"content,omitempty"`
}

// OpenAPIResponseContent struct.
type OpenAPIResponseContent struct {
	Schema *JSONSchemaDescriptor `json:"schema,omitempty"`
}

// AddJSONEndpoint adds a new endpoint to swagger
func (oas *OpenAPISpec) AddJSONEndpoint(method, path string, query *QueryDescriptor) Endpoint {
	method = strings.ToLower(method)
	if oas.Paths == nil {
		oas.Paths = &OpenAPIPaths{}
	}
	if (*oas.Paths)[path] == nil {
		(*oas.Paths)[path] = map[string]*OpenAPIAPI{}
	}
	if (*oas.Paths)[path][method] == nil {
		(*oas.Paths)[path][method] = &OpenAPIAPI{}
	}
	if (*oas.Paths)[path][method].Responses == nil {
		(*oas.Paths)[path][method].Responses = map[int]*OpenAPIResponse{}
	}

	return Endpoint{responses: &(*oas.Paths)[path][method].Responses}
}

// Endpoint struct.
type Endpoint struct {
	responses *map[int]*OpenAPIResponse
}

// AddResponse adds to endpoint a JSON response
func (endpoint Endpoint) AddResponse(statusCode int, returnBody *JSONSchemaDescriptor) {
	if (*endpoint.responses)[statusCode] == nil {
		(*endpoint.responses)[statusCode] = &OpenAPIResponse{}
	}
	if (*endpoint.responses)[statusCode].Content == nil {
		(*endpoint.responses)[statusCode].Content = map[string]*OpenAPIResponseContent{}
	}

	(*endpoint.responses)[statusCode].Content["application/json"] = &OpenAPIResponseContent{
		Schema: returnBody,
	}
}

// QueryDescriptor struct.
type QueryDescriptor struct {
}
