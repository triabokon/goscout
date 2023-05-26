package sitemap

import (
	"github.com/spf13/pflag"

	"github.com/triabokon/goscout/flags"
)

type Config struct {
	XMLNS  string
	Indent int
}

func (c *Config) Flags(prefix string) *pflag.FlagSet {
	const name = "SitemapConfig"
	f := pflag.NewFlagSet(name, pflag.PanicOnError)

	f.StringVar(
		&c.XMLNS, "xml_ns",
		"https://www.sitemaps.org/schemas/sitemap/0.9/", "xml sitemap namespace name",
	)
	f.IntVar(&c.Indent, "indent", 1, "xml sitemap indent")

	return flags.MapWithPrefix(f, name, pflag.PanicOnError, prefix)
}
