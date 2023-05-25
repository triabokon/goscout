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
	g                  *errgroup.Group
	seenURLs           *sync.Map
	parser             *parser.Parser
	activeWorkersCount int64
	queue              chan string
	workerCount        int
}

func New(p *parser.Parser, g *errgroup.Group, queue chan string) *Crawler {
	return &Crawler{g: g, parser: p, seenURLs: &sync.Map{}, queue: queue, workerCount: 10}
}

func (c *Crawler) Crawl(ctx context.Context, url string) error {
	if _, ok := c.seenURLs.Load(url); ok {
		return nil
	}
	c.seenURLs.Store(url, nil)
	fmt.Println("Visit page", url)
	navigationURLs, staticURLs, err := c.parser.ExtractURLs(url)
	if err != nil {
		return fmt.Errorf("failed to extract link from web page: %w", err)
	}
	filteredStaticLinks, err := filterStaticLinks(staticURLs)
	if err != nil {
		return fmt.Errorf("failed to filter static links: %w", err)
	}
	c.seenURLs.Store(url, append(c.filterNavigationURLs(navigationURLs), filteredStaticLinks...))

	for _, link := range navigationURLs {
		if _, ok := c.seenURLs.Load(link); ok {
			continue
		}
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue if not cancelled
		}
		textLink, mErr := isTextLink(link)
		if mErr != nil {
			return fmt.Errorf("failed to check mime type: %w", mErr)
		}
		if textLink {
			// send this job to the worker pool
			select {
			case c.queue <- link:
			default:
				// Send to jobs was not ready, do the job
				// in the current worker.
				if cErr := c.Crawl(ctx, link); cErr != nil {
					return fmt.Errorf("failed to crawl web page: %w", cErr)
				}
			}
		}
	}
	return nil
}

func (c *Crawler) Start(ctx context.Context) {
	for w := 0; w < c.workerCount; w++ {
		c.g.Go(func() error {
			if err := c.worker(ctx); err != nil {
				return fmt.Errorf("worker failed: %w", err)
			}
			return nil
		})
	}
}

func (c *Crawler) Wait() error {
	return c.g.Wait()
}

func (c *Crawler) SeenURLs() map[string][]string {
	return syncMapToMap(c.seenURLs)
}

func (c *Crawler) ActiveWorkersCount() int {
	return int(c.activeWorkersCount)
}

func (c *Crawler) worker(ctx context.Context) error {
	for l := range c.queue {
		atomic.AddInt64(&c.activeWorkersCount, 1)
		if err := c.Crawl(ctx, l); err != nil {
			return fmt.Errorf("failed to crawl web page: %w", err)
		}
		atomic.AddInt64(&c.activeWorkersCount, -1)
	}
	return nil
}

func (c *Crawler) filterNavigationURLs(urls []string) []string {
	filtered := make([]string, 0, len(urls))
	for _, l := range unique(urls) {
		if _, ok := c.seenURLs.Load(l); !ok {
			filtered = append(filtered, l)
		}
	}
	return filtered
}
