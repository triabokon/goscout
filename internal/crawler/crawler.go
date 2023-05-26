package crawler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"goscout/internal/parser"
)

type Crawler struct {
	config        Config
	seenURLs      *sync.Map
	parser        *parser.Parser
	activeWorkers int64
	queue         chan string
}

func New(c Config, p *parser.Parser) *Crawler {
	return &Crawler{
		config:   c,
		parser:   p,
		seenURLs: &sync.Map{},
		queue:    make(chan string, c.QueueSize),
	}
}

// Crawl crawls url, extracting and filtering its urls, then add found urls to the queue.
func (c *Crawler) Crawl(ctx context.Context, url string, g *errgroup.Group) error {
	// check if url has already been visited
	if _, ok := c.seenURLs.Load(url); ok {
		return nil
	}
	// store url to the map of visited urls, so other workers would not process it
	c.seenURLs.Store(url, nil)
	// extract all urls from the given web page
	webURLs, staticURLs, err := c.parser.ExtractURLs(url)
	if err != nil {
		return fmt.Errorf("failed to extract u from web page: %w", err)
	}

	filteredWebURLs, err := filterWebURLs(webURLs, c.seenURLs)
	if err != nil {
		return fmt.Errorf("failed to filter web urls: %w", err)
	}
	filteredStaticURLs, err := filterStaticURLs(staticURLs)
	if err != nil {
		return fmt.Errorf("failed to filter static urls: %w", err)
	}
	// update value in the seenURLs with newly found urls
	c.seenURLs.Store(url, append(filteredWebURLs, filteredStaticURLs...))

	for _, u := range filteredWebURLs {
		select {
		// check for context cancellation
		case <-ctx.Done():
			return ctx.Err()
		// send this url to the queue
		case c.queue <- u:
		default:
			// if the queue is full, then crawl this url immediately
			if cErr := c.Crawl(ctx, u, g); cErr != nil {
				return fmt.Errorf("failed to crawl web page: %w", cErr)
			}
		}
	}
	return nil
}

// Start initializes multiple workers based on the WorkerCount from the Config.
func (c *Crawler) Start(ctx context.Context, g *errgroup.Group) {
	for w := 0; w < c.config.WorkerCount; w++ {
		g.Go(func() error {
			if err := c.worker(ctx, g); err != nil {
				return fmt.Errorf("worker failed: %w", err)
			}
			return nil
		})
	}
}

func (c *Crawler) SeenURLs() map[string][]string {
	return seenURLsToMap(c.seenURLs)
}

// HasWorkToDo determines if there are any active workers or pending urls in the queue.
// We could close the queue and stop Crawler if not.
func (c *Crawler) HasWorkToDo() bool {
	return int(c.activeWorkers)+len(c.queue) > 0
}

func (c *Crawler) Stop() {
	close(c.queue)
}

// worker is process urls from the queue by calling the Crawl method.
func (c *Crawler) worker(ctx context.Context, g *errgroup.Group) error {
	for l := range c.queue {
		// increment the count of active workers
		atomic.AddInt64(&c.activeWorkers, 1)
		if err := c.Crawl(ctx, l, g); err != nil {
			return fmt.Errorf("failed to crawl web page: %w", err)
		}
		// once the url has been crawled, decrement the count of active workers
		atomic.AddInt64(&c.activeWorkers, -1)
	}
	return nil
}
