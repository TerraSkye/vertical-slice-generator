package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

type eventHandlerTemplate struct {
	info  *GenerationInfo
	event *eventmodel.Event
}

func NewEventHandlerTemplate(info *GenerationInfo, event *eventmodel.Event) Template {
	return &eventHandlerTemplate{
		info:  info,
		event: event,
	}
}

func (t *eventHandlerTemplate) Render(ctx context.Context) write_strategy.Renderer {
	eventPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/events")
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
	z.ImportName(eventPackage, "")
	z.ImportName(aggregatePackage, "")

	z.Func().Id("init").Call().BlockFunc(func(group *Group) {
		group.Qual(basePackage, "RegisterEvent").Params(
			Func().Params(Id("aggregate").Op("*").Qual(aggregatePackage, eventmodel.AggregateTitle(t.event.Aggregate))).Func().Params(
				Id("event").Op("*").Qual(eventPackage, eventmodel.ProcessTitle(t.event.Title))).
				BlockFunc(func(group *Group) {
					group.Return(Func().Params(Id("event").Op("*").Qual(eventPackage, eventmodel.ProcessTitle(t.event.Title))).Block(
						Id("aggregate").Dot("On" + eventmodel.ProcessTitle(t.event.Title)).Call(Id("event")),
					))
				}),
		)
		//

	})

	return z
}

func (t *eventHandlerTemplate) DefaultPath() string {
	return "handlers/" + eventmodel.ScreenTitle(t.event.Title) + ".go"
}

func (t *eventHandlerTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *eventHandlerTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
