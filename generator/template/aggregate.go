package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/generator/config"
	"strings"
)

type aggregateTemplate struct {
	info   *GenerationInfo
	slices []config.Slices
}

func (a aggregateTemplate) Prepare(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (a aggregateTemplate) DefaultPath() string {
	//TODO implement me
	panic("implement me")
}

func (a aggregateTemplate) Render(ctx context.Context) {
	path := strings.Split(a.info.OutputPackageImport, "/")
	f := NewFile(path[len(path)-1])
	f.ImportName("github.com/google/uuid", "uuid")

	// generate all aggregates
	//for _, aggregate := range config.Aggregates {
	//	aggregateFile := fmt.Sprintf("gen/%s/domain/%s.go", strings.ToLower(config.CodeGen.Domain), strings.ToLower(aggregate.Title))
	//	files[aggregateFile] = generator.GetFile("domain")
	//	files[aggregateFile].Type().Id(generator.ToCamelCase(aggregate.Title)).StructFunc(func(group *Group) {
	//		for _, field := range aggregate.Field {
	//			property := Id(generator.ToCamelCase(field.Name))
	//			if field.Cardinality != "Single" {
	//				property = property.Index()
	//			}
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
	//
	//	})
	//
	//	//log.Printf("Aggregate #%d: %s", i+1, slice)
	//}
	//
	//f.Type().Id(serviceAuditStructName).Struct(
	//	Id("Service"),
	//	Id("activity").Qual("github.com/afosto/utils-go/activity", "Service"),
	//)

	//TODO implement me
	panic("implement me")
}

func NewAggregateTemplate(info *GenerationInfo) Template {
	return &aggregateTemplate{
		info: info,
	}
}

func (a aggregateTemplate) handler() {
	//path := strings.Split(a.info.OutputPackageImport, "/")
	//f := NewFile(path[len(path)-1])

}
