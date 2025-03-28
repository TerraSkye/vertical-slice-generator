package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/generator"
	"github.com/terraskye/vertical-slice-generator/generator/config"
	"github.com/terraskye/vertical-slice-generator/generator/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
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
	for _, sliceName := range slicesToGenerate {
		for _, slice := range cfg.Slices {
			if slice.Title == sliceName {

				eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/events", strings.ToLower(cfg.CodeGen.Domain)))
				if err != nil {
					log.Fatal(err)
				}

				commandPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/domain/commands", strings.ToLower(cfg.CodeGen.Domain)))
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
							aggregateFile := fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(aggregate.Title))

							if _, ok := files[aggregateFile]; !ok {
								files[aggregateFile] = generator.GetFile("domain")
								files[aggregateFile].ImportAlias("", eventsPackage)
								files[aggregateFile].ImportAlias(commandPackage, "commands")
							}

							files[aggregateFile].Type().Id(generator.ToCamelCase(aggregate.Title)).Add(template.FieldsStruct(aggregate.Fields))
						}
					}

					//fmt.Println("\t- " + aggregate)
				}

				fmt.Println("> commands")
				for _, command := range slice.Commands {
					aggregateFile := fmt.Sprintf("gen/%s/domain/commands/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title)))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile("commands")
					}
					files[aggregateFile].Type().Id(generator.ToCamelCase(command.Title)).Add(template.FieldsStruct(command.Fields))

					files[fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(command.Aggregate))].
						Func().Params(Id(strings.ToLower(string(command.Aggregate[0]))).Op("*").Id(generator.ToCamelCase(command.Aggregate))).
						Id(generator.ToCamelCase(command.Title)).
						Params(Id("ctx").Qual("context", "Context"), Id("cmd").Op("*").Qual(commandPackage, generator.ToCamelCase(command.Title))).
						Params(Error()).BlockFunc(func(body *Group) {

						for _, dependency := range command.Dependencies {
							if dependency.Type == "OUTBOUND" && dependency.ElementType == "EVENT" {
								for _, slices := range cfg.Slices {
									for _, event := range slices.Events {
										if event.ID == dependency.ID {
											// outbound event.

											body.Id(strings.ToLower(string(command.Aggregate[0]))).Dot("AggregateSlice").Dot("AppendEvent").
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

					//handlerFile := fmt.Sprintf("gen/%s/handlers/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title)))
					//
					//if _, ok := files[handlerFile]; !ok {
					//	files[handlerFile] = generator.GetFile("handlers")
					//}

					//files[handlerFile].Func().Id("init").Call().Block(
					//	Qual("github.com/terraskye/vertical-slice-implementation/cart/infrastructure", "RegisterCommand").Params(
					//		Func().Params(Id("aggregate").Op("*").Qual("domain", generator.ToCamelCase(command.Aggregate))).
					//			Func().Params(Id("ctx").Qual("context", "Context"),
					//			Id("cmd").Op("*").Qual(commandPackage, generator.ToCamelCase(command.Title))).Params(Error()).Block(
					//			Return(Func().Params(Id("ctx").Qual("context", "Context"),
					//				Id("cmd").Op("*").Qual(commandPackage, generator.ToCamelCase(command.Title))).Params(Error()).Block(
					//				Return(Id("aggregate").Dot(generator.ToCamelCase(command.Title)).Call(Id("ctx"), Id("cmd"))),
					//			)),
					//		),
					//	),
					//)

				}
				fmt.Println("> events")
				for _, event := range slice.Events {
					fmt.Println("\t- " + event.Title)
					aggregateFile := fmt.Sprintf("gen/%s/events/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title)))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile("events")

					}
					files[aggregateFile].Type().Id(generator.ToCamelCase(event.Title)).Add(template.FieldsStruct(event.Fields))

					files[fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(event.Aggregate))].
						Func().Params(Id(strings.ToLower(string(event.Aggregate[0]))).Op("*").Id(generator.ToCamelCase(event.Aggregate))).
						Id(generator.ToCamelCase("On"+event.Title)).
						Params(Id("ctx").Qual("context", "Context"), Id("cmd").Op("*").Qual(eventsPackage, generator.ToCamelCase(event.Title))).
						Params().BlockFunc(func(body *Group) {

						//body.Return(Nil())
					}).Line()

					handlerFile := fmt.Sprintf("gen/%s/handlers/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title)))

					if _, ok := files[handlerFile]; !ok {
						files[handlerFile] = generator.GetFile("handlers")
					}

					files[handlerFile].Func().Id("init").Call().Block(
						Qual("github.com/terraskye/vertical-slice-implementation/cart/infrastructure", "RegisterEvent").Params(
							Func().Params(Id("aggregate").Op("*").Qual("domain", generator.ToCamelCase(event.Aggregate))).
								Func().Params(Id("event").Op("*").Qual(eventsPackage, generator.ToCamelCase(event.Title))).Block(
								Return(Func().Params(Id("event").Op("*").Qual(eventsPackage, generator.ToCamelCase(event.Title)))).Block(
									Id("aggregate").Dot(generator.ToCamelCase("On" + event.Title)).Call(Id("event"))),
							),
						),
					)

				}

				if commands := slice.Commands; len(commands) > 0 {
					fmt.Println("> state change")
					aggregateFile := fmt.Sprintf("gen/%s/%s/service.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					}
					files[aggregateFile].Type().Id("Service").Interface(Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()))

					for _, command := range commands {
						fmt.Println("\t- " + command.Title)
						files[aggregateFile].Line().Type().Id("Payload").Add(template.FieldsStruct(command.Fields))
					}

					files[aggregateFile].Type().Id("service").StructFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
					})

					files[aggregateFile].Func().Id("NewService").ParamsFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
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
				}

				if readModels := slice.Readmodels; len(readModels) > 0 {
					fmt.Println("> state view")

					for _, readModel := range readModels {
						fmt.Println("\t- " + readModel.Title)
						aggregateFile := fmt.Sprintf("gen/%s/%s/readmodel.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						if _, ok := files[aggregateFile]; !ok {
							files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						}
						files[aggregateFile].Type().Id("Query").Struct()
						files[aggregateFile].Line().Type().Id(generator.ToCamelCase(readModel.Title) + "ReadModel").Add(template.FieldsStruct(readModel.Fields))
						for _, dependency := range readModel.Dependencies {
							if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
								for _, slices := range cfg.Slices {
									for _, event := range slices.Events {
										if event.ID == dependency.ID {
											// outbound event.

											files[aggregateFile].ImportName(eventsPackage, "events")
											files[aggregateFile].Line().Func().Params(Id("p").Op("*").Id("projector")).
												Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(filepath.Dir(eventsPackage), generator.ToCamelCase(dependency.Title))).Params(Error()).BlockFunc(func(group *Group) {

												group.Return().Nil()
											})
										}
									}
								}
							}
						}

						projectorFile := fmt.Sprintf("gen/%s/%s/projector.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						files[projectorFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
						files[projectorFile].Type().Id("Projecter").Interface().Line()
						files[projectorFile].Type().Id("projector").Struct(
							Id("entities").Map(Qual("github.com/google/uuid", "UUID")).Op("*").Id(generator.ToCamelCase(slice.Readmodels[0].Title) + "ReadModel"),
						)
						files[projectorFile].Func().Id("NewProjector").ParamsFunc(func(group *Group) {
							//group.Add(Id("eventbus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "EventBus"))
						}).Params(Id("Projecter")).Block(Return().Op("&").Id("projector").Block(DictFunc(func(dict Dict) {
							//dict[Id("eventbus")] = Id("eventbus").Op(",")
						})))

						files[projectorFile].Line().Func().Params(Id("p").Op("*").Id("projector")).
							Id("HandleQuery").Params(Id("ctx").Qual("context", "Context"), Id("q").Id("Query")).Params(Op("*").Id(generator.ToCamelCase(slice.Readmodels[0].Title)+"ReadModel"), Error()).BlockFunc(func(group *Group) {

							group.Return(Nil(), Nil())
						})

					}
				}

				fmt.Println("> processors")

				if automations := FilterFunc(slice.Processors, func(processor config.Processor) bool {
					return slices.ContainsFunc(processor.Dependencies, func(dependencie config.Dependencies) bool {
						return dependencie.Type == "INBOUND" && dependencie.ElementType == "READMODEL"
					})
				}); len(automations) > 0 {
					fmt.Println("\t- automation")

					aggregateFile := fmt.Sprintf("gen/%s/%s/automation.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					}
					files[aggregateFile].Type().Id("Automation").InterfaceFunc(func(group *Group) {
						for _, automation := range automations {
							for _, allSlices := range cfg.Slices {
								for _, readmodel := range allSlices.Readmodels {
									if readmodel.ListElement == true {
										// list elements are being provided on which items we should iterate over.
										// these might provide a list iteration of carts we need to clear.
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
												group.Add(Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(eventsPackage, generator.ToCamelCase(dependency.Title))).Params(Err().Error()))
											}
										}
									}
								}
							}
						}

					})

					files[aggregateFile].Type().Id("automation").StructFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))

						for _, dependency := range automations[0].Dependencies {
							if dependency.Type == "INBOUND" && dependency.ElementType == "READMODEL" {
								group.Add(Id(generator.ToCamelCase(dependency.Title)).Func().Params(Id("ctx").Qual("context", "Context")))
							}
						}

					})

					files[aggregateFile].Func().Id("NewAutomation").ParamsFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
					}).Params(Id("Automation")).Block(Return().Op("&").Id("automation").Block(DictFunc(func(dict Dict) {
						dict[Id("commandBus")] = Id("commandBus").Op(",")
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

					aggregateFile := fmt.Sprintf("gen/%s/%s/translator.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					if _, ok := files[aggregateFile]; !ok {
						files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
					}
					files[aggregateFile].Type().Id("Translator").InterfaceFunc(func(group *Group) {
						group.Add(Id("OnExternal").Params(Id("ctx").Qual("context", "Context"), Id("externalEvent").Id("any")).Params(Err().Error()))

					})

					files[aggregateFile].Type().Id("translator").StructFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))

					})

					files[aggregateFile].Func().Id("NewTranslator").ParamsFunc(func(group *Group) {
						group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
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
