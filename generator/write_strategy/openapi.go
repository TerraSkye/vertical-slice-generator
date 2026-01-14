package write_strategy

import (
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

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
