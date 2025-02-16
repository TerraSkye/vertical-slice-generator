package generator

import (
	. "github.com/dave/jennifer/jen"
	"strings"
)

func GetFile(tag string) *File {
	f := NewFile(strings.ToLower(ToCamel(tag)))

	f.ImportName("github.com/go-kit/kit/endpoint", "endpoint")
	f.ImportName("net/http", "http")
	f.ImportName("github.com/google/uuid", "uuid")
	f.ImportAlias("github.com/go-kit/kit/transport/http", "kithttp")
	f.ImportName("github.com/go-kit/kit/endpoint", "endpoint")
	f.ImportName("github.com/gorilla/mux", "mux")

	return f
}
