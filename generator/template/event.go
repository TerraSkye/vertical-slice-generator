package template

import (
	"context"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"github.com/terraskye/vertical-slice-generator/template"
)

type eventTemplate struct {
	info  *GenerationInfo
	event *eventmodel.Event
}

func NewEventTemplate(info *GenerationInfo, event *eventmodel.Event) Template {
	return &eventTemplate{
		info:  info,
		event: event,
	}
}

func (t *eventTemplate) Render(ctx context.Context) write_strategy.Renderer {
	z := NewFile("events")
	z.ImportAlias(PackageEventSourcing, "cqrs")

	z.Var().Id("_").Qual(PackageEventSourcing, "Event").Op("=").Call(Op("*").Id(eventmodel.ProcessTitle(t.event.Title))).Call(Nil())

	z.Func().Id("init").Call().BlockFunc(func(group *Group) {
		group.Qual(PackageEventSourcing, "RegisterEventByType").Call(Op("&").Id(eventmodel.ProcessTitle(t.event.Title)).Block())
	})
	z.Line().Type().Id(eventmodel.ProcessTitle(t.event.Title)).Add(template.FieldsStruct(t.event.Fields, false))

	z.Func().Params(
		Id(strings.ToLower(string(t.event.Title[0]))).Op("*").Id(eventmodel.ProcessTitle(t.event.Title))).Id("AggregateID").Params().Params(String()).Block(
		ReturnFunc(func(group *Group) {

			if idAttributeFields := eventmodel.Fields(t.event.Fields).IDAttributes(); len(idAttributeFields) > 0 {

				v := Id(strings.ToLower(string(t.event.Title[0]))).Dot(eventmodel.ProcessTitle(idAttributeFields[0].Name))

				if idAttributeFields[0].Type == "UUID" {
					// in case of an uuid we need to cast it to a string.
					v.Dot("String").Call()
				}
				group.Add(v)
			} else {
				group.Add(Nil().Comment("TODO could not generate aggregateID from event"))
			}
		}),
	)

	z.Func().Params(
		Id(strings.ToLower(string(t.event.Title[0]))).Op("*").Id(eventmodel.ProcessTitle(t.event.Title))).Id("EventType").Params().Params(String()).Block(
		Return(Qual(PackageEventSourcing, "TypeName")).Call(Id(strings.ToLower(string(t.event.Title[0])))),
	)

	return z
}

func (t *eventTemplate) DefaultPath() string {

	return "../events/" + eventmodel.SnakeCase(t.event.Title) + ".go"
}

func (t *eventTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *eventTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
