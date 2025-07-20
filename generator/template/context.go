package template

import "context"

const (
	spi = "SourcePackageImport"
)

func WithSourcePackageImport(parent context.Context, val string) context.Context {
	return context.WithValue(parent, spi, val)
}

func SourcePackageImport(ctx context.Context) string {
	return ctx.Value(spi).(string)
}
