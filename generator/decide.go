package generator

import (
	"context"
	"fmt"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePackagePath(outPath string) (string, error) {

	slog.Info("Try to resolve path for", outPath, "package...")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	slog.Info("GOPATH:", gopath)

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}
	slog.Info("Resolving path:", absOutPath)

	for _, path := range strings.Split(gopath, ":") {
		gopathSrc := filepath.Join(path, "src")
		if strings.HasPrefix(absOutPath, gopathSrc) {
			return absOutPath[len(gopathSrc)+1:], nil
		}
	}
	return "", fmt.Errorf("path(%s) not in GOPATH(%s)", absOutPath, gopath)
}

func ListTemplatesForGen(ctx context.Context, model *eventmodel.EventModel, slice *eventmodel.Slice, absOutPath string) (units []*GenerationUnit, err error) {

	info := &template.GenerationInfo{
		Model:          model,
		OutputFilePath: absOutPath,
		Slice:          slice,
	}

	for _, command := range slice.Commands {
		{
			t := template.NewCommandTemplate(info, &command)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)
		}

		{
			t := template.NewCommandHandlerTemplate(info, &command)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)
		}
	}

	for _, event := range slice.Events {

		t := template.NewEventTemplate(info, &event)

		unit, err := NewGenUnit(ctx, t, absOutPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", absOutPath, err)
		}
		units = append(units, unit)

		{
			t := template.NewEventHandlerTemplate(info, &event)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)
		}
	}

	for _, aggregateName := range slice.Aggregates {
		{
			t := template.NewAggregateTemplate(info, aggregateName)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)
		}
		{
			t := template.NewAggregateHandlerTemplate(info, aggregateName)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)
		}
	}

	for _, readModel := range slice.Readmodels {
		{

			t := template.NewReadModelTemplate(info, &readModel)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)

		}

		{

			t := template.NewProjectorTemplate(info, &readModel)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)

		}
		{

			t := template.NewQueryHandlerTemplate(info, &readModel)

			unit, err := NewGenUnit(ctx, t, absOutPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", absOutPath, err)
			}
			units = append(units, unit)

		}

	}

	{
		t := template.NewOpenApiTemplate(info)

		unit, err := NewGenUnit(ctx, t, absOutPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", absOutPath, err)
		}
		units = append(units, unit)
	}

	//if len(slice.Commands) > 0 {
	//	// this is a state change
	//	{
	//		t := template.NewStateChangeTemplate(info)
	//
	//		unit, err := NewGenUnit(ctx, t, absOutPath)
	//		if err != nil {
	//			return nil, fmt.Errorf("%s: %v", absOutPath, err)
	//		}
	//		units = append(units, unit)
	//	}
	//}

	//lg.Logger.Logln(3, "\nGeneration Info:", info.String())
	///*stubSvc, err := NewGenUnit(ctx, template.NewStubInterfaceTemplate(info), absOutPath)
	//if err != nil {
	//	return nil, err
	//}
	//units = append(units, stubSvc)*/
	//
	//genTags := mstrings.FetchTags(iface.Docs, TagMark+MicrogenMainTag)
	//lg.Logger.Logln(2, "Tags:", strings.Join(genTags, ", "))
	//uniqueTemplate := make(map[string]template.Template)
	//for _, tag := range genTags {
	//	templates := tagToTemplate(tag, info)
	//	if templates == nil {
	//		lg.Logger.Logln(1, "Warning: Unexpected tag", tag)
	//		continue
	//	}
	//	for _, t := range templates {
	//		uniqueTemplate[t.DefaultPath()] = t
	//	}
	//}
	//for _, t := range uniqueTemplate {
	//	unit, err := NewGenUnit(ctx, t, absOutPath)
	//	if err != nil {
	//		return nil, fmt.Errorf("%s: %v", absOutPath, err)
	//	}
	//	units = append(units, unit)
	//}

	return units, nil
}
