package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"log"
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

	var slices = make([]string, len(config.Slices))

	for i, slice := range config.Slices {
		slices[i] = slice.Title
	}
	slicesToGenerate := Checkboxes(
		"Wich slice would you like to generate", slices,
	)
	//fmt.Println(slicesToGenerate)

	for _, sliceName := range slicesToGenerate {

		fmt.Println(sliceName)
		for _, slice := range config.Slices {
			if slice.Title == sliceName {

				fmt.Println("=== commands ===")
				for _, command := range slice.Commands {
					fmt.Println(command.Title)
				}

				fmt.Println("=== events ===")
				for _, event := range slice.Events {
					fmt.Println(event.Title)
				}
				fmt.Println("=== readmodel ===")
				for _, readmodel := range slice.Readmodels {
					fmt.Println(readmodel.Title)
				}
				fmt.Println("=== aggregates ===")

				for _, aggregate := range slice.Aggregates {
					fmt.Println(aggregate)
				}

				//fmt.Printf("commands : %s")
				//slice.Commands
				//fmt.Println(slice)

			}
		}

	}

}
