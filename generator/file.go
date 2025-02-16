package generator

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"os"
	"path/filepath"
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

func ResolvePackagePath(outPath string) (string, error) {
	//fmt.Println("Try to resolve path for", outPath, "package...")

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	//fmt.Println("GOPATH:", gopath)

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}
	//fmt.Println("Resolving path:", absOutPath)

	for _, path := range strings.Split(gopath, ":") {
		gopathSrc := filepath.Join(path, "src")
		if strings.HasPrefix(absOutPath, gopathSrc) {
			return absOutPath[len(gopathSrc)+1:], nil
		}
	}
	return "", fmt.Errorf("path(%s) not in GOPATH(%s)", absOutPath, gopath)
}
