package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

type projectorTemplate struct {
	info      *GenerationInfo
	readmodel *eventmodel.Readmodel
}

func NewProjectorTemplate(info *GenerationInfo, readmodel *eventmodel.Readmodel) Template {
	return &projectorTemplate{
		info:      info,
		readmodel: readmodel,
	}
}

func (t *projectorTemplate) Render(ctx context.Context) write_strategy.Renderer {
	z := NewFile("domain")
	z.ImportAlias(PackageEventSourcing, "cqrs")

	eventsPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/events")
	if err != nil {
		panic(err)
	}

	var inboundEvents []*eventmodel.Event

	{
		for _, dependency := range t.readmodel.Dependencies {
			if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
				inboundEvents = append(inboundEvents, t.info.Model.FindEventByID(dependency.ID))

			}
		}
	}

	idAttributes := eventmodel.Fields(t.readmodel.Fields).IDAttributes()

	if len(idAttributes) <= 1 {

	} else {

	}

	z.Line().Type().Id(eventmodel.ProcessTitle(t.readmodel.Title) + "Repository").Interface()

	z.Line().Type().Id("projector").Struct(
		Id("repository").Id(eventmodel.ProcessTitle(t.readmodel.Title) + "Repository"),
	)

	z.Line().Func().Id("NewProjector").Params(Id("repository").Id(eventmodel.ProcessTitle(t.readmodel.Title) + "Repository")).Params(Op("*").Qual(PackageEventSourcing, "EventGroupProcessor")).BlockFunc(func(group *Group) {
		group.Id("p").Op(":=").Op("&").Id("projector").Block(
			Dict{
				Id("repository"): Id("repository").Op(","),
			})

		group.Return(Qual(PackageEventSourcing, "NewEventGroupProcessor").
			Call(ListFunc(func(group *Group) {
				group.Line().Lit(eventmodel.SliceTitle(t.info.Slice.Title))
				for _, event := range inboundEvents {
					group.Line().Qual(PackageEventSourcing, "NewGroupEventHandler").Call(Id("p").Dot("On" + eventmodel.ProcessTitle(event.Title)))
				}
			})))
	})

	for _, event := range inboundEvents {
		z.Line().Add(
			Func().Params(
				Id("p").Op("*").Id("projector")).Id("On"+eventmodel.ProcessTitle(event.Title)).Params(
				// function params
				Id("ctx").Qual("context", "Context"),
				Id("event").Op("*").Qual(eventsPackage, eventmodel.ProcessTitle(event.Title)),
			).Params(Error())).Block(Return(Nil()))
	}

	//z.Line().Type().Id(eventmodel.ReadModelTitle(t.readmodel.Title)).Add(template.FieldsStruct(t.readmodel.Fields))

	//z.Line().Type().Id("Query").Add(template.FieldsStruct(eventmodel.Fields(t.readmodel.Fields).IDAttributes()))

	//z.Func().Params(
	//	Id(strings.ToLower(string(t.command.Title[0]))).Op("*").Id(eventmodel.ProcessTitle(t.command.Title))).Id("AggregateID").Params().Params(String()).Block(
	//	ReturnFunc(func(group *Group) {
	//
	//		if idAttributeFields := eventmodel.Fields(t.command.Fields).IDAttributes(); len(idAttributeFields) > 0 {
	//
	//			v := Id(strings.ToLower(string(t.command.Title[0]))).Dot(eventmodel.ProcessTitle(idAttributeFields[0].Name))
	//
	//			if idAttributeFields[0].Type == "UUID" {
	//				// in case of an uuid we need to cast it to a string.
	//				v.Dot("String").Call()
	//			}
	//			group.Add(v)
	//		} else {
	//			group.Add(Lit("").Comment("TODO generation could not decide the aggregateID"))
	//		}
	//	}),
	//)

	return z
}

func (t *projectorTemplate) DefaultPath() string {
	return "/" + eventmodel.SliceTitle(t.info.Slice.Title) + "/projector.go"
}

func (t *projectorTemplate) Prepare(ctx context.Context) error {

	//TODO implement survey to ask if it should be a live report or not.
	//var live bool
	//survey.AskOne(&survey.Input{
	//	Message: "Is it a live model",
	//}, &live)
	//t.command = &t.info.Slice.Commands[0]
	return nil
}

func (t *projectorTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
