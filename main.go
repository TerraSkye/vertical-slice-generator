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
		"Which slice would you like to generate", slices,
	)
	fmt.Println(slicesToGenerate)

}
