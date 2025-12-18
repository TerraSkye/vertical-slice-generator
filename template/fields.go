package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
)

func FieldsStruct(fields []eventmodel.Field, withJsonTags bool) *Statement {
	return StructFunc(func(group *Group) {
		for _, field := range fields {
			property := Id(eventmodel.ProcessTitle(field.Name))
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
			case "Decimal":
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
				property = property.Add(FieldsStruct(field.SubFields, withJsonTags))
			default:
				property = property.Any()
			}

			if withJsonTags {

				property.Tag(map[string]string{"json": eventmodel.SnakeCase(eventmodel.ProcessTitle(field.Name))})
			}
			group.Add(property)
		}
	})

}
