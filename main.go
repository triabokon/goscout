package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	// todo: logging and cobra and check schema to be https
	ctx := context.Background()
	startUrl := "https://monzo.com/"
	client := &http.Client{ // todo: config
		Timeout: time.Second * 30, // Increase the timeout to give slow servers a chance to respond
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse // stop after 10 redirects
			}
			return nil
		},
	}
	parser := NewParser(client)
	c := NewCrawler(ctx, parser)
	c.g.Go(func() error {
		return c.Crawl(ctx, startUrl)
	})
	if err := c.g.Wait(); err != nil {
		fmt.Println("Encountered error:", err)
	}

	s := NewSitemap("sitemap", "https://www.sitemaps.org/schemas/sitemap/0.9/")
	s.SetSitemap(syncMapToMap(c.visited), startUrl)
	err := s.WriteToFile("output/sitemap.xml", 2) // todo: make it config
	if err != nil {
		return
	}
}

func syncMapToMap(syncMap *sync.Map) map[string][]string {
	result := make(map[string][]string)

	syncMap.Range(func(key, value interface{}) bool {
		if strKey, ok := key.(string); ok {
			if strSlice, ok := value.([]string); ok {
				result[strKey] = strSlice
			}
		}
		return true
	})

	return result
}
