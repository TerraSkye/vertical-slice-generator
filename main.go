package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	. "github.com/dave/jennifer/jen"
	"github.com/gosimple/slug"
	"github.com/terraskye/vertical-slice-generator/generator"
	"github.com/terraskye/vertical-slice-generator/generator/config"
	"github.com/terraskye/vertical-slice-generator/generator/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

var (
	//go:embed cart.json
	configuration []byte
)

func Checkboxes(label string, opts []string) []string {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	survey.AskOne(prompt, &res)

	return res
}

func main() {
	var cfg config.Configuration

	if err := json.Unmarshal(configuration, &cfg); err != nil {
		log.Fatal(err)
	}

	var tasks = make([]string, len(cfg.Slices))

	for i, slice := range cfg.Slices {
		tasks[i] = slice.Title
	}
	//slicesToGenerate := Checkboxes(
	//	"Which slice would you like to generate", append([]string{"all"}, tasks...),
	//)

	slicesToGenerate := []string{"slice: cart items", "all"}
	if slices.Contains(slicesToGenerate, "all") {
		slicesToGenerate = tasks[1:]
	}

	var files = make(map[string]*File)

	alreadyGenerate := make(map[string]struct{})

	registeryFile := fmt.Sprintf("%s/registery.go", strings.ToLower(cfg.CodeGen.Domain))

	files[registeryFile] = generator.GetFile(strings.ToLower(cfg.CodeGen.Domain))
	files[registeryFile].ImportAlias("github.com/terraskye/eventsourcing", "cqrs")

	files[registeryFile].Comment("CommandHandler type is being used to dispatch commands onto the aggregate").Line().Type().Id("CommandHandler").Types(
		Id("A").Any(),
		Id("C").Any(),
	).Func().Call(Id("aggregate").Op("A")).Func().Params(Id("ctx").Qual("context", "Context"), Id("command").Op("C")).Params(Error())

	files[registeryFile].Comment("EventHandler type is being used to dispatch event onto the aggregate").Line().Type().Id("EventHandler").Types(
		Id("A").Any(),
		Id("E").Qual("github.com/terraskye/eventsourcing", "Event"),
	).Func().Call(Id("aggregate").Op("A")).Func().Params(Id("ctx").Qual("context", "Context"), Id("event").Op("E")).Params(Error())

	files[registeryFile].Comment("AggregateHandler type is being used to initialize an aggregate").Line().Type().Id("AggregateHandler").Types(
		Id("A").Any(),
	).Func().Params(Id("id").Qual("github.com/google/uuid", "UUID")).Params(Op("A"))

	files[registeryFile].Var().Defs(
		Id("commandRegistry").Op("=").Make(Map(String()).Any()),
		Id("eventRegistry").Op("=").Make(Map(String()).Any()),
		Id("aggregateRegistry").Op("=").Make(Map(String()).Func().Params(Id("id").Qual("github.com/google/uuid", "UUID")).Params(Any())),
		Id("commandToAggregate").Op("=").Make(Map(String()).String()),
		Id("eventDecoder").Op("=").Make(Map(String()).Func().Params(Id("raw").Index().Byte()).Params(Any(), Error())),
	)

	files[registeryFile].Comment("AggregateForCommand").Line().Func().Id("AggregateForCommand").Params(Id("cmd").Qual("github.com/terraskye/eventsourcing", "Command")).Params(Qual("github.com/terraskye/eventsourcing", "Aggregate"), Error()).BlockFunc(func(group *Group) {

		group.Add(Id("cmdType").Op(":=").Qual("github.com/terraskye/eventsourcing", "TypeName").Call(Id("cmd")))

		group.Line().List(Id("aggType"), Id("ok").Op(":=").Id("commandToAggregate").Types(Id("cmdType")))

		group.Line().If(Op("!").Id("ok")).Block(Return(Nil(), Qual("fmt", "Errorf").Call(Op("\"no command to aggregate mapping found for: %s\""), Id("cmdType"))))

		group.Line().List(Id("aggHandler"), Id("ok").Op(":=").Id("aggregateRegistry").Types(Id("aggType")))

		group.Line().If(Op("!").Id("ok")).Block(Return(Nil(), Qual("fmt", "Errorf").Call(Op("\"invalid aggregate for: %s\""), Id("aggType"))))

		group.Line().List(Id("agg"), Id("ok").Op(":=").Id("aggHandler").Call(Id("cmd").Dot("AggregateID").Call())).Assert(Qual("github.com/terraskye/eventsourcing", "Aggregate"))

		group.Line().If(Op("!").Id("ok")).Block(Return(Nil(), Qual("fmt", "Errorf").Call(Op("\"invalid aggregate for: %s\""), Id("aggType"))))

		group.Return(Id("agg"), Nil())

	})

	files[registeryFile].Comment("RegisterCommand").Line().Func().Id("RegisterCommand").Types(Id("A").Any(), Id("T").Any()).ParamsFunc(func(group *Group) {
		group.Add(Id("handler").Op("CommandHandler").Types(Id("A"), Id("T")))
	}).BlockFunc(func(group *Group) {

		group.Var().Id("cmd").Op("T")

		group.Add(Id("cmdType").Op(":=").Qual("github.com/terraskye/eventsourcing", "TypeName").Call(Id("cmd")))

		group.Add(Id("commandRegistry").Index(Id("cmdType"))).Op("=").Func().Params(Id("aggregate").Qual("github.com/terraskye/eventsourcing", "Aggregate")).Params(Func().Params(Qual("context", "Context"), Qual("github.com/terraskye/eventsourcing", "Command")).Params(Error())).BlockFunc(func(group *Group) {
			group.Add(Return().Func().Call())
		})
		//group.Params(Id("aggregate").Id("A")).Params(, Id("command").Id("T"))

		//group.Add(Id("commandToAggregate").Index().Qual("github.com/terraskye/eventsourcing", "TypeName").Call(Id("handler")))
	})

	for _, sliceName := range slicesToGenerate {
		for _, slice := range cfg.Slices {
			if slice.Title == sliceName {

				rootPackage, err := generator.ResolvePackagePath(fmt.Sprintf("%s", strings.ToLower(cfg.CodeGen.Domain)))
				if err != nil {
					log.Fatal(err)
				}

				domainPackage, err := generator.ResolvePackagePath(fmt.Sprintf("%s/domain", strings.ToLower(cfg.CodeGen.Domain)))
				if err != nil {
					log.Fatal(err)
				}

				eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("%s/events", strings.ToLower(cfg.CodeGen.Domain)))
				if err != nil {
					log.Fatal(err)
				}

				commandPackage, err := generator.ResolvePackagePath(fmt.Sprintf("%s/domain/commands", strings.ToLower(cfg.CodeGen.Domain)))
				if err != nil {
					log.Fatal(err)
				}

				generatedSliceName := strings.ToLower(generator.ToCamelCase(slice.Title[7:]))

				_ = generatedSliceName
				//fmt.Printf("===== [slice: %s] =====\n", slice.Title)

				//fmt.Println("> aggregate")
				for _, aggregateName := range slice.Aggregates {
					for _, aggregate := range cfg.Aggregates {
						if _, ok := alreadyGenerate[aggregate.Title]; !ok && aggregate.Title == aggregateName {
							alreadyGenerate[aggregate.Title] = struct{}{}
							aggregateFile := fmt.Sprintf("%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(aggregate.Title))

							if _, ok := files[aggregateFile]; !ok {
								files[aggregateFile] = generator.GetFile("domain")
								files[aggregateFile].ImportAlias("", eventsPackage)
								files[aggregateFile].ImportAlias(commandPackage, "commands")
							}

							files[aggregateFile].Type().Id(generator.ToCamelCase(aggregate.Title)).Add(StructFunc(func(group *Group) {
								group.Add(Op("*").Qual("github.com/terraskye/eventsourcing", "AggregateBase"))
								for _, statement := range template.Fields(aggregate.Fields) {
									group.Add(statement)
								}
							}))

							handlerFile := fmt.Sprintf("%s/handlers/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(aggregate.Title))

							if _, ok := files[handlerFile]; !ok {
								files[handlerFile] = generator.GetFile("handlers")
								files[handlerFile].ImportAlias("github.com/terraskye/eventsourcing", "cqrs")
							}

							files[handlerFile].Func().Id("init").Call().Block(
								Qual(rootPackage, "RegisterAggregate").Params(
									Func().Params(Id("id").Qual("github.com/google/uuid", "UUID")).Params(Qual(domainPackage, generator.ToCamelCase(aggregate.Title))).Block(
										Return(Op("&").Qual(domainPackage, generator.ToCamelCase(aggregate.Title)).Add(Block(Dict{
											Id("AggregateBase"): Qual("github.com/terraskye/eventsourcing", "NewAggregateBase").Call(Id("id")).Op(","),
										})))),
								),
							)
						}
					}
				}

				fmt.Println("> commands")
				for _, command := range slice.Commands {
					aggregateFile := fmt.Sprintf("%s/domain/commands/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title)))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile("commands")
					}
					files[aggregateFile].Var().Id("_").Qual("github.com/terraskye/eventsourcing", "Command").Op("=").Call(Op("*").Id(generator.ToCamelCase(command.Title))).Call(Nil())
					files[aggregateFile].Type().Id(generator.ToCamelCase(command.Title)).Add(template.FieldsStruct(command.Fields))

					files[aggregateFile].Func().Call(Id("c").Id(generator.ToCamelCase(command.Title))).Id("AggregateID").Params().Params(Qual("github.com/google/uuid", "UUID")).Block(
						Return(Id("c").Dot("AggregateId")))

					// add the "Apply function to the aggregate.
					files[fmt.Sprintf("%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(command.Aggregate))].
						Func().Params(Id(strings.ToLower(string(command.Aggregate[0]))).Op("*").Id(generator.ToCamelCase(command.Aggregate))).
						Id(generator.ToCamelCase(command.Title)).
						Params(Id("ctx").Qual("context", "Context"), Id("cmd").Op("*").Qual(commandPackage, generator.ToCamelCase(command.Title))).
						Params(Error()).BlockFunc(func(body *Group) {

						// populate the event that this command produces.
						for _, dependency := range command.Dependencies {
							if dependency.Type == "OUTBOUND" && dependency.ElementType == "EVENT" {
								for _, slices := range cfg.Slices {
									for _, event := range slices.Events {
										if event.ID == dependency.ID {
											// outbound event.
											body.Id(strings.ToLower(string(command.Aggregate[0]))).Dot("AppendEvent").
												Call(Id("ctx"), Op("&").Qual(eventsPackage, generator.ToCamelCase(event.Title)).Block(DictFunc(func(dict Dict) {
													for _, field := range event.Fields {
														property := Id(generator.ToCamelCase(field.Name))
														//if field.Cardinality != "Single" {
														//	property = property.Index()
														//}

														//dict[property] = Id("cmd").Dot(generator.ToCamelCase(field.Name))
														if len(event.Fields) > 1 {
															dict[property] = Id("cmd").Dot(generator.ToCamelCase(field.Name))
														} else {
															dict[property] = Id("cmd").Dot(generator.ToCamelCase(field.Name)).Op(",")
														}
													}
												})))
										}
									}
								}
							}
						}

						body.Return(Nil())
					})

					fmt.Println("\t- " + command.Title)

					handlerFile := fmt.Sprintf("%s/handlers/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title)))

					if _, ok := files[handlerFile]; !ok {
						files[handlerFile] = generator.GetFile("handlers")
					}

					files[handlerFile].Func().Id("init").Call().Block(
						Qual(rootPackage, "RegisterCommand").Params(
							Func().Params(Id("aggregate").Op("*").Qual(domainPackage, generator.ToCamelCase(command.Aggregate))).
								Func().Params(Id("ctx").Qual("context", "Context"),
								Id("cmd").Op("*").Qual(commandPackage, generator.ToCamelCase(command.Title))).Params(Error()).Block(
								Return(Func().Params(Id("ctx").Qual("context", "Context"),
									Id("cmd").Op("*").Qual(commandPackage, generator.ToCamelCase(command.Title))).Params(Error()).Block(
									Return(Id("aggregate").Dot(generator.ToCamelCase(command.Title)).Call(Id("ctx"), Id("cmd"))),
								)),
							),
						),
					)

				}
				fmt.Println("> events")
				for _, event := range slice.Events {
					fmt.Println("\t- " + event.Title)
					eventFile := fmt.Sprintf("%s/events/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title)))
					if _, ok := files[eventFile]; !ok {
						files[eventFile] = generator.GetFile("events")

					}
					files[eventFile].Var().Id("_").Qual("github.com/terraskye/eventsourcing", "Event").Op("=").Call(Op("*").Id(generator.ToCamelCase(event.Title))).Call(Nil())
					files[eventFile].Type().Id(generator.ToCamelCase(event.Title)).Add(template.FieldsStruct(event.Fields))

					files[eventFile].Func().Call(Id("e").Id(generator.ToCamelCase(event.Title))).Id("AggregateID").Params().Params(Qual("github.com/google/uuid", "UUID")).Block(
						Return(Id("e").Dot("AggregateId")))

					// add the "on" function to the aggregate
					files[fmt.Sprintf("%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(event.Aggregate))].
						Func().Params(Id(strings.ToLower(string(event.Aggregate[0]))).Op("*").Id(generator.ToCamelCase(event.Aggregate))).
						Id(generator.ToCamelCase("On" + event.Title)).
						Params(Id("event").Op("*").Qual(eventsPackage, generator.ToCamelCase(event.Title))).
						Params().BlockFunc(func(body *Group) {
					}).Line()

					// register the handler.
					handlerFile := fmt.Sprintf("%s/handlers/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title)))
					if _, ok := files[handlerFile]; !ok {
						files[handlerFile] = generator.GetFile("handlers")
					}

					// handler function, TODO add tracing by default here.
					files[handlerFile].Func().Id("init").Call().Block(
						Qual(rootPackage, "RegisterEvent").Params(
							Func().Params(Id("aggregate").Op("*").Qual(domainPackage, generator.ToCamelCase(event.Aggregate))).
								Func().Params(Id("event").Op("*").Qual(eventsPackage, generator.ToCamelCase(event.Title))).Block(
								Return(Func().Params(Id("event").Op("*").Qual(eventsPackage, generator.ToCamelCase(event.Title)))).Block(
									Id("aggregate").Dot(generator.ToCamelCase("On" + event.Title)).Call(Id("event"))),
							),
						),
					)

				}

				if commands := slice.Commands; len(commands) > 0 {
					fmt.Println("> state change")
					aggregateFile := fmt.Sprintf("%s/%s/service.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						files[aggregateFile].ImportAlias("github.com/terraskye/eventsourcing", "cqrs")
					}
					files[aggregateFile].Type().Id("Service").Interface(Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()))

					for _, command := range commands {
						fmt.Println("\t- " + command.Title)
						files[aggregateFile].Line().Type().Id("Payload").Add(template.FieldsStruct(command.Fields))
					}

					files[aggregateFile].Type().Id("service").StructFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/terraskye/eventsourcing", "CommandBus"))
					})

					files[aggregateFile].Func().Id("NewService").ParamsFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/terraskye/eventsourcing", "CommandBus"))
					}).Params(Id("Service")).Block(Return().Op("&").Id("service").Block(DictFunc(func(dict Dict) {
						dict[Id("commandBus")] = Id("commandBus").Op(",")
					})))

					files[aggregateFile].Func().Params(Id("s").Op("*").Op("service")).Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()).BlockFunc(func(group *Group) {
						group.Add(Id("cmd").Op(":=").Op("&").Qual(commandPackage, generator.ToCamelCase(slice.Commands[0].Title)).Block(DictFunc(func(dict Dict) {
							for _, field := range slice.Commands[0].Fields {
								property := Id(generator.ToCamelCase(field.Name))
								if len(slice.Commands[0].Fields) > 1 {
									dict[property] = Id("payload").Dot(generator.ToCamelCase(field.Name))
								} else {
									dict[property] = Id("payload").Dot(generator.ToCamelCase(field.Name)).Op(",")
								}
							}
						})))

						group.If(Err().Op(":=").Id("s").Dot("commandBus").Dot("Send").Call(Id("ctx"), Id("cmd")).Op(";").Add(Err().Op("!=").Nil()).Block(Return().Err())).Line()

						group.Add(Return().Nil())
					})

					apiHandlerFile := fmt.Sprintf("%s/%s/http.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					files[apiHandlerFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					files[apiHandlerFile].ImportAlias("github.com/afosto/go-json", "json")
					files[apiHandlerFile].ImportAlias("github.com/afosto/utils-go/http/request", "afreq")

					files[apiHandlerFile].Func().Id("MakeHttpHandler").Params(Id("r").Op("*").Qual("github.com/gorilla/mux", "Router"), Id("s").Id("Service")).Params(Qual("net/http", "Handler")).BlockFunc(func(group *Group) {
						//TODO paramter filter for type.
						group.Id("r").Dot("Methods").Call(Id("\"POST\"")).Dot("Path").Call(Id(fmt.Sprintf("\"%s/{id:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}}/%s\"", strings.ToLower(cfg.CodeGen.Domain), slug.Make(generator.ToCamelCase(slice.Title[7:]))))).Dot("Handler").Call(Qual("github.com/go-kit/kit/transport/http", "NewServer").Params(
							Func().Params(Id("ctx").Qual("context", "Context"), Id("request").Any()).Params(Any(), Error()).BlockFunc(func(group *Group) {
								// endpoint body
								group.If(Err().Op(":=").Id("s").Dot(generator.ToCamelCase(slice.Title[7:])).Call(Id("ctx"), Id("request").Assert(Id("Payload"))), Err().Op("!=").Nil()).Block(Return(Nil(), Err()))

								group.Return(Struct().Block(), Nil())
							}),
							Id(fmt.Sprintf("decode%sRequest", generator.ToCamelCase(slice.Title[7:]))),
							Id(fmt.Sprintf("encode%sRequest", generator.ToCamelCase(slice.Title[7:]))),
						),
						)
						group.Return(Id("r"))
					}).Line()

					files[apiHandlerFile].Func().Id(fmt.Sprintf("decode%sRequest", generator.ToCamelCase(slice.Title[7:]))).Params(Id("ctx").Qual("context", "Context"), Id("r").Op("*").Qual("net/http", "Request")).Params(Interface(), Error()).BlockFunc(func(group *Group) {

						group.Line().Var().Id("payload").Struct(Id("Data").StructFunc(func(group *Group) {

							for _, field := range config.Fields(slice.Commands[0].Fields).DataAttributes() {
								property := Id(generator.ToCamelCase(field.Name))
								if field.Cardinality != "Single" {
									property = property.Index()
								}
								switch field.Type {
								case "String":
									property = property.String().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "UUID":
									property = property.String().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name), "validate": "omitempty,uuid4"})
								case "Boolean":
									property = property.Bool().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "Double":
									property = property.Float64().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "Date":
									property = property.Qual("time", "Time").Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "DateTime":
									property = property.Qual("time", "Time").Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "Long":
									property = property.Int64().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "Int":
									property = property.Int().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								case "Custom":
									property = property.Interface().Tag(map[string]string{"json": generator.ToSnakeCase(field.Name)})
								}
								group.Add(property)
							}

						}).Tag(map[string]string{"json": "data"}))

						group.Line().Line()

						//json decode payload into payload struct
						group.IfFunc(func(c *Group) {
							c.Id("err").Op(":=").
								Qual("github.com/afosto/go-json", "NewDecoder").Params(
								Id("r").Dot("Body")).Dot("UseAutoTrimSpace").Call().Dot("Decode").Call(Op("&").Id("payload")).Op(";").Err().Op("!=").Nil()
						}).BlockFunc(func(group *Group) {
							group.Return(Nil(), Qual("github.com/pkg/errors", "WithStack").Params(Err()))
						})

						//go-playground validate data
						group.IfFunc(func(c *Group) {
							c.Id("err").Op(":=").
								Qual("github.com/afosto/utils-go/intl18", "Validator").Dot("StructCtx").Call(Id("ctx"), Id("payload")).Op(";").Err().Op("!=").Nil()
						}).BlockFunc(func(group *Group) {
							group.Return(Nil(), Qual("github.com/pkg/errors", "WithStack").Params(Err()))

						}).Line()

						group.Return(Id("Payload").BlockFunc(func(group *Group) {

							v := Dict{}
							for _, field := range config.Fields(slice.Commands[0].Fields).IDAttributes() {

								if field.Type == "UUID" {
									v[Id(generator.ToCamelCase(field.Name))] = Qual("github.com/google/uuid", "MustParse").Call(Id("mux").Dot("Vars").Call(Id("r")).Types(Id(strconv.Quote(field.Name))))
								} else {

									v[Id(generator.ToCamelCase(field.Name))] = Id("mux").Dot("Vars").Call(Id("r")).Types(Id(strconv.Quote(field.Name)))
								}
							}
							for _, field := range config.Fields(slice.Commands[0].Fields).DataAttributes() {
								if field.Type == "UUID" {
									v[Id(generator.ToCamelCase(field.Name))] = Qual("github.com/afosto/utils-go/http/request", "UUIDOrNil").Call(Id("payload").Dot("Data").Dot(generator.ToCamelCase(field.Name)))
								} else {

									v[Id(generator.ToCamelCase(field.Name))] = Id("payload").Dot("Data").Dot(generator.ToCamelCase(field.Name))
								}
							}

							if len(v) > 1 {
								group.Add(v)
							}

						}), Nil())
					})
					files[apiHandlerFile].Func().Id(fmt.Sprintf("encode%sRequest", generator.ToCamelCase(slice.Title[7:]))).Params(Id("ctx").Qual("context", "Context"), Id("writer").Qual("net/http", "ResponseWriter"), Id("i").Interface()).Params(Error()).BlockFunc(func(group *Group) {

						group.Id("writer").Dot("WriteHeader").Call(Qual("net/http", "StatusAccepted"))

						group.Return(Nil())
					})

				}

				if readModels := slice.Readmodels; len(readModels) > 0 {
					fmt.Println("> state view")

					for _, readModel := range readModels {
						fmt.Println("\t- " + readModel.Title)
						aggregateFile := fmt.Sprintf("%s/%s/readmodel.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						if _, ok := files[aggregateFile]; !ok {
							files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						}

						files[aggregateFile].Type().Id("Query").Add(template.FieldsStruct(readModel.Fields.IDAttributes()))

						files[aggregateFile].Func().Params(Id("q").Id("Query")).Id("ID").Params().Params(Index().Byte()).Block(Comment("//TODO implement me").Line().Return().Nil())

						files[aggregateFile].Line().Type().Id("ReadModel").Add(template.FieldsStruct(readModel.Fields))

						projectorFile := fmt.Sprintf("%s/%s/projector.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						files[projectorFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))

						files[projectorFile].Type().Id("projector").Struct(
						//Id("entities").Map(Qual("github.com/google/uuid", "UUID")).Op("*").Id(generator.ToCamelCase(slice.Readmodels[0].Title) + "ReadModel"),
						)
						files[projectorFile].Func().Id("NewProjector").ParamsFunc(func(group *Group) {
							//group.Add(Id("eventbus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "EventBus"))
						}).Params(Op("*").Qual("github.com/terraskye/eventsourcing", "EventGroupProcessor")).Block(
							Id("p").Op(":=").Op("&").Id("projector").Block(),
							ReturnFunc(func(group *Group) {

								z := []Code{
									(Op("\"" + generator.ToCamelCase(slice.Title[7:]) + "\"")),
								}
								for _, dependency := range readModel.Dependencies {
									if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
										for _, slices := range cfg.Slices {
											for _, event := range slices.Events {
												if event.ID == dependency.ID {
													// outbound event.
													z = append(z, Qual("github.com/terraskye/eventsourcing", "NewGroupEventHandler").Call(Id("p").Dot("On"+generator.ToCamelCase(dependency.Title))))
												}
											}
										}
									}
								}

								group.Add(Qual("github.com/terraskye/eventsourcing", "NewEventGroupProcessor").Call(z...))
							},
							),
						)

						// event handlers to populate the aggregate
						for _, dependency := range readModel.Dependencies {
							if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
								for _, slices := range cfg.Slices {
									for _, event := range slices.Events {
										if event.ID == dependency.ID {
											// outbound event.

											files[projectorFile].ImportName(eventsPackage, "events")
											files[projectorFile].Line().Func().Params(Id("p").Op("*").Id("projector")).
												Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(filepath.Dir(eventsPackage), generator.ToCamelCase(dependency.Title))).Params(Error()).BlockFunc(func(group *Group) {

												group.Return().Nil()
											})
										}
									}
								}
							}
						}

						queryHandlerFile := fmt.Sprintf("%s/%s/query.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						files[queryHandlerFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))

						files[queryHandlerFile].Type().Id("queryHandler").Struct(
						//Id("repository").Map(Qual("github.com/google/uuid", "UUID")).Op("*").Id("ReadModel"),
						)

						files[queryHandlerFile].Func().Id("NewQueryHandler").ParamsFunc(func(group *Group) {
							//group.Add(Id("eventbus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "EventBus"))
						}).Params(Qual("github.com/terraskye/eventsourcing", "GenericQueryHandler").Types(Id("Query"), Id("ReadModel"))).Block(
							Return(Op("&").Id("queryHandler").Block()),
						)

						files[queryHandlerFile].Line().Func().Params(Id("p").Op("*").Id("queryHandler")).
							Id("HandleQuery").Params(Id("ctx").Qual("context", "Context"), Id("q").Id("Query")).Params(Id("ReadModel"), Error()).BlockFunc(func(group *Group) {

							group.Return(Id("ReadModel").Block(), Nil())
						})

						apiHandlerFile := fmt.Sprintf("%s/%s/http.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						files[apiHandlerFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))

						files[apiHandlerFile].Func().Id("MakeHttpHandler").Params(Id("r").Op("*").Qual("github.com/gorilla/mux", "Router"), Id("bus").Op("*").Qual("github.com/io-da/query", "Bus")).Params(Qual("net/http", "Handler")).BlockFunc(func(group *Group) {
							group.Id("queryHandler").Op(":=").Qual("github.com/terraskye/eventsourcing", "NewQueryGateway").Types(Op("*").Id("Query"), Id("ReadModel")).Call(Id("bus")).Line()
							group.Id("r").Dot("Methods").Call(Id("\"GET\"")).Dot("Path").Call(Id(fmt.Sprintf("\"%s/{id:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}}/%s\"", strings.ToLower(cfg.CodeGen.Domain), slug.Make(generator.ToCamelCase(slice.Title[7:]))))).Dot("Handler").Call(Qual("github.com/go-kit/kit/transport/http", "NewServer").Params(
								Func().Params(Id("ctx").Qual("context", "Context"), Id("request").Any()).Params(Any(), Error()).BlockFunc(func(group *Group) {

									group.List(Id("model"), Err()).Op(":=").Id("queryHandler").Dot("Query").Call(Id("ctx"), Id("request").Assert(Op("*").Id("Query")))

									// endpoint body
									group.If(Err().Op("!=").Nil()).Block(Return(Nil(), Err()))

									group.Return(Id("model").Dot("First").Call(), Nil())
								}),
								Id(fmt.Sprintf("decode%sRequest", generator.ToCamelCase(slice.Title[7:]))),
								Id(fmt.Sprintf("encode%sRequest", generator.ToCamelCase(slice.Title[7:]))),
							),
							)
							group.Return(Id("r"))
						})

						files[apiHandlerFile].Func().Id(fmt.Sprintf("decode%sRequest", generator.ToCamelCase(slice.Title[7:]))).Params(Id("ctx").Qual("context", "Context"), Id("r").Op("*").Qual("net/http", "Request")).Params(Interface(), Error()).BlockFunc(func(group *Group) {
							group.Comment("TODO decode query params onto Query object")
							group.Return(Nil(), Nil())
						})
						files[apiHandlerFile].Func().Id(fmt.Sprintf("encode%sRequest", generator.ToCamelCase(slice.Title[7:]))).Params(
							Id("ctx").Qual("context", "Context"), Id("writer").Qual("net/http", "ResponseWriter"), Id("i").Interface()).
							Params(Error()).
							BlockFunc(func(group *Group) {
								group.Id("writer").Dot("WriteHeader").Call(Qual("net/http", "StatusOK"))
								group.Comment("TODO map readmodel to http")
								group.Return(Nil())
							})

					}
				}

				fmt.Println("> processors")

				if automations := FilterFunc(slice.Processors, func(processor config.Processor) bool {
					return slices.ContainsFunc(processor.Dependencies, func(dependency config.Dependencies) bool {
						return dependency.Type == "INBOUND" && dependency.ElementType == "READMODEL"
					})
				}); len(automations) > 0 {
					fmt.Println("\t- automation")

					aggregateFile := fmt.Sprintf("%s/%s/automation.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))

						files[aggregateFile].ImportAlias("github.com/terraskye/eventsourcing", "cqrs")
						files[aggregateFile].ImportName("github.com/io-da/query", "query")
						files[aggregateFile].ImportName(commandPackage, "commands")
						files[aggregateFile].ImportName(eventsPackage, "events")

					}
					//files[aggregateFile].Type().Id("Automation").InterfaceFunc(func(group *Group) {
					//	for _, automation := range automations {
					//		for _, allSlices := range cfg.Slices {
					//			for _, readmodel := range allSlices.Readmodels {
					//				if readmodel.ListElement == true {
					//					// list elements are being provided on which items we should iterate over.
					//					// these might provide a list iteration of carts we need to clear.
					//					continue
					//				}
					//				var valid bool
					//				for _, dependency := range readmodel.Dependencies {
					//					if dependency.ID == automation.ID && dependency.Type == "OUTBOUND" {
					//						valid = true
					//					}
					//				}
					//
					//				if valid {
					//					for _, dependency := range readmodel.Dependencies {
					//						if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
					//							group.Add(Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(eventsPackage, generator.ToCamelCase(dependency.Title))).Params(Err().Error()))
					//						}
					//					}
					//				}
					//			}
					//		}
					//	}
					//})

					files[aggregateFile].Type().Id("automation").StructFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/terraskye/eventsourcing", "CommandBus"))

						for _, automation := range automations {
							for _, slices := range cfg.Slices {
								for _, readmodel := range slices.Readmodels {
									if readmodel.ListElement != true {
										continue
									}
									var valid bool

									for _, dependency := range readmodel.Dependencies {
										if dependency.ID == automation.ID && dependency.Type == "OUTBOUND" {
											valid = true
										}
									}
									if valid {
										readModelPackage, err := generator.ResolvePackagePath(fmt.Sprintf("%s/%s", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(readmodel.Title))))
										if err != nil {
											log.Fatal(err)
										}

										group.Add(Id(generator.LowercaseFirstCharacter(generator.ToCamelCase(readmodel.Title)))).Qual("github.com/terraskye/eventsourcing", "GenericQueryGateway").Types(Qual(readModelPackage, "Query"), Qual(readModelPackage, "ReadModel"))
									}
								}
							}
						}

						//for _, dependency := range automations[0].Dependencies {
						//	if dependency.Type == "INBOUND" && dependency.ElementType == "READMODEL" {
						//
						//		group.Add(Id(generator.ToCamelCase(dependency.Title))).Qual("github.com/terraskye/eventsourcing", "GenericQueryGateway").Types(Qual(readModelPackage, "Query"), Qual(readModelPackage, "ReadModel"))
						//
						//		// one of these is causing the trigger, and should be  removed.
						//		//group.Add(Id(generator.ToCamelCase(dependency.Title)).Func().Params(Id("ctx").Qual("context", "Context")))
						//	}
						//}

					})

					//TODO add querybus as well since we might need to lookup the TODO list.
					files[aggregateFile].Func().Id("NewAutomation").ParamsFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/terraskye/eventsourcing", "CommandBus"))
						group.Add(Id("queryBus").Op("*").Qual("github.com/io-da/query", "Bus"))

					}).Params(Op("*").Qual("github.com/terraskye/eventsourcing", "EventGroupProcessor")).Block(Id("a").Op(":= &").Id("automation").Block(DictFunc(func(dict Dict) {
						dict[Id("commandBus")] = Id("commandBus")

						for _, automation := range automations {
							for _, slices := range cfg.Slices {
								for _, readmodel := range slices.Readmodels {
									if readmodel.ListElement != true {
										continue
									}
									var valid bool

									for _, dependency := range readmodel.Dependencies {
										if dependency.ID == automation.ID && dependency.Type == "OUTBOUND" {
											valid = true
										}
									}
									if valid {
										readModelPackage, err := generator.ResolvePackagePath(fmt.Sprintf("%s/%s", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(readmodel.Title))))
										if err != nil {
											log.Fatal(err)
										}

										dict[Id(generator.LowercaseFirstCharacter(generator.ToCamelCase(readmodel.Title)))] = Qual("github.com/terraskye/eventsourcing", "NewQueryGateway").Types(Qual(readModelPackage, "Query"), Qual(readModelPackage, "ReadModel")).Call(Id("queryBus"))
									}
								}
							}
						}

					})), Return(Qual("github.com/terraskye/eventsourcing", "NewEventGroupProcessor").CallFunc(func(group *Group) {
						group.Add(Id(fmt.Sprintf("\"%s\"", generator.ToCamelCase(slice.Title[7:]))))

						for _, automation := range automations {
							for _, slices := range cfg.Slices {
								for _, readmodel := range slices.Readmodels {
									if readmodel.ListElement == true {
										//these are the todolist we need to fetch the iterator off.
										// list elements are being provided on which items we should iterate over.

										//fmt.Println(readmodel)
										continue
									}
									var valid bool
									for _, dependency := range readmodel.Dependencies {
										if dependency.ID == automation.ID && dependency.Type == "OUTBOUND" {
											valid = true
										}
									}

									if valid {
										for _, dependency := range readmodel.Dependencies {
											if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
												group.Add(Qual("github.com/terraskye/eventsourcing", "NewGroupEventHandler").Call(Id("a").Dot("On" + generator.ToCamelCase(dependency.Title))))
											}
										}
									}
								}
							}
						}

					})))

					for _, automation := range automations {
						for _, slices := range cfg.Slices {
							for _, readmodel := range slices.Readmodels {
								if readmodel.ListElement == true {
									// list elements are being provided on which items we should iterate over.
									continue
								}
								var valid bool
								for _, dependency := range readmodel.Dependencies {
									if dependency.ID == automation.ID && dependency.Type == "OUTBOUND" {
										valid = true
									}
								}

								if valid {
									for _, dependency := range readmodel.Dependencies {
										if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {

											// the ON<Event> function  -> command onto the command bus
											files[aggregateFile].Func().Params(Id("a").Op("*").Op("automation")).Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(eventsPackage, generator.ToCamelCase(dependency.Title))).Params(Error()).BlockFunc(func(group *Group) {
												group.Add(Comment("//TODO write logic here"))

												group.Add(Id("cmd").Op(":=").Op("&").Qual(commandPackage, generator.ToCamelCase(slice.Commands[0].Title)).Block(DictFunc(func(dict Dict) {
													for _, field := range slice.Commands[0].Fields {
														property := Id(generator.ToCamelCase(field.Name))
														if len(slice.Commands[0].Fields) > 1 {
															dict[property] = Id("ev").Dot(generator.ToCamelCase(field.Name))
														} else {
															dict[property] = Id("ev").Dot(generator.ToCamelCase(field.Name)).Op(",")
														}
													}
												})))
												group.If(Err().Op(":=").Id("a").Dot("commandBus").Dot("Send").Call(Id("ctx"), Id("cmd")).Op(";").Add(Err().Op("!=").Nil()).Block(Return().Err())).Line()
												group.Add(Return().Nil())
											})
										}
									}

								}
							}
						}
					}

				}

				if translators := FilterFunc(slice.Processors, func(processor config.Processor) bool {
					return !slices.ContainsFunc(processor.Dependencies, func(dependencie config.Dependencies) bool {
						return dependencie.Type == "INBOUND" && dependencie.ElementType == "READMODEL"
					})
				}); len(translators) > 0 {

					fmt.Println("\t- translator")

					aggregateFile := fmt.Sprintf("%s/%s/translator.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					}
					files[aggregateFile].Type().Id("Translator").InterfaceFunc(func(group *Group) {
						group.Add(Id("OnExternal").Params(Id("ctx").Qual("context", "Context"), Id("externalEvent").Id("any")).Params(Err().Error()))

					})

					files[aggregateFile].Type().Id("translator").StructFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/terraskye/eventsourcing", "CommandBus"))

					})

					files[aggregateFile].Func().Id("NewTranslator").ParamsFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/terraskye/eventsourcing", "CommandBus"))
					}).Params(Id("Translator")).Block(Return().Op("&").Id("translator").Block(DictFunc(func(dict Dict) {
						dict[Id("commandBus")] = Id("commandBus").Op(",")
					})))

					files[aggregateFile].Func().Params(Id("a").Op("*").Op("automation")).Id("OnExternal").Params(Id("ctx").Qual("context", "Context"), Id("externalEvent").Id("any")).Params(Error()).BlockFunc(func(group *Group) {
						group.Add(Id("cmd").Op(":=").Op("&").Qual(commandPackage, generator.ToCamelCase(slice.Commands[0].Title)).Block(DictFunc(func(dict Dict) {
							for _, field := range slice.Commands[0].Fields {
								property := Id(generator.ToCamelCase(field.Name))
								if len(slice.Commands[0].Fields) > 1 {
									dict[property] = Id("externalEvent").Dot(generator.ToCamelCase(field.Name))
								} else {
									dict[property] = Id("externalEvent").Dot(generator.ToCamelCase(field.Name)).Op(",")
								}
							}
						})))
						group.If(Err().Op(":=").Id("a").Dot("commandBus").Dot("Send").Call(Id("ctx"), Id("cmd")).Op(";").Add(Err().Op("!=").Nil()).Block(Return().Err())).Line()
						group.Add(Return().Nil())
					})
					//for _, translator := range translators {
					//	fmt.Println("\t- translator")
					//}

				}
			}
		}
	}

	for path, template := range files {
		os.MkdirAll(filepath.Dir(path), os.FileMode(0775))
		f, _ := os.Create(path)

		if err := template.Render(f); err != nil {
			log.Fatal(err.Error())
			io.WriteString(f, err.Error())
		}

	}

}

// FilterFunc filters elements in a slice based on the given function.
func FilterFunc[S ~[]E, E any](s S, f func(E) bool) S {
	for i := len(s) - 1; i >= 0; i-- {
		if !f(s[i]) {
			s = append(s[:i], s[i+1:]...)
		}
	}
	return s
}
