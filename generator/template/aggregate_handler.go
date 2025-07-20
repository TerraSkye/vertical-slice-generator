package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

type aggregateHandlerTemplate struct {
	info      *GenerationInfo
	aggregate string
}

func NewAggregateHandlerTemplate(info *GenerationInfo, aggregate string) Template {
	return &aggregateHandlerTemplate{
		info:      info,
		aggregate: aggregate,
	}
}

func (t *aggregateHandlerTemplate) Render(ctx context.Context) write_strategy.Renderer {
	aggregatePackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain")
	if err != nil {
		panic(err)
	}

	basePackage, err := ResolvePackagePath(t.info.OutputFilePath)
	if err != nil {
		panic(err)
	}

	z := NewFile("handlers")
	z.ImportAlias(PackageEventSourcing, "cqrs")
	z.ImportName(aggregatePackage, "")

	z.Func().Id("init").Call().BlockFunc(func(group *Group) {
		group.Qual(basePackage, "RegisterAggregate").Params(
			Func().Params(Id("id").String()).Params(Op("*").Qual(aggregatePackage, eventmodel.AggregateTitle(t.aggregate))).
				BlockFunc(func(group *Group) {
					group.Return(
						Op("&").Qual(aggregatePackage, eventmodel.AggregateTitle(t.aggregate)).Block(
							Dict{
								Id("AggregateBase"): Qual(PackageEventSourcing, "NewAggregateBase").Call(Id("id")).Op(","),
							},
						),
					)
				}),
		)
		//

	})

	return z
}

func (t *aggregateHandlerTemplate) DefaultPath() string {
	return "handlers/" + eventmodel.ScreenTitle(t.aggregate) + ".go"
}

func (t *aggregateHandlerTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *aggregateHandlerTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
