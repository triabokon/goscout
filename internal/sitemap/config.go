package sitemap

import (
	"github.com/spf13/pflag"

	"goscout/flags"
)

type Config struct {
	Indent int
}

func (c *Config) Flags(prefix string) *pflag.FlagSet {
	const name = "SitemapConfig"
	f := pflag.NewFlagSet(name, pflag.PanicOnError)

	f.IntVar(&c.Indent, "indent", 1, "xml sitemap indent")

	return flags.MapWithPrefix(f, name, pflag.PanicOnError, prefix)
}
