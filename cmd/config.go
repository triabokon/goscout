package cmd

import (
	"time"

	"github.com/spf13/pflag"

	"goscout/internal/crawler"
	"goscout/internal/sitemap"
)

type Config struct {
	SiteURL       string
	FileName      string
	CheckInterval time.Duration
	HTTPTimeout   time.Duration

	Crawler crawler.Config
	Sitemap sitemap.Config
}

func (c *Config) Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("GoScoutConfig", pflag.PanicOnError)

	f.StringVar(&c.SiteURL, "site_url", "https://monzo.com/", "url of the site to crawl")
	f.StringVar(&c.FileName, "file_name", "sitemap.xml", "filename where to write sitemap")
	f.DurationVar(
		&c.CheckInterval, "check_interval",
		time.Second, "time interval to check if there are any pages left to crawl",
	)
	f.DurationVar(
		&c.HTTPTimeout, "http_timeout",
		10*time.Second, "time limit for requests for http client",
	)

	f.AddFlagSet(c.Crawler.Flags("crawler"))
	f.AddFlagSet(c.Sitemap.Flags("sitemap"))
	return f
}
