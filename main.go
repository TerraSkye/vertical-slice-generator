package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"log"
	"slices"
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
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		log.Fatal(err)
	}

	var tasks = make([]string, len(config.Slices))

	for i, slice := range config.Slices {
		tasks[i] = slice.Title
	}
	slicesToGenerate := Checkboxes(
		"Which slice would you like to generate", append([]string{"all"}, tasks...),
	)
	if slices.Contains(slicesToGenerate, "all") {
		slicesToGenerate = tasks[1:]
	}

	for _, sliceName := range slicesToGenerate {
		for _, slice := range config.Slices {
			if slice.Title == sliceName {
				fmt.Printf("===== [slice: %s] =====\n", slice.Title)

				fmt.Println("> aggregate")
				for _, aggregate := range slice.Aggregates {
					fmt.Println("\t- " + aggregate)
				}

				fmt.Println("> commands")
				for _, command := range slice.Commands {
					fmt.Println("\t- " + command.Title)
				}
				fmt.Println("> events")
				for _, events := range slice.Events {
					fmt.Println("\t- " + events.Title)
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
					if slices.ContainsFunc(processor.Dependencies, func(dependencie Dependencies) bool {
						return dependencie.Type == "INBOUND" && dependencie.ElementType == "READMODEL"
					}) {
						fmt.Println("\t- automation")
					}
				}
				fmt.Println("> commands")
				for _, processor := range slice.Processors {
					if !slices.ContainsFunc(processor.Dependencies, func(dependencie Dependencies) bool {
						return dependencie.Type == "INBOUND" && dependencie.ElementType == "READMODEL"
					}) {
						fmt.Println("\t- translator")
					}

				}

			}
		}

	}

}
