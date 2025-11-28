package template

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path/filepath"
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

	return &OpenAPIRenderer{t.doc}
	for i, command := range t.info.Slice.Commands {

		schema := ConvertFieldsToModel(command.Fields)

		//generate a new endpoint to the docs.
		_ = i
		_ = command

		_ = schema
		commandTitle := eventmodel.ProcessTitle(command.Title)

		commandIdAttributes := eventmodel.Fields(command.Fields).IDAttributes()

		_ = commandIdAttributes
		operation := openapi3.NewOperation()
		operation.Tags = []string{command.Aggregate}
		operation.RequestBody = &openapi3.RequestBodyRef{
			//Value: openapi3.NewRequestBody().WithSchema(schema, []string{"application/json"}),
			//Ref: openapi3.NewSchemaRef(commandTitle, schema).Ref,
		}
		operation.Description = command.Description
		operation.OperationID = commandTitle
		openapi3.NewParameters()

		t.doc.AddOperation("/"+eventmodel.Slugify(commandTitle), "POST", operation)

	}

	for _, readmodel := range t.info.Slice.Readmodels {
		//generate new List endpoints
		readModelTitle := eventmodel.ProcessTitle(readmodel.Title)
		readModelIdAttributes := eventmodel.Fields(readmodel.Fields).IDAttributes()

		schema := ConvertFieldsToModel(readmodel.Fields)

		if readmodel.ListElement {

			operation := openapi3.NewOperation()
			operation.Tags = []string{readmodel.Aggregate}
			operation.AddResponse(200, openapi3.NewResponse().WithContent(openapi3.NewContentWithJSONSchema(schema)))
			operation.Description = readmodel.Description
			operation.OperationID = readModelTitle

			t.doc.AddOperation("/"+eventmodel.Slugify(readModelTitle), "GET", operation)

		} else if len(readModelIdAttributes) <= 1 {

		} else {

		}

	}
	//TODO add endpoints
	return &OpenAPIRenderer{t.doc}
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
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromFile(filepath.Join(t.info.OutputFilePath, t.DefaultPath()))
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

// Wrapper struct
type OpenAPIRenderer struct {
	Doc *openapi3.T
}

// Implements the Render interface
func (r *OpenAPIRenderer) Render(w io.Writer) error {
	// Marshal to YAML node
	node, err := r.Doc.MarshalYAML()
	if err != nil {
		return err
	}

	// Use yaml.Encoder to write node to the writer
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(node)
}

func ConvertFieldsToModel(fields []eventmodel.Field) *openapi3.Schema {

	model := openapi3.NewObjectSchema()

	for _, field := range fields {
		switch field.Type {
		case "String":
			model = model.WithProperty(field.Name, openapi3.NewStringSchema())
		case "UUID":
			model = model.WithProperty(field.Name, openapi3.NewStringSchema().WithFormat("uuid"))
		case "Boolean":
			model = model.WithProperty(field.Name, openapi3.NewBoolSchema())
		case "Double":
			model = model.WithProperty(field.Name, openapi3.NewFloat64Schema())
		case "Date":
			model = model.WithProperty(field.Name, openapi3.NewInt64Schema())
		case "DateTime":
			model = model.WithProperty(field.Name, openapi3.NewInt64Schema())
		case "Long":
			model = model.WithProperty(field.Name, openapi3.NewInt64Schema())
		case "Int":
			model = model.WithProperty(field.Name, openapi3.NewInt64Schema())
		case "Custom":
			model = model.WithProperty(field.Name, ConvertFieldsToModel(field.SubFields))
		}
	}

	return model

}
