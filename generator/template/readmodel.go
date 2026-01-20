package template

import (
	"context"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"github.com/terraskye/vertical-slice-generator/template"
)

type readModelTemplate struct {
	info      *GenerationInfo
	readmodel *eventmodel.Readmodel
}

func NewReadModelTemplate(info *GenerationInfo, readmodel *eventmodel.Readmodel) Template {
	return &readModelTemplate{
		info:      info,
		readmodel: readmodel,
	}
}

func (t *readModelTemplate) Render(ctx context.Context) write_strategy.Renderer {
	z := NewFile(eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))
	z.ImportAlias(PackageEventSourcing, "cqrs")

	z.Line().Type().Id(eventmodel.ProcessTitle(t.readmodel.Title)).Add(template.FieldsStruct(t.readmodel.Fields, false))

	return z
}

func (t *readModelTemplate) DefaultPath() string {
	return "slices/" + eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")) + "/readmodel.go"
}

func (t *readModelTemplate) Prepare(ctx context.Context) error {

	//TODO implement survey to ask if it should be a live report or not.
	//var live bool
	//survey.AskOne(&survey.Input{
	//	Message: "Is it a live model",
	//}, &live)
	//t.command = &t.info.Slice.Commands[0]
	return nil
}

func (t *readModelTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
