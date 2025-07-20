package template

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
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

	for i, command := range t.info.Slice.Commands {
		//generate a new endpoint to the docs.
	}

	for i, readmodel := range t.info.Slice.Readmodels {
		//generate new List endpoints
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
			Paths: &openapi3.Paths{},
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
