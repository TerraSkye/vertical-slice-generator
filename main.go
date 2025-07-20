package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator"
	"github.com/terraskye/vertical-slice-generator/generator/template"
	"github.com/vetcher/go-astra"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	flagFileName  = flag.String("file", "config.json", "Path to input file with interface.")
	flagOutputDir = flag.String("out", ".", "Output directory.")
	flagHelp      = flag.Bool("help", false, "Show help.")
	flagVerbose   = flag.Int("v", 1, "Sets vertical slice  eventmodel verbose level.")
	flagDebug     = flag.Bool("debug", false, "Print all vertical slice  eventmodel  messages. Equivalent to -v=100.")
)

func init() {
	flag.Parse()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//os.RemoveAll("/home/afosto/go/src/github.com/tskye/debug-gen/cart")
	//
	//lg.Logger.Level = *flagVerbose
	//if *flagDebug {
	//	lg.Logger.Level = 100
	//}
	//lg.Logger.Logln(1, "@microgen", Version)
	if *flagHelp || *flagFileName == "" {
		flag.Usage()
		os.Exit(0)
	}

	configuration, err := os.ReadFile(*flagFileName)

	if err != nil {
		logger.Error("failed to read file %v", err)
	}

	var cfg eventmodel.EventModel

	if err := json.Unmarshal(configuration, &cfg); err != nil {
		logger.Error("failed to decode configuration %v", err)
	}

	var tasks = make([]string, len(cfg.Slices))

	for i, slice := range cfg.Slices {
		tasks[i] = slice.Title
	}
	//slicesToGenerate := Checkboxes(
	//	"Which slice would you like to generate", append([]string{"all"}, tasks...),
	//)

	slicesToGenerate := []string{"slice: cart items"}
	if slices.Contains(slicesToGenerate, "all") {
		slicesToGenerate = tasks[1:]
	}

	absOutputDir, err := filepath.Abs(*flagOutputDir)
	if err != nil {
		logger.Error("failed to get outputpath %v", err)
		os.Exit(1)
	}

	_ = absOutputDir

	ctx, err := prepareContext(*flagFileName)
	if err != nil {
		logger.Error("failed preparing context %v", err)
		os.Exit(1)
	}

	for _, sliceName := range slicesToGenerate {

		slice := cfg.FindSlice(sliceName)

		units, err := generator.ListTemplatesForGen(ctx, &cfg, slice, absOutputDir+"/"+strings.ToLower(cfg.CodeGen.Domain))

		if err != nil {
			logger.Error("failed preparing context %v", err)
			os.Exit(1)
		}

		for _, unit := range units {
			if err := unit.Generate(ctx); err != nil {
				logger.Error("failed preparing context %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

		}
		//task := template.NewAggregateTemplate(&template.GenerationInfo{
		//	Model:          &cfg,
		//	OutputFilePath: absOutputDir + "/" + strings.ToLower(cfg.CodeGen.Domain),
		//	Slice:          slice,
		//})
		//
		//if err := task.Prepare(ctx); err != nil {
		//	logger.Error("failed to get outputpath %v", err)
		//	os.Exit(1)
		//}
		//
		//strategy, err := task.ChooseStrategy(ctx)
		//if err != nil {
		//	logger.Error("failed preparing strategy %v", err)
		//	os.Exit(1)
		//}
		//
		//if err := strategy.Write(task.Render(ctx)); err != nil {
		//	logger.Error("failed rendering %v", err)
		//	os.Exit(1)
		//}
		//
		//for _, command := range slice.Commands {
		//	_ = command
		//	task := template.NewCommandTemplate(&template.GenerationInfo{
		//		Model:          &cfg,
		//		OutputFilePath: absOutputDir + "/" + strings.ToLower(cfg.CodeGen.Domain),
		//		Slice:          slice,
		//	})
		//
		//	if err := task.Prepare(ctx); err != nil {
		//		logger.Error("failed to get outputpath %v", err)
		//		os.Exit(1)
		//	}
		//
		//	strategy, err := task.ChooseStrategy(ctx)
		//	if err != nil {
		//		logger.Error("failed preparing strategy %v", err)
		//		os.Exit(1)
		//	}
		//
		//	if err := strategy.Write(task.Render(ctx)); err != nil {
		//		fmt.Println(err)
		//		logger.Error("failed rendering %v", err)
		//		os.Exit(1)
		//	}
		//
		//}
		//
		//for _, event := range slice.Events {
		//	_ = event
		//	task := template.NewEventTemplate(&template.GenerationInfo{
		//		Model:          &cfg,
		//		OutputFilePath: absOutputDir + "/" + strings.ToLower(cfg.CodeGen.Domain),
		//		Slice:          slice,
		//	})
		//
		//	if err := task.Prepare(ctx); err != nil {
		//		logger.Error("failed to get outputpath %v", err)
		//		os.Exit(1)
		//	}
		//
		//	strategy, err := task.ChooseStrategy(ctx)
		//	if err != nil {
		//		logger.Error("failed preparing strategy %v", err)
		//		os.Exit(1)
		//	}
		//
		//	if err := strategy.Write(task.Render(ctx)); err != nil {
		//		fmt.Println(err)
		//		logger.Error("failed rendering %v", err)
		//		os.Exit(1)
		//	}
		//
		//}
		//
		////importPackagePath, err := template.resolvePackagePath(filepath.Dir(sourcePath))
		//if err != nil {
		//	return nil, err
		//}
		//absSourcePath, err := filepath.Abs(sourcePath)
		//if err != nil {
		//	return nil, err
		//}
		//outImportPath, err := resolvePackagePath(absOutPath)
		//if err != nil {
		//	return nil, err
		//}
		//
		//info := &template.GenerationInfo{
		//	SourcePackageImport: importPackagePath,
		//	SourceFilePath:      absSourcePath,
		//	OutputPackageImport: outImportPath,
		//	OutputFilePath:      absOutPath,
		//}
		//
		//if len(slice.Commands) > 0 {
		//
		//}
		//generator.ListTemplatesForGen
		//fmt.Sprintf("%s/%s/domain/%s.go", flagOutputDir, cfg.CodeGen.Domain, slice.Aggregate[0])

		//cfg.CodeGen.Domain

		//ctx, err := prepareContext(*flagFileName)
		//if err != nil {
		//	lg.Logger.Logln(0, "fatal:", err)
		//	os.Exit(1)
		//}
		//
		//template.NewAggregateTemplate()

	}

	//for _, unit := range units {
	//	err := unit.Generate(ctx)
	//	if err != nil && err != generator.EmptyStrategyError {
	//		lg.Logger.Logln(0, "fatal:", unit.Path(), err)
	//		os.Exit(1)
	//	}
	//}

	////lg.Logger.Logln(4, "Source file:", *flagFileName)
	//info, err := astra.ParseFile(*flagFileName)
	//if err != nil {
	//	//lg.Logger.Logln(0, "fatal:", err)
	//	os.Exit(1)
	//}
	//
	//i := findInterface(info)
	//if i == nil {
	//	//lg.Logger.Logln(0, "fatal: could not find interface with @microgen tag")
	//	//lg.Logger.Logln(4, "All founded interfaces:")
	//	//lg.Logger.Logln(4, listInterfaces(info.Interfaces))
	//	os.Exit(1)
	//}
	//
	//if err := generator.ValidateInterface(i); err != nil {
	//	//lg.Logger.Logln(0, "validation:", err)
	//	os.Exit(1)
	//}

	//ctx, err := prepareContext(*flagFileName, i)
	//if err != nil {
	//	lg.Logger.Logln(0, "fatal:", err)
	//	os.Exit(1)
	//}

	//absOutputDir, err := filepath.Abs(*flagOutputDir)
	//if err != nil {
	//	lg.Logger.Logln(0, "fatal:", err)
	//	os.Exit(1)
	//}
	//units, err := generator.ListTemplatesForGen(ctx, i, absOutputDir, *flagFileName, *flagGenProtofile, *flagGenMain)
	//if err != nil {
	//	lg.Logger.Logln(0, "fatal:", err)
	//	os.Exit(1)
	//}
	//for _, unit := range units {
	//	err := unit.Generate(ctx)
	//	if err != nil && err != generator.EmptyStrategyError {
	//		lg.Logger.Logln(0, "fatal:", unit.Path(), err)
	//		os.Exit(1)
	//	}
	//}
	//lg.Logger.Logln(1, "all files successfully generated")
}

func Checkboxes(label string, opts []string) []string {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	survey.AskOne(prompt, &res)

	return res
}

func prepareContext(filename string) (context.Context, error) {
	ctx := context.Background()
	p, err := astra.ResolvePackagePath(filename)
	if err != nil {
		return nil, err
	}
	ctx = template.WithSourcePackageImport(ctx, p)

	return ctx, nil
}
