package main

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"path"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Crawler struct {
	g      *errgroup.Group
	parser *Parser

	visited *sync.Map
}

func NewCrawler(ctx context.Context, p *Parser) *Crawler {
	g, ctx := errgroup.WithContext(ctx)
	return &Crawler{g: g, parser: p, visited: &sync.Map{}}
}

func (c *Crawler) Crawl(ctx context.Context, url string) error {
	if _, ok := c.visited.Load(url); ok {
		return nil
	}
	fmt.Println("Visit page", url)
	c.visited.Store(url, []string{})
	outLinks, staticLinks, err := c.parser.ExtractAllLinks(url)
	if err != nil {
		return err
	}
	outLinks = c.filterLinks(outLinks)
	staticLinks = unique(staticLinks)
	c.visited.Store(url, append(outLinks, staticLinks...))

	for _, l := range outLinks {
		if checkMimeType(l) {
			// goroutines have a reference to the variable and not its value at the time of the goroutine's creation.
			// If you don't create a new copy for each goroutine, they could all end up seeing the last value of the loop variable, due to the goroutines being scheduled and executed after the loop has finished iterating.
			link := l
			c.g.Go(func() error {
				return c.Crawl(ctx, link)
			})
		}
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue if not cancelled
		}
	}
	return nil
}

func (c *Crawler) filterLinks(links []string) []string {
	filteredLinks := make([]string, 0, len(links))
	for _, l := range unique(links) {
		if _, ok := c.visited.Load(l); !ok {
			filteredLinks = append(filteredLinks, l)
		}
	}
	return filteredLinks
}

func unique(s []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getExtension(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return path.Ext(parsedURL.Path), nil
}

func checkMimeType(url string) bool {
	extension, _ := getExtension(url)
	if extension == "" {
		return true
	}
	mimeType := mime.TypeByExtension(extension)
	return strings.HasPrefix(mimeType, "text/")
}
