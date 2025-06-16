package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/generator"
	"github.com/terraskye/vertical-slice-generator/generator/config"
)

type GenerationInfo struct {
	OutputFilePath       string
	OutputPackageImport  string
	EventPackageImport   string
	CommandPackageImport string
}

func FieldsStruct(fields []config.Field) *Statement {
	return StructFunc(func(group *Group) {
		for _, field := range fields {

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

func Fields(fields []config.Field) []*Statement {
	var statements = make([]*Statement, 0)

	for _, field := range fields {
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
		statements = append(statements, property)
	}

	return statements
}
