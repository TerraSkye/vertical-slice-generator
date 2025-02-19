package template

import (
	"context"
	. "github.com/dave/jennifer/jen"
	"github.com/terraskye/vertical-slice-generator/generator/config"
	"strings"
)

type eventTemplate struct {
	info   *GenerationInfo
	slices []config.Slices
}

func (a eventTemplate) Prepare(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (a eventTemplate) DefaultPath() string {
	//TODO implement me
	panic("implement me")
}

func (a eventTemplate) Render(ctx context.Context) {
	path := strings.Split(a.info.OutputPackageImport, "/")

	f := NewFile(path[len(path)-1])

	f.ImportName("github.com/google/uuid", "uuid")

	//f.Type().Id(serviceAuditStructName).Struct(
	//	Id("Service"),
	//	Id("activity").Qual("github.com/afosto/utils-go/activity", "Service"),
	//)

	//TODO implement me
	panic("implement me")
}

//func NewAggregateTemplate(info *GenerationInfo) Template {
//	return &eventTemplate{
//		info: info,
//	}
//}
