package write_strategy

import (
	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Wrapper struct
type OpenAPIRenderer struct {
	Doc *openapi3.T
}

// Implements the Render interface
func (r *OpenAPIRenderer) Render(w io.Writer) error {
	for _, s := range r.Doc.Paths.InMatchingOrder() {
		logrus.Warn(s)
	}
	// Marshal to YAML node
	node, err := r.Doc.MarshalYAML()
	if err != nil {
		return err
	}

	// Use yaml.Encoder to write node to the writer
	enc := yaml.NewEncoder(w)
	//defer enc.Close()
	if err := enc.Encode(node); err != nil {
		return err
	}
	return enc.Close()

	return err
}
