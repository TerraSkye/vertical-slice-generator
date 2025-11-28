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
	z := NewFile(eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))
	z.ImportAlias(PackageEventSourcing, "cqrs")

	eventPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/../events")
	if err != nil {
		panic(err)
	}

	//var commandName = eventmodel.ProcessTitle(t.command.Title)
	var commandName = "Command"

	z.Const().Call(Line().Comment("//Define the error codes here").Line().Comment("//ErrItemAlreadyExists infra.ErrorCode = 1000 + iota").Line())

	z.Var().Id("_").Qual(PackageEventSourcing, "Command").Op("=").Call(Op("*").Id(commandName)).Call(Nil())
	z.Line().Type().Id(commandName).Add(template.FieldsStruct(t.command.Fields, false))
	z.Func().Params(
		Id("c").Id(commandName)).Id("AggregateID").Params().Params(String()).Block(
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

	stateName := "state"
	stateName = strings.ToLower(stateName[:1]) + stateName[1:]

	//TODO state should be kept internal, lowercase it.
	z.Comment("// TODO-AI keep attributes in state optional")
	z.Type().Id(stateName).StructFunc(func(group *Group) {})

	//TODO
	z.Var().Id("initial" + "State").Op("=").Id(stateName).Block()

	z.Func().Id("evolve").Params(Id("state").Id(stateName), Id("envelope").Op("*").Qual(PackageEventSourcing, "Envelope")).Params(Id(stateName)).BlockFunc(func(group *Group) {
		group.Comment("build state based on its history")

		group.Switch(Id("event").Op(":=").Id("envelope").Dot("Event").Assert(Type())).Block(
			Default().Id("_").Op("=").Id("event"))
		group.Return(Id("state"))

	})

	z.Commentf("/*\nAI-TODO start: implement according to the specifications provided.\nStick to the specification, don´t add new fields, which are not specified.\n\nin case an error is expected - throw an error\n\nRemove the TODO Comment afterwards.\n\n%s\nAI-TODO end\n*/", t.info.Slice.Instructions())
	z.Func().Id("decide").Params(Id("state").Id(stateName), Id("cmd").Id(commandName)).Params(Index().Qual(PackageEventSourcing, "Event"), Error()).BlockFunc(func(group *Group) {
		group.Comment(t.info.Slice.Instructions())

		group.Var().Id("eventList").Op("=").Make(Index().Qual(PackageEventSourcing, "Event"), Op("0"))
		for _, s := range t.command.ProducesEvents() {
			event := t.info.Model.FindEventByID(s)

			group.Id("eventList").Op("=").Append(Id("eventList"), Op("&").Qual(eventPackage, eventmodel.ProcessTitle(event.Title)).Block(DictFunc(func(dict Dict) {
				for _, field := range event.Fields {
					property := Id(eventmodel.ProcessTitle(field.Name))
					//if field.Cardinality != "Single" {
					//	property = property.Index()
					//}

					//dict[property] = Id("cmd").Dot(generator.ToCamelCase(field.Name))
					if len(event.Fields) > 1 {
						dict[property] = Id("cmd").Dot(eventmodel.ProcessTitle(field.Name))
					} else {
						dict[property] = Id("cmd").Dot(eventmodel.ProcessTitle(field.Name)).Op(",")
					}
				}
			})))
		}
		group.Return(Id("eventList"), Nil())
	})

	//// NewCreateItemHandler returns a command handler for the CreateItem Command
	//func NewCreateItemHandler(eventStore cqrs.EventStore) func(ctx context.Context, cmd CreateItem) (cqrs.AppendResult, error) {
	//	return infra.CommandHandler(eventStore, initialState, evolve, decide, infra.WithRevision(cqrs.Any{}))
	//}

	z.Func().Id("New" + commandName + "Handler").Params(Id("eventStore").Qual(PackageEventSourcing, "EventStore")).Params(Func().Params(
		Id("ctx").Qual("context", "Context"),
		// function params
		Id("cmd").Qual("", eventmodel.ProcessTitle(commandName)),
	).Params(Qual(PackageEventSourcing, "AppendResult"), Error())).BlockFunc(func(group *Group) {
		group.Return(Qual(PackageEventSourcing, "NewCommandHandler").Call(
			Id("eventStore"),
			Id("initialState"),
			Id("evolve"),
			Id("decide"),
			Qual(PackageEventSourcing, "WithRevision").Call(Qual(PackageEventSourcing, "Any").Block()),
		))
	})

	return z
}

func (t *commandTemplate) DefaultPath() string {
	return eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")) + "/command.go"
}

func (t *commandTemplate) Prepare(ctx context.Context) error {

	//t.command = &t.info.Slice.Commands[0]
	return nil
}

func (t *commandTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
