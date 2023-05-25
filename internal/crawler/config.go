package crawler

import (
	"github.com/spf13/pflag"

	"goscout/flags"
)

type Config struct {
	WorkerCount int
	QueueSize   int
}

func (c *Config) Flags(prefix string) *pflag.FlagSet {
	const name = "CrawlerConfig"
	f := pflag.NewFlagSet(name, pflag.PanicOnError)

	f.IntVar(&c.WorkerCount, "worker_count", 100, "number of worker count for crawler")
	f.IntVar(&c.QueueSize, "queue_size", 100, "maximum number of tasks that queue can store")

	return flags.MapWithPrefix(f, name, pflag.PanicOnError, prefix)
}
