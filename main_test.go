package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"
)

func Test_Automations(t *testing.T) {

	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		t.Fatal(err)
	}

	for _, slice := range config.Slices {
		for _, command := range slice.Commands {
			var automationID string
			var eventIds []string
			for _, dependency := range command.Dependencies {
				if dependency.ElementType == "AUTOMATION" {
					automationID = dependency.ID
				}
			}

			if automationID != "" {

				for _, slices := range config.Slices {
					for _, readmodel := range slices.Readmodels {
						if readmodel.ListElement == true {
							// list elements are being provided on which items we should iterate over.
							continue
						}
						var valid bool
						for _, dependency := range readmodel.Dependencies {
							if dependency.ID == automationID && dependency.Type == "OUTBOUND" {
								valid = true
							}
						}

						if valid {
							for _, dependency := range readmodel.Dependencies {
								if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
									fmt.Println("-- EVENT :" + dependency.Title)
									eventIds = append(eventIds, dependency.ID)
								}
							}

						}
					}
				}

				if len(eventIds) == 0 {
					continue
				}

				fmt.Println("for automation : ", slice.Title)

				for _, dependency := range command.Dependencies {
					if dependency.ElementType == "EVENT" && dependency.Type == "INBOUND" {
						fmt.Println(slice.Title, dependency)
					}
				}
			}

			//fmt.Println(eventIds)
			//_ = automationID

			for _, processor := range slice.Processors {
				if processor.ID == automationID {
					fmt.Println(processor)
				}
			}

			//if processor.Type == "AUTOMATION" {
			//
			//	fmt.Println(processor)
			//}
		}

		//fmt.Println(slice.Processors)
	}
}

func Test_Translators(t *testing.T) {

	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		t.Fatal(err)
	}

	for _, slice := range config.Slices {

		for _, command := range slice.Commands {

			//if command.Slice != "slice: Archive Item" {
			//	continue
			//}

			//var automation *Processor

			var automationID string

			var eventIds []string
			//var event *Event
			//var cmd *Command
			//var processor *Processor

			for _, dependency := range command.Dependencies {

				if dependency.ElementType == "AUTOMATION" {
					automationID = dependency.ID
					//fmt.Println(slice.Title, dependency)
				}

				//fmt.Println(dependency)
			}

			if automationID != "" {

				//fmt.Println(slice.Title)

				for _, slices := range config.Slices {
					for _, readmodel := range slices.Readmodels {
						if readmodel.ListElement == true {
							// list elements are being provided on which items we should iterate over.
							continue
						}
						var valid bool
						for _, dependency := range readmodel.Dependencies {
							if dependency.ID == automationID && dependency.Type == "OUTBOUND" {
								valid = true
							}
						}

						if valid {
							for _, dependency := range readmodel.Dependencies {
								if dependency.Type == "INBOUND" && dependency.ElementType == "EVENT" {
									//fmt.Println("-- EVENT :" + dependency.Title)
									eventIds = append(eventIds, dependency.ID)
								}
							}

						}
					}
				}

				if len(eventIds) != 0 {
					continue
				}

				fmt.Println("for translation : ", slice.Title)

				for _, dependency := range command.Dependencies {
					if dependency.ElementType == "EVENT" && dependency.Type == "INBOUND" {
						fmt.Println(slice.Title, dependency)
					}
				}
			}

			//fmt.Println(eventIds)
			//_ = automationID

			for _, processor := range slice.Processors {
				if processor.ID == automationID {
					fmt.Println(processor)
				}
			}

			//if processor.Type == "AUTOMATION" {
			//
			//	fmt.Println(processor)
			//}
		}

		//fmt.Println(slice.Processors)
	}
}

func Test_ListReadModels(t *testing.T) {

	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		t.Fatal(err)
	}

	for _, slice := range config.Slices {

		for _, readModel := range slice.Readmodels {

			fmt.Println("=== " + readModel.Title)
			//if readModel.Slice != "slice: Inventories" {
			//	continue
			//}

			//var automation *Processor

			//var automationID string

			var eventIds []string

			consumedEvents := make(map[string]struct{}, 0)

			for _, dependency := range readModel.Dependencies {
				if dependency.ElementType == "EVENT" && dependency.Type == "INBOUND" {
					consumedEvents[dependency.ID] = struct{}{}
					eventIds = append(eventIds, dependency.ID)
				}
			}

			if len(eventIds) > 0 {

				for _, slices := range config.Slices {
					for _, event := range slices.Events {

						if _, ok := consumedEvents[event.ID]; !ok {
							continue
						}

						fmt.Println(event.Title)

					}
				}

				for _, dependency := range readModel.Dependencies {
					if dependency.ElementType == "AUTOMATION" && dependency.Type == "OUTBOUND" {
						fmt.Println("AUTO --", dependency)
					}
				}
			}

			//fmt.Println(eventIds)
			//_ = automationID

			//for _, processor := range slice.Processors {
			//	if processor.ID == automationID {
			//		fmt.Println(slice.Title, processor)
			//	}
			//}

			//if processor.Type == "AUTOMATION" {
			//
			//	fmt.Println(processor)
			//}
		}

		//fmt.Println(slice.Processors)
	}
}

func Test_ThingsToRenderForASlice(t *testing.T) {

	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		t.Fatal(err)
	}

	for _, slice := range config.Slices {

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
