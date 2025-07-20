package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

type commandhandlerTemplate struct {
	info    *GenerationInfo
	command *eventmodel.Command
}

func NewCommandHandlerTemplate(info *GenerationInfo, command *eventmodel.Command) Template {
	return &commandhandlerTemplate{
		info:    info,
		command: command,
	}
}

func (t *commandhandlerTemplate) Render(ctx context.Context) write_strategy.Renderer {
	commandPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain/commands")
	if err != nil {
		panic(err)
	}
	aggregatePackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain")
	if err != nil {
		panic(err)
	}

	basePackage, err := ResolvePackagePath(t.info.OutputFilePath)
	if err != nil {
		panic(err)
	}

	z := NewFile("handlers")
	z.ImportName(PackageEventSourcing, "cqrs")
	z.ImportName(commandPackage, "")
	z.ImportName(aggregatePackage, "")

	z.Func().Id("init").Call().BlockFunc(func(group *Group) {
		group.Qual(basePackage, "RegisterCommand").Params(
			Func().Params(Id("aggregate").Op("*").Qual(aggregatePackage, eventmodel.AggregateTitle(t.command.Aggregate))).Func().Params(
				Id("ctx").Qual(PackagePathContext, "Context"), Id("command").Op("*").Qual(commandPackage, eventmodel.ProcessTitle(t.command.Title))).Params(Error()).
				BlockFunc(func(group *Group) {
					group.Return(Func().Params(Id("ctx").Qual(PackagePathContext, "Context"), Id("command").Op("*").Qual(commandPackage, eventmodel.ProcessTitle(t.command.Title))).Params(Error()).Block(Return(
						Id("aggregate").Dot(eventmodel.ProcessTitle(t.command.Title)).Call(Id("ctx"), Id("command")),
					)))
				}),
		)
		//

	})

	return z
}

func (t *commandhandlerTemplate) DefaultPath() string {
	return "handlers/" + eventmodel.ScreenTitle(t.info.Slice.Commands[0].Title) + ".go"
}

func (t *commandhandlerTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *commandhandlerTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
