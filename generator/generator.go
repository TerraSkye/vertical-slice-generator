package generator

import (
	"context"
	"errors"
	"fmt"
	"github.com/terraskye/vertical-slice-generator/generator/template"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

var (
	EmptyTemplateError = errors.New("empty template")
	EmptyStrategyError = errors.New("empty strategy")
)

type Generator interface {
	Generate() error
}

type GenerationUnit struct {
	template template.Template

	writeStrategy write_strategy.Strategy
	absOutPath    string
}

func NewGenUnit(ctx context.Context, tmpl template.Template, outPath string) (*GenerationUnit, error) {
	err := tmpl.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare error: %v", tmpl.DefaultPath(), err)
	}
	strategy, err := tmpl.ChooseStrategy(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: choose strategy: %v", tmpl.DefaultPath(), err)
	}
	return &GenerationUnit{
		template:      tmpl,
		absOutPath:    outPath,
		writeStrategy: strategy,
	}, nil
}

func (g *GenerationUnit) Generate(ctx context.Context) error {
	if g.template == nil {
		return EmptyTemplateError
	}
	if g.writeStrategy == nil {
		return EmptyStrategyError
	}
	code := g.template.Render(ctx)
	err := g.writeStrategy.Write(code)
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	return nil
}

func (g GenerationUnit) Path() string {
	return g.absOutPath
}
