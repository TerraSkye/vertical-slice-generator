package template

import (
	"context"
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/gosimple/slug"
	"github.com/terraskye/vertical-slice-generator/eventmodel"
	"github.com/terraskye/vertical-slice-generator/generator/write_strategy"
	"github.com/terraskye/vertical-slice-generator/template"
	"strconv"
	"strings"
)

type commandResourceTemplate struct {
	info    *GenerationInfo
	command *eventmodel.Command
}

func NewCommandResourceTemplate(info *GenerationInfo, command *eventmodel.Command) Template {
	return &commandResourceTemplate{
		info:    info,
		command: command,
	}
}

func (t *commandResourceTemplate) Render(ctx context.Context) write_strategy.Renderer {
	commandPackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain/commands")
	if err != nil {
		panic(err)
	}
	aggregatePackage, err := ResolvePackagePath(t.info.OutputFilePath + "/domain")
	if err != nil {
		panic(err)
	}

	z := NewFile(eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))
	z.ImportAlias(PackageEventSourcing, "cqrs")
	z.ImportName(commandPackage, "")
	z.ImportName(aggregatePackage, "")

	z.ImportAlias("github.com/go-kit/kit/transport/http", "kithttp")
	z.ImportAlias("github.com/afosto/go-json", "json")
	z.ImportAlias("github.com/afosto/utils-go/http/request", "afreq")

	idFields := eventmodel.Fields(t.command.Fields).IDAttributes()

	if len(idFields) == 1 {
		// then the URL contains the id field

		switch idFields[0].Type {
		case "String":
			//idFields[0].Name
		case "UUID":
		case "Boolean":
		case "Double":
		case "Decimal":
		case "Date":
		case "DateTime":
		case "Long":
		case "Int":
		case "Custom":

		}

		if idFields[0].Type == "UUID" {
			// placeholder
		}

	}

	// HTTP endpoint
	z.Func().Id("MakeHttpHandler").Params(
		Id("r").Op("*").Qual("github.com/gorilla/mux", "Router"),
		Id("store").Qual(PackageEventSourcing, "EventStore"),
	).Params(Qual("net/http", "Handler")).BlockFunc(func(group *Group) {
		//TODO paramter filter for type.
		group.Id("r").Dot("Methods").Call(Lit("POST")).Dot("Path").Call(Id(fmt.Sprintf("\"debug/%s/{id}/%s\"", strings.ToLower(t.info.Model.CodeGen.Domain), slug.Make(strings.ReplaceAll(t.info.Slice.Title, "slice:", ""))))).Dot("Handler").Call(Line().Qual("github.com/go-kit/kit/transport/http", "NewServer").Params(
			Func().Params(Id("ctx").Qual("context", "Context"), Id("request").Any()).Params(Any(), Error()).BlockFunc(func(group *Group) {

				group.Id("handler").Op(":=").Id("NewCommandHandler").Call(Id("store"))
				group.Line().List(Id("result"), Err()).Op(":=").Id("handler").Call(Id("ctx"), Id("request").Assert(Id("Command")))

				// endpoint body
				group.Line().If(Err().Op("!=").Nil()).Block(Return(Nil(), Err()))

				group.Line().Return(Id("result"), Nil())
			}),
			Id(fmt.Sprintf("decode%sRequest", eventmodel.ProcessTitle(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))),
			Id(fmt.Sprintf("encode%sRequest", eventmodel.ProcessTitle(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))),
		),
		)
		group.Return(Id("r"))
	}).Line()

	z.Func().Id(fmt.Sprintf("decode%sRequest", eventmodel.ProcessTitle(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))).Params(
		Id("ctx").Qual("context", "Context"),
		Id("r").Op("*").Qual("net/http", "Request"),
	).Params(
		Interface(),
		Error(),
	).BlockFunc(func(group *Group) {

		if idAttributes := eventmodel.Fields(t.command.Fields).IDAttributes(); len(idAttributes) > 0 {
			group.Line().Id(idAttributes[0].Name).Op(":=").Qual("github.com/google/uuid", "MustParse").Call(Qual("github.com/gorilla/mux", "Vars").Call(Id("r")).Types(Id(strconv.Quote("id"))))
		}

		//TODO improve field mapping
		group.Line().Var().Id("payload").Struct(Id("Data").Add(
			template.FieldsStruct(eventmodel.Fields(t.command.Fields).DataAttributes(), true),
		).Tag(map[string]string{"json": "data"})). //Id("Data").Struct(Id("ID").String().Tag(
			//	map[string]string{"json": "id", "validate": "omitempty,uuid4"}),
			//).Tag(map[string]string{"json": "data"}
			//),

			Line().
			Line().
			IfFunc(func(c *Group) {
				c.Id("err").Op(":=").
					Qual("github.com/afosto/go-json", "NewDecoder").Params(
					Id("r").Dot("Body")).Dot("UseAutoTrimSpace").Call().Dot("Decode").Call(Op("&").Id("payload")).Op(";").Err().Op("!=").Nil()
			}).BlockFunc(func(group *Group) {
			group.Return(Nil(), Qual("github.com/pkg/errors", "WithStack").Params(Err()))
			//group.Return(Nil(),Qual("github.com/afosto/utils-go/http/response", "NewError").Call(Qual("tracking/errors","ErrFailedToParsePayload"),Id("err")))
		}).
			Line().
			Line().
			IfFunc(func(c *Group) {
				c.Id("err").Op(":=").
					Qual("github.com/afosto/utils-go/intl18", "Validator").Dot("StructCtx").Call(Id("ctx"), Id("payload")).Op(";").Err().Op("!=").Nil()
			}).BlockFunc(func(group *Group) {
			group.Return(Nil(), Qual("github.com/pkg/errors", "WithStack").Params(Err()))
			//group.Return(Nil(),Qual("github.com/afosto/utils-go/http/response", "NewError").Call(Qual("tracking/errors","ErrFailedToParsePayload"),Id("err")))
		}).Line()

		group.Var().Id("cmd").Op("=").Id("Command").Block(DictFunc(func(dict Dict) {

			eventmodel.SortFields(t.command.Fields)
			for _, field := range eventmodel.Fields(t.command.Fields).IDAttributes() {
				var statement *Statement
				statement = Id(field.Name)

				if len(t.command.Fields) == 1 {
					dict[Id(eventmodel.ProcessTitle(field.Name))] = statement.Op(",")
				} else {
					dict[Id(eventmodel.ProcessTitle(field.Name))] = statement

				}
			}

			for _, field := range eventmodel.Fields(t.command.Fields).DataAttributes() {

				var statement *Statement
				statement = Id("payload").Dot("Data").Dot(eventmodel.ProcessTitle(field.Name))

				if len(t.command.Fields) == 1 {
					dict[Id(eventmodel.ProcessTitle(field.Name))] = statement.Op(",")
				} else {
					dict[Id(eventmodel.ProcessTitle(field.Name))] = statement

				}
			}

		}))

		//TODO map struct

		group.Return(Id("cmd"), Nil())
	})
	//).Block(Return(Nil(), Nil()))
	z.Func().Id(fmt.Sprintf("encode%sRequest", eventmodel.ProcessTitle(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")))).Params(
		Id("ctx").Qual("context", "Context"),
		Id("w").Qual("net/http", "ResponseWriter"),
		Id("response").Interface(),
	).Params(Error()).Block(
		Return(Nil()))

	return z
}

func (t *commandResourceTemplate) DefaultPath() string {
	return eventmodel.SnakeCase(strings.ReplaceAll(t.info.Slice.Title, "slice:", "")) + "/http.go"
}

func (t *commandResourceTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *commandResourceTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}
