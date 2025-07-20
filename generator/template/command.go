package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"github.com/terraskye/vertical-slice-generator/template"
	"strings"
)

type commandTemplate struct {
	info    *GenerationInfo
	command *eventmodel.Command
}

func NewCommandTemplate(info *GenerationInfo, command *eventmodel.Command) Template {
	return &commandTemplate{
		info:    info,
		command: command,
	}
}

func (t *commandTemplate) Render(ctx context.Context) write_strategy.Renderer {
	z := NewFile("domain")
	z.ImportAlias(PackageEventSourcing, "cqrs")

	z.Var().Id("_").Qual(PackageEventSourcing, "Command").Op("=").Call(Op("*").Id(eventmodel.ProcessTitle(t.command.Title))).Call(Nil())

	z.Line().Type().Id(eventmodel.ProcessTitle(t.command.Title)).Add(template.FieldsStruct(t.command.Fields))

	z.Func().Params(
		Id(strings.ToLower(string(t.command.Title[0]))).Op("*").Id(eventmodel.ProcessTitle(t.command.Title))).Id("AggregateID").Params().Params(String()).Block(
		ReturnFunc(func(group *Group) {

			if idAttributeFields := eventmodel.Fields(t.command.Fields).IDAttributes(); len(idAttributeFields) > 0 {

				v := Id(strings.ToLower(string(t.command.Title[0]))).Dot(eventmodel.ProcessTitle(idAttributeFields[0].Name))

				if idAttributeFields[0].Type == "UUID" {
					// in case of an uuid we need to cast it to a string.
					v.Dot("String").Call()
				}
				group.Add(v)
			} else {
				group.Add(Lit("").Comment("TODO generation could not decide the aggregateID"))
			}
		}),
	)

	return z
}

func (t *commandTemplate) DefaultPath() string {
	return "domain/commands/" + eventmodel.ScreenTitle(t.info.Slice.Commands[0].Title) + ".go"
}

func (t *commandTemplate) Prepare(ctx context.Context) error {

	//t.command = &t.info.Slice.Commands[0]
	return nil
}

func (t *commandTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
