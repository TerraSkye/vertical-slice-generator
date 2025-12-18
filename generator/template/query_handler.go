package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"github.com/terraskye/vertical-slice-generator/template"
)

type queryHandlerTemplate struct {
	info      *GenerationInfo
	readmodel *eventmodel.Readmodel
}

func NewQueryHandlerTemplate(info *GenerationInfo, readmodel *eventmodel.Readmodel) Template {
	return &queryHandlerTemplate{
		info:      info,
		readmodel: readmodel,
	}
}

func (t *queryHandlerTemplate) Render(ctx context.Context) write_strategy.Renderer {
	z := NewFile("domain")
	z.ImportAlias(PackageEventSourcing, "cqrs")

	idAttributes := eventmodel.Fields(t.readmodel.Fields).IDAttributes()

	if len(idAttributes) <= 1 {

	} else {

	}

	z.Line().Type().Id("Query").Add(template.FieldsStruct(eventmodel.Fields(t.readmodel.Fields).IDAttributes(), false))

	z.Line().Add(Func().Params(Id("q").Id("Query")).Id("ID").Params().Params(Index().Byte()).Block(Return(Nil())))

	z.Line().Type().Id("QueryHandler").Struct(
		Id("repository").Id(eventmodel.ProcessTitle(t.readmodel.Title) + "Repository"),
	)

	z.Line().Func().Id("NewQueryHandler").Params(
		Id("repository").Id(eventmodel.ProcessTitle(t.readmodel.Title) + "Repository")).
		Params(
			Qual(PackageEventSourcing, "QueryHandler").Types(Id("Query"), Op("*").Id(eventmodel.ProcessTitle(t.readmodel.Title))),
		).BlockFunc(func(group *Group) {
		group.Return(Op("&").Id("QueryHandler").Block(
			Dict{
				Id("repository"): Id("repository").Op(","),
			}))

	})

	z.Line().Add(
		Func().Params(
			Id("q").Op("*").Id("QueryHandler")).Id("HandleQuery").Params(
			// function params
			Id("ctx").Qual("context", "Context"),
			Id("qry").Id("Query"),
		),
		//TODO generate the repositry query
		// https://github.com/dilgerma/nebulit-code-generators/blob/f7a56a6857c8ac68b78635e81f766d06489631e6/generators/axon/slices/index.js#L241
	).Params(Op("*").Id(eventmodel.ProcessTitle(t.readmodel.Title)), Error()).Block(Return(Nil(), Nil()))

	return z
}

func (t *queryHandlerTemplate) DefaultPath() string {
	return "slices/" + eventmodel.SliceTitle(t.info.Slice.Title) + "/query.go"
}

func (t *queryHandlerTemplate) Prepare(ctx context.Context) error {

	//TODO implement survey to ask if it should be a live report or not.
	//var live bool
	//survey.AskOne(&survey.Input{
	//	Message: "Is it a live model",
	//}, &live)
	//t.command = &t.info.Slice.Commands[0]
	return nil
}

func (t *queryHandlerTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
