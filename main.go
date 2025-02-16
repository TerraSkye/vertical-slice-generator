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
		for _, aggregate := range config.Aggregates {
			aggregateFile := fmt.Sprintf("gen/%s/domain/%s.go", config.CodeGen.Domain, strings.ToLower(aggregate.Title))
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

		for _, slice := range config.Slices {

			for _, event := range slice.Events {
				aggregateFile := fmt.Sprintf("gen/%s/events/%s.go", config.CodeGen.Domain, strings.ToLower(generator.ToCamelCase(event.Title)))
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

			for _, command := range slice.Commands {
				aggregateFile := fmt.Sprintf("gen/%s/domain/commands/%s.go", config.CodeGen.Domain, strings.ToLower(generator.ToCamelCase(command.Title)))
				commandPackage, err := generator.ResolvePackagePath(aggregateFile)
				if err != nil {
					log.Fatal(err)
				}

				eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/events/%s.go", config.CodeGen.Domain, strings.ToLower(generator.ToCamelCase(command.Title))))
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
				files[fmt.Sprintf("gen/%s/domain/%s.go", config.CodeGen.Domain, strings.ToLower(command.Aggregate))].ImportName(filepath.Dir(commandPackage), "commands")
				files[fmt.Sprintf("gen/%s/domain/%s.go", config.CodeGen.Domain, strings.ToLower(command.Aggregate))].ImportName(filepath.Dir(eventsPackage), "events")
				files[fmt.Sprintf("gen/%s/domain/%s.go", config.CodeGen.Domain, strings.ToLower(command.Aggregate))].
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
													if field.Cardinality != "Single" {
														property = property.Index()
													}

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

				//fmt.Println(filepath.Dir(commandPackage), err)
				//files[aggregateFile] = generator.GetFile("domain")

				//aggregateFile :=

				//fmt.Println(command.Aggregate)
				//fmt.Println(command)

			}
		}
	}

	for _, slice := range config.Slices {
		aggregateFile := fmt.Sprintf("gen/%s/%s/service.go", config.CodeGen.Domain, strings.ToLower(generator.ToCamelCase(slice.Title[7:])))
		files[aggregateFile] = generator.GetFile("domain")
		//files[aggregateFile].Type().Id(generator.ToCamelCase(slice.Title)).StructFunc(func(group *Group) {
		//
		//})

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
								eventsPackage, err := generator.ResolvePackagePath(fmt.Sprintf("gen/%s/events/%s.go", config.CodeGen.Domain, strings.ToLower(generator.ToCamelCase(event.Title))))
								if err != nil {
									log.Fatal(err)
								}
								files[aggregateFile].ImportName(filepath.Dir(eventsPackage), "events")
								files[aggregateFile].Line().Func().Params(Id("r").Op("*").Id(generator.ToCamelCase(readModel.Title)+"ReadModel")).
									Id("On"+generator.ToCamelCase(dependency.Title)).Params(Id("ctx").Qual("context", "Context"), Id("ev").Qual(filepath.Dir(eventsPackage), generator.ToCamelCase(dependency.Title))).Params(Error()).BlockFunc(func(group *Group) {
									for _, field := range event.Fields {
										group.Add(Id("r").Dot(generator.ToCamelCase(field.Name)).Op("=").Id("ev").Dot(generator.ToCamelCase(field.Name)))
									}
									group.Return().Nil()
								})
							}
						}
					}
				}
			}

		}
	}

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
