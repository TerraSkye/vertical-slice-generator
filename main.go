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
	//go:embed config.json
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

	slicesToGenerate := []string{"slice: Create Quiz"}
	if slices.Contains(slicesToGenerate, "all") {
		slicesToGenerate = tasks[1:]
	}

	var files = make(map[string]*File)

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
						if aggregate.Title == aggregateName {
							aggregateFile := fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(aggregate.Title))
							files[aggregateFile] = generator.GetFile("domain")
							files[aggregateFile].Type().Id(generator.ToCamelCase(aggregate.Title)).Add(template.FieldsStruct(aggregate.Fields))
						}
					}

					//fmt.Println("\t- " + aggregate)
				}

				fmt.Println("> commands")
				for _, command := range slice.Commands {
					aggregateFile := fmt.Sprintf("gen/%s/domain/commands/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(command.Title)))
					files[aggregateFile] = generator.GetFile("commands")
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
				}
				fmt.Println("> events")
				for _, event := range slice.Events {
					fmt.Println("\t- " + event.Title)
					aggregateFile := fmt.Sprintf("gen/%s/events/%s.go", strings.ToLower(cfg.CodeGen.Domain), strings.ToLower(generator.ToCamelCase(event.Title)))
					files[aggregateFile] = generator.GetFile("events")
					files[aggregateFile].Type().Id(generator.ToCamelCase(event.Title)).Add(template.FieldsStruct(event.Fields))
				}

				fmt.Println("> state change")
				for _, command := range slice.Commands {
					fmt.Println("\t- " + command.Title)
				}

				fmt.Println("> state view")
				for _, readModel := range slice.Readmodels {
					fmt.Println("\t- " + readModel.Title)
				}

				fmt.Println("> processors")
				for _, processor := range slice.Processors {
					if slices.ContainsFunc(processor.Dependencies, func(dependencie config.Dependencies) bool {
						return dependencie.Type == "INBOUND" && dependencie.ElementType == "READMODEL"
					}) {
						fmt.Println("\t- automation")
					}
				}
				fmt.Println("> commands")
				for _, processor := range slice.Processors {
					if !slices.ContainsFunc(processor.Dependencies, func(dependencie config.Dependencies) bool {
						return dependencie.Type == "INBOUND" && dependencie.ElementType == "READMODEL"
					}) {
						fmt.Println("\t- translator")
					}

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
