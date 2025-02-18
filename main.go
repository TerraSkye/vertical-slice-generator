package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/generator"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	//go:embed config.json
	configuration []byte
)

func main() {

	//os.RemoveAll("gen")
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		log.Fatal(err)
	}

	var files = make(map[string]*File)

	{
		// generate all aggregates
		for _, aggregate := range config.Aggregates {
			aggregateFile := fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(aggregate.Title))
			files[aggregateFile] = generator.GetFile("domain")
			files[aggregateFile].Type().Id(generator.ToCamelCase(aggregate.Title)).StructFunc(func(group *Group) {
				for _, field := range aggregate.Fields {
					property := Id(generator.ToCamelCase(field.Name))
					if field.Cardinality != "Single" {
						property = property.Index()
					}
					switch field.Type {
					case "String":
						property = property.String()
					case "UUID":
						property = property.Qual("github.com/google/uuid", "UUID")
					case "Boolean":
						property = property.Bool()
					case "Double":
						property = property.Float64()
					case "Date":
						property = property.Qual("time", "Time")
					case "DateTime":
						property = property.Qual("time", "Time")
					case "Long":
						property = property.Int64()
					case "Int":
						property = property.Int()
					case "Custom":
						property = property.Interface()
					}

					group.Add(property)
				}

			})

			//log.Printf("Aggregate #%d: %s", i+1, slice)
		}
	}

	{

		//generate all events
		for _, slice := range config.Slices {
			if slice.Context == "EXTERNAL" {
				continue
			}

			for _, event := range slice.Events {
				aggregateFile := fmt.Sprintf("gen/%s/events/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title)))
				files[aggregateFile] = generator.GetFile("events")
				files[aggregateFile].Type().Id(generator.ToCamelCase(event.Title)).StructFunc(func(group *Group) {
					// TODO generate struct fields
					for _, field := range event.Fields {
						property := Id(generator.ToCamelCase(field.Name))
						if field.Cardinality != "Single" {
							property = property.Index()
						}
						switch field.Type {
						case "String":
							property = property.String()
						case "UUID":
							property = property.Qual("github.com/google/uuid", "UUID")
						case "Boolean":
							property = property.Bool()
						case "Double":
							property = property.Float64()
						case "Date":
							property = property.Qual("time", "Time")
						case "DateTime":
							property = property.Qual("time", "Time")
						case "Long":
							property = property.Int64()
						case "Int":
							property = property.Int()
						case "Custom":
							property = property.Interface()
						}

						group.Add(property)
					}

				})

				//fmt.Println(event)

			}
		}
	}

	{

		for _, slice := range config.Slices {

			//generate all commands
			for _, command := range slice.Commands {
				aggregateFile := fmt.Sprintf("gen/%s/domain/commands/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title)))
				commandPackage, err := generator.ResolvePackagePath(aggregateFile)
				if err != nil {
					log.Fatal(err)
				}

				eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/events/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title))))
				if err != nil {
					log.Fatal(err)
				}
				files[aggregateFile] = generator.GetFile("commands")
				files[aggregateFile].Type().Id(generator.ToCamelCase(command.Title)).StructFunc(func(group *Group) {

					for _, field := range command.Fields {

						property := Id(generator.ToCamelCase(field.Name))

						if field.Cardinality != "Single" {
							property = property.Index()
						}

						switch field.Type {
						case "String":
							property = property.String()
						case "UUID":
							property = property.Qual("github.com/google/uuid", "UUID")
						case "Boolean":
							property = property.Bool()
						case "Double":
							property = property.Float64()
						case "Date":
							property = property.Qual("time", "Time")
						case "DateTime":
							property = property.Qual("time", "Time")
						case "Long":
							property = property.Int64()
						case "Int":
							property = property.Int()
						case "Custom":
							property = property.Interface()
						}

						group.Add(property)
					}

				})

				// we register the command as a function within the aggregate
				files[fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(command.Aggregate))].ImportName(filepath.Dir(commandPackage), "commands")
				files[fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(command.Aggregate))].ImportName(filepath.Dir(eventsPackage), "events")

				files[fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(command.Aggregate))].
					Func().Params(Id(strings.ToLower(string(command.Aggregate[0]))).Op("*").Id(generator.ToCamelCase(command.Aggregate))).
					Id(generator.ToCamelCase(command.Title)).
					Params(Id("ctx").Qual("context", "Context"), Id("cmd").Op("*").Qual(filepath.Dir(commandPackage), generator.ToCamelCase(command.Title))).
					Params(Error()).BlockFunc(func(body *Group) {

					for _, dependency := range command.Dependencies {
						if dependency.Type == "OUTBOUND" && dependency.ElementType == "EVENT" {
							for _, slices := range config.Slices {
								for _, event := range slices.Events {
									if event.ID == dependency.ID {
										// outbound event.

										body.Id(strings.ToLower(string(command.Aggregate[0]))).Dot("AggregateSlice").Dot("AppendEvent").
											Call(Id("ctx"), Op("&").Qual(filepath.Dir(eventsPackage), generator.ToCamelCase(event.Title)).Block(DictFunc(func(dict Dict) {
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

					body.Return().Nil()
				})

			}
		}
	}

	// generate all command handlers.
	for _, slice := range config.Slices {
		if len(slice.Commands) == 0 || true {
			continue
		}

		aggregateFile := fmt.Sprintf("gen/%s/%s/service.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
		files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))

		files[aggregateFile].Type().Id("Service").Interface(Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()))

		for _, command := range slice.Commands {
			files[aggregateFile].Line().Type().Id("Payload").StructFunc(func(group *Group) {
				for _, field := range command.Fields {
					property := Id(generator.ToCamelCase(field.Name))
					if field.Cardinality != "Single" {
						property = property.Index()
					}
					switch field.Type {
					case "String":
						property = property.String()
					case "UUID":
						property = property.Qual("github.com/google/uuid", "UUID")
					case "Boolean":
						property = property.Bool()
					case "Double":
						property = property.Float64()
					case "Date":
						property = property.Qual("time", "Time")
					case "DateTime":
						property = property.Qual("time", "Time")
					case "Long":
						property = property.Int64()
					case "Int":
						property = property.Int()
					case "Custom":
						property = property.Interface()
					}

					group.Add(property)
				}
			})
		}

		files[aggregateFile].Type().Id("service").StructFunc(func(group *Group) {
			group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
		})

		files[aggregateFile].Func().Id("New").ParamsFunc(func(group *Group) {
			group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
		}).Params(Id("Service")).Block(Return().Op("&").Id("service").Block(DictFunc(func(dict Dict) {
			dict[Id("commandBus")] = Id("commandBus").Op(",")
		})))

		commandFile := fmt.Sprintf("gen/%s/domain/commands/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Commands[0].Title)))
		commandPackage, err := generator.ResolvePackagePath(commandFile)
		if err != nil {
			log.Fatal(err)
		}
		files[aggregateFile].ImportAlias(filepath.Dir(commandPackage), "")
		files[aggregateFile].Func().Params(Id("s").Op("*").Op("service")).Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()).BlockFunc(func(group *Group) {
			group.Add(Id("cmd").Op(":=").Op("&").Qual(filepath.Dir(commandPackage), generator.ToCamelCase(slice.Commands[0].Title)).Block(DictFunc(func(dict Dict) {
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

	// generate all query handlers.
	for _, slice := range config.Slices {
		if len(slice.Readmodels) == 0 || true {
			continue
		}

		// generate the projector file
		projectorFile := fmt.Sprintf("gen/%s/%s/projector.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
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

		//files[projectorFile].Type().Id("Service").Interface(Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()))
		aggregateFile := fmt.Sprintf("gen/%s/%s/readmodel.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
		files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
		files[aggregateFile].Type().Id("Query").Struct()

		for _, readModel := range slice.Readmodels {
			files[aggregateFile].Line().Type().Id(generator.ToCamelCase(readModel.Title) + "ReadModel").StructFunc(func(group *Group) {
				for _, field := range readModel.Fields {

					property := Id(generator.ToCamelCase(field.Name))

					if field.Cardinality != "Single" {
						property = property.Index()
					}

					switch field.Type {
					case "String":
						property = property.String()
					case "UUID":
						property = property.Qual("github.com/google/uuid", "UUID")
					case "Boolean":
						property = property.Bool()
					case "Double":
						property = property.Float64()
					case "Date":
						property = property.Qual("time", "Time")
					case "DateTime":
						property = property.Qual("time", "Time")
					case "Long":
						property = property.Int64()
					case "Int":
						property = property.Int()
					case "Custom":
						property = property.Interface()
					}

					group.Add(property)
				}
			})

			for _, dependency := range readModel.Dependencies {
				if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
					for _, slices := range config.Slices {
						for _, event := range slices.Events {
							if event.ID == dependency.ID {
								// outbound event.
								eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/events/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title))))
								if err != nil {
									log.Fatal(err)
								}
								files[projectorFile].ImportName(filepath.Dir(eventsPackage), "events")
								files[projectorFile].Line().Func().Params(Id("p").Op("*").Id("projector")).
									Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(filepath.Dir(eventsPackage), generator.ToCamelCase(dependency.Title))).Params(Error()).BlockFunc(func(group *Group) {

									group.Return().Nil()
								})
							}
						}
					}
				}
			}
		}
	}

	//generate processors
	for _, slice := range config.Slices {
		if len(slice.Processors) == 0 {
			continue
		}

		aggregateFile := fmt.Sprintf("gen/%s/%s/service.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
		files[aggregateFile] = generator.GetFile(strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
		files[aggregateFile].Type().Id("Service").Interface(Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()))

		for _, command := range slice.Commands {
			files[aggregateFile].Line().Type().Id("Payload").StructFunc(func(group *Group) {
				for _, field := range command.Fields {

					property := Id(generator.ToCamelCase(field.Name))

					if field.Cardinality != "Single" {
						property = property.Index()
					}

					switch field.Type {
					case "String":
						property = property.String()
					case "UUID":
						property = property.Qual("github.com/google/uuid", "UUID")
					case "Boolean":
						property = property.Bool()
					case "Double":
						property = property.Float64()
					case "Date":
						property = property.Qual("time", "Time")
					case "DateTime":
						property = property.Qual("time", "Time")
					case "Long":
						property = property.Int64()
					case "Int":
						property = property.Int()
					case "Custom":
						property = property.Interface()
					}

					group.Add(property)
				}
			})
		}

		//  todo check if we have commands, we need to include thye commandbus
		files[aggregateFile].Type().Id("service").StructFunc(func(group *Group) {
			if len(slice.Commands) > 0 {
				group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
			}
		})

		files[aggregateFile].Func().Id("New").ParamsFunc(func(group *Group) {
			if len(slice.Commands) > 0 {
				group.Add(Id("commandBus").Qual("github.com/ThreeDotsLabs/watermill/components/cqrs", "CommandBus"))
			}
		}).Params(Id("Service")).Block(Return().Op("&").Id("service").Block(DictFunc(func(dict Dict) {
			if len(slice.Commands) > 0 {
				dict[Id("commandBus")] = Id("commandBus").Op(",")
			}
		})))

		commandFile := fmt.Sprintf("gen/%s/domain/commands/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(slice.Commands[0].Title)))
		commandPackage, err := generator.ResolvePackagePath(commandFile)
		if err != nil {
			log.Fatal(err)
		}
		files[aggregateFile].ImportAlias(filepath.Dir(commandPackage), "")
		files[aggregateFile].Func().Params(Id("s").Op("*").Op("service")).Id(generator.ToCamelCase(slice.Title[7:])).Params(Id("ctx").Qual("context", "Context"), Id("payload").Qual("", "Payload")).Params(Error()).BlockFunc(func(group *Group) {
			group.Add(Id("cmd").Op(":=").Op("&").Qual(filepath.Dir(commandPackage), generator.ToCamelCase(slice.Commands[0].Title)).Block(DictFunc(func(dict Dict) {
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

		for _, command := range slice.Commands {
			for _, dependency := range command.Dependencies {
				if dependency.Type == "INBOUND" {
					fmt.Println(dependency)
				}
			}
		}

		for _, readModel := range slice.Readmodels {
			fmt.Println(readModel.Title)
		}
		//	files[aggregateFile].Line().Type().Id(generator.ToCamelCase(readModel.Title) + "ReadModel").StructFunc(func(group *Group) {
		//		for _, field := range readModel.Fields {
		//
		//			property := Id(generator.ToCamelCase(field.Name))
		//
		//			if field.Cardinality != "Single" {
		//				property = property.Index()
		//			}
		//
		//			switch field.Type {
		//			case "String":
		//				property = property.String()
		//			case "UUID":
		//				property = property.Qual("github.com/google/uuid", "UUID")
		//			case "Boolean":
		//				property = property.Bool()
		//			case "Double":
		//				property = property.Float64()
		//			case "Date":
		//				property = property.Qual("time", "Time")
		//			case "DateTime":
		//				property = property.Qual("time", "Time")
		//			case "Long":
		//				property = property.Int64()
		//			case "Int":
		//				property = property.Int()
		//			case "Custom":
		//				property = property.Interface()
		//			}
		//
		//			group.Add(property)
		//		}
		//	})
		//
		//	for _, dependency := range readModel.Dependencies {
		//		if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
		//			for _, slices := range config.Slices {
		//				for _, event := range slices.Events {
		//					if event.ID == dependency.ID {
		//						// outbound event.
		//						eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/events/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title))))
		//						if err != nil {
		//							log.Fatal(err)
		//						}
		//						files[aggregateFile].ImportName(filepath.Dir(eventsPackage), "events")
		//						files[aggregateFile].Line().Func().Params(Id("r").Op("*").Id(generator.ToCamelCase(readModel.Title)+"ReadModel")).
		//							Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Op("*").Qual(filepath.Dir(eventsPackage), generator.ToCamelCase(dependency.Title))).Params(Error()).BlockFunc(func(group *Group) {
		//							for _, field := range event.Fields {
		//								group.Add(Id("r").Dot(generator.ToCamelCase(field.Name)).Op("=").Id("ev").Dot(generator.ToCamelCase(field.Name)))
		//							}
		//							group.Return().Nil()
		//						})
		//					}
		//				}
		//			}
		//		}
		//	}
		//
		//}

		//for i, slice := range config.Slices {
		//	for i2, screen := range slice.Screens {
		//
		//	}
		//}

		//for _, processor := range slice.Processors {
		//	files[aggregateFile].Line().Type().Id(generator.ToCamelCase(processor.Title)).StructFunc(func(group *Group) {
		//		for _, field := range processor.Fields {
		//
		//			property := Id(generator.ToCamelCase(field.Name))
		//
		//			if field.Cardinality != "Single" {
		//				property = property.Index()
		//			}
		//
		//			switch field.Type {
		//			case "String":
		//				property = property.String()
		//			case "UUID":
		//				property = property.Qual("github.com/google/uuid", "UUID")
		//			case "Boolean":
		//				property = property.Bool()
		//			case "Double":
		//				property = property.Float64()
		//			case "Date":
		//				property = property.Qual("time", "Time")
		//			case "DateTime":
		//				property = property.Qual("time", "Time")
		//			case "Long":
		//				property = property.Int64()
		//			case "Int":
		//				property = property.Int()
		//			case "Custom":
		//				property = property.Interface()
		//			}
		//
		//			group.Add(property)
		//		}
		//	})
		//
		//}
	}

	//log.Printf("Aggregate #%d: %s", i+1, slice)

	for path, template := range files {
		os.MkdirAll(filepath.Dir(path), os.FileMode(0775))
		f, _ := os.Create(path)

		if err := template.Render(f); err != nil {
			log.Fatal(err.Error())
			io.WriteString(f, err.Error())
		}

		f.Close()

	}

}
