package template

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gosimple/slug"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

type openApiTemplate struct {
	info *GenerationInfo
	doc  *openapi3.T
}

func NewOpenApiTemplate(info *GenerationInfo) Template {
	return &openApiTemplate{
		info: info,
	}
}

func (t *openApiTemplate) Render(ctx context.Context) write_strategy.Renderer {

	// Initialize components if not present
	if t.doc.Components == nil {
		t.doc.Components = &openapi3.Components{}
	}
	if t.doc.Components.Schemas == nil {

		t.doc.Components.Schemas = openapi3.Schemas{}
	}
	if t.doc.Paths == nil {
		t.doc.Paths = openapi3.NewPaths()
	}

	if _, exists := t.doc.Components.Schemas["ResponseBody"]; !exists {
		t.doc.Components.Schemas["ResponseBody"] = ResponseBodySchema()
	}
	if _, exists := t.doc.Components.Schemas["InputBody"]; !exists {
		t.doc.Components.Schemas["InputBody"] = InputBodySchema()
	}
	if _, exists := t.doc.Components.Schemas["Error"]; !exists {
		t.doc.Components.Schemas["Error"] = ErrorSchema()
	}
	if _, exists := t.doc.Components.Schemas["AcceptedResponseBody"]; !exists {
		t.doc.Components.Schemas["AcceptedResponseBody"] = AcceptedResponseBodySchema()
	}
	if _, exists := t.doc.Components.Schemas["HandleCommandResponse"]; !exists {
		t.doc.Components.Schemas["HandleCommandResponse"] = HandledCommandResponseSchema()
	}

	//return &OpenAPIRenderer{t.doc}
	for _, command := range t.info.Slice.Commands {

		schema := ConvertFieldsToModel(eventmodel.Fields(command.Fields).DataAttributes())
		commandTitle := eventmodel.ProcessTitle(command.Title)
		commandIdAttributes := eventmodel.Fields(command.Fields).IDAttributes()

		// Add schema to components
		requestSchemaName := commandTitle + "Request"
		t.doc.Components.Schemas[requestSchemaName] = openapi3.NewSchemaRef("", schema)

		// Create operation
		operation := openapi3.NewOperation()
		operation.Tags = []string{command.Aggregate}
		operation.Description = command.Description
		operation.OperationID = commandTitle

		// Add path parameters if there are ID attributes
		if len(commandIdAttributes) > 0 {
			operation.Parameters = openapi3.Parameters{}
			for _, idAttr := range commandIdAttributes {
				param := &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        idAttr.Name,
						In:          "path",
						Required:    true,
						Description: "ID parameter for " + idAttr.Name,
						Schema:      getSchemaForFieldType(idAttr.Type),
					},
				}
				operation.Parameters = append(operation.Parameters, param)
			}
		}

		// Add request body
		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								AllOf: openapi3.SchemaRefs{
									{Ref: "#/components/schemas/InputBody"}, // base schema
									{Value: func() *openapi3.Schema { // inline override
										s := openapi3.NewObjectSchema()
										s.Properties = map[string]*openapi3.SchemaRef{
											"data": {Ref: "#/components/schemas/" + requestSchemaName},
										}
										return s
									}()},
								},
							},
						},
					},
				},
			},
		}

		// Add response
		operation.Responses = openapi3.NewResponses(
			openapi3.WithStatus(201, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: ptr("Created successfully"),
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									AllOf: openapi3.SchemaRefs{
										{Ref: "#/components/schemas/ResponseBody"}, // base schema
										{Value: func() *openapi3.Schema { // inline override
											s := openapi3.NewObjectSchema()
											s.Properties = map[string]*openapi3.SchemaRef{
												"data": {Ref: "#/components/schemas/HandleCommandResponse"},
											}
											return s
										}()},
									},
								},
							},
						},
					},
				},
			}),
		)

		// Build path with parameters
		path := buildPathWithParams(fmt.Sprintf("/api/debug/%s/%s", strings.ToLower(t.info.Model.CodeGen.Domain), slug.Make(strings.ReplaceAll(t.info.Slice.Title, "slice:", ""))), commandIdAttributes)

		t.doc.AddOperation(path, "POST", operation)

	}

	// Add schema to components
	//schemaName := readModelTitle
	//t.doc.Components.Schemas[schemaName] = openapi3.NewSchemaRef("", schema)

	for _, readModel := range t.info.Slice.Readmodels {
		schema := ConvertFieldsToModel(readModel.Fields)
		readModelTitle := eventmodel.ProcessTitle(readModel.Title)
		readModelIdAttributes := eventmodel.Fields(readModel.Fields).IDAttributes()

		if readModel.ListElement {

		}

		// Add schema to components
		requestSchemaName := readModelTitle + "ReadModel"
		t.doc.Components.Schemas[requestSchemaName] = openapi3.NewSchemaRef("", schema)

		// Create operation
		operation := openapi3.NewOperation()
		operation.Tags = []string{readModel.Aggregate}
		operation.Description = readModel.Description
		operation.OperationID = readModelTitle

		// Add path parameters if there are ID attributes
		if len(readModelIdAttributes) > 0 {
			operation.Parameters = openapi3.Parameters{}
			for _, idAttr := range readModelIdAttributes {
				param := &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        idAttr.Name,
						In:          "path",
						Required:    true,
						Description: "ID parameter for " + idAttr.Name,
						Schema:      getSchemaForFieldType(idAttr.Type),
					},
				}
				operation.Parameters = append(operation.Parameters, param)
			}
		}

		//TODO Readmodel lists
		//TODO add pagination for lists
		operation.Responses = openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: ptr(fmt.Sprintf("The requested %s", readModelTitle)),
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									AllOf: openapi3.SchemaRefs{
										{Ref: "#/components/schemas/ResponseBody"}, // base schema
										{Value: func() *openapi3.Schema { // inline override
											s := openapi3.NewObjectSchema()
											if readModel.ListElement {
												s.Properties = map[string]*openapi3.SchemaRef{
													"data": {
														Extensions: nil,
														Origin:     nil,
														Ref:        "",
														Value: func() *openapi3.Schema {
															s2 := openapi3.NewArraySchema()

															s2.Items = &openapi3.SchemaRef{
																Ref: "#/components/schemas/" + requestSchemaName,
															}
															return s2
														}(),
													},
												}
											} else {
												s.Properties = map[string]*openapi3.SchemaRef{
													"data": {Ref: "#/components/schemas/" + requestSchemaName},
												}
											}

											return s
										}()},
									},
								},
							},
						},
					},
				},
			}),
		)

		// Build path with parameters
		path := buildPathWithParams(fmt.Sprintf("/api/debug/%s/%s", strings.ToLower(t.info.Model.CodeGen.Domain), slug.Make(strings.ReplaceAll(t.info.Slice.Title, "slice:", ""))), readModelIdAttributes)

		t.doc.AddOperation(path, "GET", operation)

	}
	for _, screen := range t.info.Slice.Screens {
		if idAttributes := eventmodel.Fields(screen.Fields).IDAttributes(); len(idAttributes) > 0 {

		}

		schema := ConvertFieldsToModel(screen.Fields)
		readModelTitle := eventmodel.ProcessTitle(screen.Title)
		readModelIdAttributes := eventmodel.Fields(screen.Fields).IDAttributes()

		if screen.ListElement {

		}

		// Add schema to components
		requestSchemaName := readModelTitle + "ReadModel"
		t.doc.Components.Schemas[requestSchemaName] = openapi3.NewSchemaRef("", schema)

		// Create operation
		operation := openapi3.NewOperation()
		operation.Tags = []string{screen.Aggregate}
		operation.Description = screen.Description
		operation.OperationID = readModelTitle

		// Add path parameters if there are ID attributes
		if len(readModelIdAttributes) > 0 {
			operation.Parameters = openapi3.Parameters{}
			for _, idAttr := range readModelIdAttributes {
				param := &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        idAttr.Name,
						In:          "path",
						Required:    true,
						Description: "ID parameter for " + idAttr.Name,
						Schema:      getSchemaForFieldType(idAttr.Type),
					},
				}
				operation.Parameters = append(operation.Parameters, param)
			}
		}

		//TODO Readmodel lists
		//TODO add pagination for lists
		operation.Responses = openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: ptr(fmt.Sprintf("The requested %s", readModelTitle)),
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									AllOf: openapi3.SchemaRefs{
										{Ref: "#/components/schemas/ResponseBody"}, // base schema
										{Value: func() *openapi3.Schema { // inline override
											s := openapi3.NewObjectSchema()
											if screen.ListElement {
												s.Properties = map[string]*openapi3.SchemaRef{
													"data": {
														Extensions: nil,
														Origin:     nil,
														Ref:        "",
														Value: func() *openapi3.Schema {
															s2 := openapi3.NewArraySchema()

															s2.Items = &openapi3.SchemaRef{
																Ref: "#/components/schemas/" + requestSchemaName,
															}
															return s2
														}(),
													},
												}
											} else {
												s.Properties = map[string]*openapi3.SchemaRef{
													"data": {Ref: "#/components/schemas/" + requestSchemaName},
												}
											}

											return s
										}()},
									},
								},
							},
						},
					},
				},
			}),
		)

		//var url string
		//
		//if screen.APIEndpoint != "" {
		//	url = screen.APIEndpoint
		//} else {
		//	url = fmt.Sprintf("/api/debug/%s/{id}", strings.ToLower(t.info.Model.CodeGen.Domain))
		//}
		//
		//id := eventmodel.Fields(screen.Fields).IDAttributes()
		//
		//schema := ConvertFieldsToModel(eventmodel.Fields(screen.Fields).DataAttributes())
		//commandTitle := eventmodel.ProcessTitle(screen.Title)
		//commandIdAttributes := eventmodel.Fields(screen.Fields).IDAttributes()
		path := buildPathWithParams(fmt.Sprintf("/api/%s/%s", strings.ToLower(t.info.Model.CodeGen.Domain), slug.Make(strings.ReplaceAll(t.info.Slice.Title, "slice:", ""))), readModelIdAttributes)

		t.doc.AddOperation(path, "GET", operation)

	}

	// Build path with parameters
	//path := buildPathWithParams(url, commandIdAttributes)

	//generate new List endpoints
	//readModelTitle := eventmodel.ProcessTitle(readModel.Title)
	//readModelIdAttributes := eventmodel.Fields(readModel.Fields).IDAttributes()
	//
	//schema := ConvertFieldsToModel(readmodel.Fields)
	//
	//if readmodel.ListElement {
	//
	//	operation := openapi3.NewOperation()
	//	operation.Tags = []string{readmodel.Aggregate}
	//	operation.AddResponse(200, openapi3.NewResponse().WithContent(openapi3.NewContentWithJSONSchema(schema)))
	//	operation.Description = readmodel.Description
	//	operation.OperationID = readModelTitle
	//
	//	t.doc.AddOperation("/"+eventmodel.Slugify(readModelTitle), "GET", operation)
	//
	//} else if len(readModelIdAttributes) <= 1 {
	//
	//} else {
	//
	//}

	return &write_strategy.OpenAPIRenderer{t.doc}
}

func (t *openApiTemplate) DefaultPath() string {
	return "/../openapi.yml"
}

func (t *openApiTemplate) Prepare(ctx context.Context) error {
	if err := statFile(t.info.OutputFilePath, t.DefaultPath()); os.IsNotExist(err) {
		t.doc = &openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:       "My API",
				Version:     "1.0.0",
				Description: "API with Hello endpoint",
			},
		}
	} else {
		//loader := openapi3.Loader{Context: ctx}
		loader := openapi3.NewLoader()

		source := filepath.Join(t.info.OutputFilePath, t.DefaultPath())
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		doc, err := loader.LoadFromData(data)
		if err != nil {
			log.Fatal(err)
		}

		t.doc = doc
	}

	return nil
}

func (t *openApiTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {

	return write_strategy.NewCreateRawFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil

}

// ConvertFieldsToModel converts fields to OpenAPI schema
func ConvertFieldsToModel(fields []eventmodel.Field) *openapi3.Schema {
	model := openapi3.NewObjectSchema()
	required := []string{}

	for _, field := range fields {

		var fieldName = eventmodel.SnakeCase(field.Name)
		var propSchema *openapi3.Schema

		switch field.Type {
		case "String":
			propSchema = openapi3.NewStringSchema()
		case "UUID":
			propSchema = openapi3.NewUUIDSchema()
		case "Boolean":
			propSchema = openapi3.NewBoolSchema()
		case "Double":
			propSchema = openapi3.NewFloat64Schema()
		case "Date":
			propSchema = openapi3.NewInt64Schema()
		case "DateTime":
			propSchema = openapi3.NewInt64Schema()
		case "Long":
			propSchema = openapi3.NewInt64Schema()
		case "Int":
			propSchema = openapi3.NewInt64Schema()
		case "Custom":
			propSchema = ConvertFieldsToModel(field.SubFields)
		default:
			propSchema = openapi3.NewStringSchema()
		}
		//TODO snakecase the field names.

		model = model.WithProperty(fieldName, propSchema)

		if !field.Optional {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		model.Required = required
	}

	return model
}

// Helper function to get schema for field type
func getSchemaForFieldType(fieldType string) *openapi3.SchemaRef {
	switch fieldType {
	case "UUID":
		return openapi3.NewSchemaRef("", openapi3.NewUUIDSchema())
	case "Int":
		return openapi3.NewSchemaRef("", openapi3.NewInt64Schema())
	case "Long":
		return openapi3.NewSchemaRef("", openapi3.NewInt64Schema())
	case "String":
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	default:
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	}
}

// Helper function to build path with parameters
func buildPathWithParams(basePath string, idAttributes []eventmodel.Field) string {
	path := basePath
	for _, attr := range idAttributes {
		path = path + "/{" + attr.Name + "}"
	}
	return path
}

func ResponseBodySchema() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: openapi3.NewObjectSchema().
			WithProperties(map[string]*openapi3.Schema{
				"data":  openapi3.NewObjectSchema(),
				"page":  openapi3.NewObjectSchema(),
				"error": openapi3.NewObjectSchema(),
				"meta":  openapi3.NewObjectSchema(),
			}),
	}
}

func InputBodySchema() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: openapi3.NewObjectSchema().
			WithProperties(map[string]*openapi3.Schema{
				"data": openapi3.NewObjectSchema(),
			}),
	}
}

func ErrorSchema() *openapi3.SchemaRef {
	pointer := openapi3.NewObjectSchema()
	pointer.Properties = map[string]*openapi3.SchemaRef{
		"path": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewStringSchema()
				s.Description = "Pointer to error"
				s.Example = "headers.authorization.bearer"
				return s
			}(),
		},
		"message": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewStringSchema()
				s.Description = "Message that belongs to pointer"
				s.Example = "Invalid or missing"
				return s
			}(),
		},
	}

	detail := openapi3.NewObjectSchema()
	detail.Properties = map[string]*openapi3.SchemaRef{
		"reference": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewStringSchema()
				s.Description = "Internal reference to error code"
				s.Example = "CNT-DF-0002"
				return s
			}(),
		},
		"pointers": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewArraySchema()
				s.Items = &openapi3.SchemaRef{Value: pointer}
				return s
			}(),
		},
	}

	errorObj := openapi3.NewObjectSchema()
	errorObj.Properties = map[string]*openapi3.SchemaRef{
		"code": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewIntegerSchema()
				s.Description = "HTTP statuscode returned by the API"
				s.Example = 400
				return s
			}(),
		},
		"message": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewStringSchema()
				s.Description = "Descriptive message of the error"
				s.Example = "Invalid API key."
				return s
			}(),
		},
		"details": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewArraySchema()
				s.Items = &openapi3.SchemaRef{Value: detail}
				return s
			}(),
		},
	}

	root := openapi3.NewObjectSchema()
	root.AllOf = openapi3.SchemaRefs{
		{Ref: "#/components/schemas/ResponseBody"},
	}
	root.Properties = map[string]*openapi3.SchemaRef{
		"page": {
			Value: func() *openapi3.Schema {
				s := openapi3.NewObjectSchema()
				s.Example = map[string]any{}
				return s
			}(),
		},
		"error": {Value: errorObj},
	}

	return &openapi3.SchemaRef{Value: root}
}

func AcceptedResponseBodySchema() *openapi3.SchemaRef {
	root := openapi3.NewObjectSchema()

	status := openapi3.NewStringSchema()
	status.Description = "The status of the accepted job"
	status.Example = "PENDING"

	url := openapi3.NewStringSchema()
	url.Description = "The URL at which to get the current status of the export"
	url.Example = "https://afosto.io/api/jobs/cef865e5-42dc-4cb7-a486-09319c95fab4"

	id := openapi3.NewStringSchema()
	id.Format = "uuid"
	id.Description = "The id of the started job"
	id.Example = "cef865e5-42dc-4cb7-a486-09319c95fab4"

	tracking := openapi3.NewObjectSchema()
	tracking.Properties = map[string]*openapi3.SchemaRef{
		"url": {Value: url},
		"id":  {Value: id},
	}

	root.Properties = map[string]*openapi3.SchemaRef{
		"status":   {Value: status},
		"tracking": {Value: tracking},
	}

	return &openapi3.SchemaRef{Value: root}
}

func HandledCommandResponseSchema() *openapi3.SchemaRef {
	root := openapi3.NewObjectSchema()
	status := openapi3.NewBoolSchema()

	nextRevision := openapi3.NewIntegerSchema()
	nextRevision.Description = "next revision"

	root.Properties = map[string]*openapi3.SchemaRef{
		"status":        {Value: status},
		"next_revision": {Value: nextRevision},
	}

	root.Required = []string{
		"status",
		"next_revision",
	}

	return &openapi3.SchemaRef{Value: root}
}

func ptr(s string) *string {
	return &s
}
