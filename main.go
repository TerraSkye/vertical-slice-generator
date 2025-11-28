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
	slicesToGenerate := Checkboxes(
		"Which slice would you like to generate", append([]string{"all"}, tasks...),
	)

	//slicesToGenerate := []string{"all"}
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

	}

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
