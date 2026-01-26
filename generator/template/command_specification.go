package template

import (
	"context"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
)

type commandSpecificationTemplate struct {
	info    *GenerationInfo
	command *eventmodel.Command
}

func NewCommandSpecificationTemplate(info *GenerationInfo, command *eventmodel.Command) Template {
	return &commandSpecificationTemplate{
		info:    info,
		command: command,
	}
}

func (t *commandSpecificationTemplate) Render(ctx context.Context) write_strategy.Renderer {
	z := NewFile(eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))
	z.ImportAlias(PackageEventSourcing, "cqrs")
	z.ImportAlias(PackagePathJson, "json")

	eventPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/events")
	if err != nil {
		panic(err)
	}

	z.Func().Id("Test_decide").Params(Id("t").Op("*").Qual("testing", "T")).BlockFunc(func(group *Group) {

		group.Type().Id("args").Struct(
			Id("history").Index().Op("*").Qual(PackageEventSourcing, "Envelope"),
			Id("cmd").Id("Command"),
		)

		group.Line().Line()

		group.Id("tests").Op(":=").Index().Struct(
			Id("name").String(),
			Id("args").Id("args"),
			Id("want").Index().Qual(PackageEventSourcing, "Event"),
			Id("wantErr").Bool(),
		).BlockFunc(func(group *Group) {
			// Generate test cases from specifications
			for _, spec := range t.info.Slice.Specifications {
				group.Values(Dict{
					Id("name"): Lit(spec.Title),
					Id("args"): Id("args").Values(Dict{
						Id("history"): Index().Op("*").Qual(PackageEventSourcing, "Envelope").ValuesFunc(func(g *Group) {
							for _, given := range spec.Given {
								event := t.info.Model.FindEventByID(given.LinkedID)
								if event == nil {
									continue
								}
								g.Values(Dict{
									Line().Id("Event"): Op("&").Qual(eventPackage, eventmodel.ProcessTitle(event.Title)).Values(t.buildFieldsDict(given.Fields)),
								})
							}
						}),
						Id("cmd"): Id("Command").Values(t.buildFieldsDict(t.getCommandFields(spec))),
					}),
					Id("want"):    t.buildWantEvents(spec, eventPackage),
					Id("wantErr"): Lit(t.expectsError(spec)),
				}).Op(",")
			}
		})

		group.For(List(Id("_"), Id("tt"))).Op(":=").Op("range").Id("tests").BlockFunc(func(group *Group) {
			group.Id("t").Dot("Run").Call(Id("tt").Dot("name"), Func().Params(Id("t").Op("*").Qual("testing", "T")).BlockFunc(func(group *Group) {
				// Initialize state
				group.Line().Id("s").Op(":=").Id("initialState")

				// Apply history events
				group.For(List(Id("_"), Id("envelope"))).Op(":=").Op("range").Id("tt").Dot("args").Dot("history").BlockFunc(func(group *Group) {
					group.Id("s").Op("=").Id("evolve").Call(Id("s"), Id("envelope"))
				})

				group.Line()

				// Call decide
				group.List(Id("got"), Id("err")).Op(":=").Id("decide").Call(Id("s"), Id("tt").Dot("args").Dot("cmd"))

				// Check error
				group.If(Parens(Id("err").Op("!=").Nil()).Op("!=").Id("tt").Dot("wantErr")).BlockFunc(func(g *Group) {
					g.Id("t").Dot("Errorf").Call(Lit("decide() error = %v, wantErr %v"), Id("err"), Id("tt").Dot("wantErr"))
					g.Return()
				})

				group.Line()

				// Check event count
				group.If(Len(Id("got")).Op("!=").Len(Id("tt").Dot("want"))).BlockFunc(func(g *Group) {
					g.Id("t").Dot("Errorf").Call(Lit("decide() expected %d but got %d events"), Len(Id("tt").Dot("want")), Len(Id("got")))
					g.Return()
				})

				group.Line()

				// JSON comparison
				group.List(Id("wantData"), Id("_")).Op(":=").Qual("encoding/json", "Marshal").Call(Id("got"))
				group.List(Id("gotData"), Id("_")).Op(":=").Qual("encoding/json", "Marshal").Call(Id("tt").Dot("want"))

				group.Line()

				group.If(Op("!").Qual("bytes", "Equal").Call(Id("wantData"), Id("gotData"))).BlockFunc(func(g *Group) {
					g.Id("t").Dot("Errorf").Call(Lit("decide() expected %s but got %s "), Id("wantData"), Id("gotData"))
					g.Return()
				})
			}))
		})
	})

	return z
}

// buildFieldsDict builds a dictionary of field values from their Example values
func (t *commandSpecificationTemplate) buildFieldsDict(fields []eventmodel.Field) Dict {
	dict := Dict{}
	for _, field := range fields {
		if field.Example == "" {
			continue
		}
		fieldName := Id(eventmodel.ProcessTitle(field.Name))
		dict[fieldName] = t.fieldExampleValue(field)
	}
	return dict
}

// fieldExampleValue returns the appropriate literal value for a field's example
func (t *commandSpecificationTemplate) fieldExampleValue(field eventmodel.Field) *Statement {
	switch field.Type {
	case "Boolean":
		return Lit(field.Example == "true")
	case "Int", "Long":
		// Try to parse as int, fallback to string literal
		return Lit(field.Example)
	case "Double", "Decimal":
		return Lit(field.Example)
	default:
		return Lit(field.Example)
	}
}

// getCommandFields extracts command fields from a specification's When clause
func (t *commandSpecificationTemplate) getCommandFields(spec eventmodel.Specification) []eventmodel.Field {
	if len(spec.When) > 0 {
		return spec.When[0].Fields
	}
	return nil
}

// expectsError returns true if the specification expects an error (SPEC_ERROR in Then)
func (t *commandSpecificationTemplate) expectsError(spec eventmodel.Specification) bool {
	for _, then := range spec.Then {
		if then.Type == "SPEC_ERROR" {
			return true
		}
	}
	return false
}

// buildWantEvents builds the want slice of events from the specification's Then clause
func (t *commandSpecificationTemplate) buildWantEvents(spec eventmodel.Specification, eventPackage string) *Statement {
	if t.expectsError(spec) {
		return Nil()
	}

	return Index().Qual(PackageEventSourcing, "Event").ValuesFunc(func(g *Group) {
		for _, then := range spec.Then {
			if then.Type == "SPEC_ERROR" {
				continue
			}
			event := t.info.Model.FindEventByID(then.LinkedID)
			if event == nil {
				continue
			}
			g.Op("&").Qual(eventPackage, eventmodel.ProcessTitle(event.Title)).Values(t.buildFieldsDict(then.Fields))
		}
	})
}

func (t *commandSpecificationTemplate) DefaultPath() string {
	return "slices/" + eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")) + "/command_test.go"
}

func (t *commandSpecificationTemplate) Prepare(ctx context.Context) error {

	//t.command = &t.info.Slice.Commands[0]
	return nil
}

func (t *commandSpecificationTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
