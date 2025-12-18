package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"github.com/terraskye/vertical-slice-generator/template"
	"strings"
)

type commandserviceTemplate struct {
	info    *GenerationInfo
	command *eventmodel.Command
}

func NewCommandServiceTemplate(info *GenerationInfo, command *eventmodel.Command) Template {
	return &commandserviceTemplate{
		info:    info,
		command: command,
	}
}

func (t *commandserviceTemplate) Render(ctx context.Context) write_strategy.Renderer {
	commandPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain/commands")
	if err != nil {
		panic(err)
	}
	aggregatePackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain")
	if err != nil {
		panic(err)
	}

	z := NewFile(eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))
	z.ImportAlias(PackageEventSourcing, "cqrs")
	z.ImportName(commandPackage, "")
	z.ImportName(aggregatePackage, "")

	z.Type().Id("Service").Interface(Id(eventmodel.ProcessTitle(strings.ReplaceAll(t.info.Slice.Title, "slice:", ""))).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()))

	z.Line().Type().Id("Payload").Add(template.FieldsStruct(t.command.Fields, false))

	z.Line().Type().Id("service").StructFunc(func(group *Group) {
		group.Add(Id("commandBus").Qual(PackageEventSourcing, "CommandBus"))
	})

	z.Func().Id("NewService").ParamsFunc(func(group *Group) {
		group.Add(Id("commandBus").Qual(PackageEventSourcing, "CommandBus"))
	}).Params(Id("Service")).Block(Return().Op("&").Id("service").Block(DictFunc(func(dict Dict) {
		dict[Id("commandBus")] = Id("commandBus").Op(",")
	})))

	z.Func().Params(Id("s").Op("*").Op("service")).Id(eventmodel.ProcessTitle(strings.ReplaceAll(t.info.Slice.Title, "slice:", ""))).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()).BlockFunc(func(group *Group) {
		group.Add(Id("cmd").Op(":=").Op("&").Qual("", eventmodel.ProcessTitle(t.command.Title)).Block(DictFunc(func(dict Dict) {
			for _, field := range t.command.Fields {
				property := Id(eventmodel.ProcessTitle(field.Name))
				if len(t.command.Fields) > 1 {
					dict[property] = Id("payload").Dot(eventmodel.ProcessTitle(field.Name))
				} else {
					dict[property] = Id("payload").Dot(eventmodel.ProcessTitle(field.Name)).Op(",")
				}
			}
		})))

		group.If(Err().Op(":=").Id("s").Dot("commandBus").Dot("Send").Call(Id("ctx"), Id("cmd")).Op(";").Add(Err().Op("!=").Nil()).Block(Return().Err())).Line()

		group.Add(Return().Nil())
	})

	return z
}

func (t *commandserviceTemplate) DefaultPath() string {
	return eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")) + "/service.go"
}

func (t *commandserviceTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *commandserviceTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
