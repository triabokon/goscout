package cmd

import (
	"github.com/spf13/pflag"

	"goscout/internal/sitemap"
)

type Config struct {
	Sitemap sitemap.Config
}

func (c *Config) Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("GoScoutConfig", pflag.PanicOnError)

	f.AddFlagSet(c.Sitemap.Flags("sitemap"))

	return f
}
