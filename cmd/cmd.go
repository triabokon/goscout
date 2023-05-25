package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"goscout/internal/crawler"
	"goscout/internal/parser"
	"goscout/internal/sitemap"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "goscout",
		Aliases: []string{"gs"},
		Short:   "GoScout is a simple web-crawler tool.",
	}

	var config Config
	cmd.Flags().AddFlagSet(config.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		ctx := context.Background()
		client := &http.Client{Timeout: config.HTTPTimeout}
		g, gctx := errgroup.WithContext(ctx)
		c := crawler.New(config.Crawler, parser.New(client))

		fmt.Printf("Start crawler with %d workers\n", config.Crawler.WorkerCount)
		c.Start(gctx, g)

		fmt.Printf("Crawling website %s\n", config.SiteURL)
		started := time.Now()
		if err = c.Crawl(gctx, config.SiteURL, g); err != nil {
			return fmt.Errorf("failed to crawl web page: %w", err)
		}
		crawling := true
		var elapsedTime time.Duration
		for crawling {
			<-time.Tick(config.CheckInterval)
			fmt.Print(".")
			if !c.HasWorkToDo() {
				c.Stop()
				crawling = false
				elapsedTime = time.Since(started)
				fmt.Print("\n")
			}
		}
		if gErr := g.Wait(); gErr != nil {
			return gErr
		}

		seenURLs := c.SeenURLs()
		fmt.Printf("Crawler foung %d urls in %s time\n", len(seenURLs), elapsedTime)

		fmt.Println("Generating sitemap ...")
		s := sitemap.New(config.Sitemap)
		s.GenerateSitemap(seenURLs, config.SiteURL)

		fmt.Printf("Writing sitemap to %s ...\n", config.FileName)
		if wErr := s.WriteToFile(config.FileName); wErr != nil {
			return fmt.Errorf("failed to write sitemap: %w", err)
		}
		fmt.Println("Sitemap successfully written!")
		return nil
	}
	return cmd
}

func Execute() {
	if err := Cmd().Execute(); err != nil {
		fmt.Println(fmt.Errorf("encountered error: %w", err))
		os.Exit(1)
	}
}
