package crawler

import (
	"github.com/spf13/pflag"

	"github.com/triabokon/goscout/flags"
)

const (
	MinWorkerCount = 10
	MinQueueSize   = 100
)

type Config struct {
	WorkerCount int
	QueueSize   int
	Depth       int
}

func (c *Config) Flags(prefix string) *pflag.FlagSet {
	const name = "CrawlerConfig"
	f := pflag.NewFlagSet(name, pflag.PanicOnError)

	f.IntVar(&c.WorkerCount, "worker_count", 100, "number of workers for crawler (min 10)")
	f.IntVar(
		&c.QueueSize, "queue_size",
		1000, "maximum number of tasks that queue can store (min 100)",
	)
	f.IntVar(&c.Depth, "depth", 100, "maximum depth the crawler would go")

	return flags.MapWithPrefix(f, name, pflag.PanicOnError, prefix)
}
