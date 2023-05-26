package crawler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"goscout/internal/parser"
)

type Crawler struct {
	config Config
	parser *parser.Parser

	wg            *sync.WaitGroup
	seenURLs      *sync.Map
	activeWorkers int64
	queue         chan Job
	errc          chan error
	errors        []error
}

type Job struct {
	URL   string
	Depth int
}

func New(c Config, p *parser.Parser) *Crawler {
	return &Crawler{
		config:   c,
		parser:   p,
		wg:       &sync.WaitGroup{},
		seenURLs: &sync.Map{},
		queue:    make(chan Job, c.QueueSize),
		errc:     make(chan error),
	}
}

// Crawl crawls url, extracting and filtering its urls, then add found urls to the queue.
func (c *Crawler) Crawl(ctx context.Context, url string, depth int) error {
	// check if url has already been visited
	if _, ok := c.seenURLs.Load(url); ok {
		return nil
	}
	// store url to the map of visited urls, so other workers would not process it
	c.seenURLs.Store(url, nil)
	if depth > c.config.Depth {
		return ErrExceedsDepth
	}
	// extract all urls from the given web page
	webURLs, staticURLs, err := c.parser.ExtractURLs(url)
	if err != nil {
		return fmt.Errorf("failed to extract url from web page: %w", err)
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
		case c.queue <- Job{URL: u, Depth: depth + 1}:
		default:
			// if the queue is full, then crawl this url immediately
			cErr := c.Crawl(ctx, u, depth+1)
			switch cErr {
			case nil:
			case ErrExceedsDepth:
				return nil
			default:
				return fmt.Errorf("failed to crawl web page: %w", cErr)
			}
		}
	}
	return nil
}

// Start initializes multiple workers based on the WorkerCount from the Config.
func (c *Crawler) Start(ctx context.Context) {
	for w := 0; w < c.config.WorkerCount; w++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.worker(ctx)
		}()
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for e := range c.errc {
			c.errors = append(c.errors, e)
		}
	}()
}

func (c *Crawler) SeenURLs() map[string][]string {
	return seenURLsToMap(c.seenURLs)
}

func (c *Crawler) Errors() []error {
	return c.errors
}

// HasWorkToDo determines if there are any active workers or pending urls in the queue.
// We could close the queue and stop Crawler if not.
func (c *Crawler) HasWorkToDo() bool {
	return int(c.activeWorkers)+len(c.queue) > 0
}

func (c *Crawler) Wait() {
	c.wg.Wait()
}

func (c *Crawler) Stop() {
	close(c.queue)
	close(c.errc)
}

// worker is process urls from the queue by calling the Crawl method.
func (c *Crawler) worker(ctx context.Context) {
	for j := range c.queue {
		// increment the count of active workers
		atomic.AddInt64(&c.activeWorkers, 1)
		err := c.Crawl(ctx, j.URL, j.Depth)
		// once the url has been crawled, decrement the count of active workers
		atomic.AddInt64(&c.activeWorkers, -1)
		switch err {
		case nil, ErrExceedsDepth:
			continue
		default:
			c.errc <- err
		}
	}
}
