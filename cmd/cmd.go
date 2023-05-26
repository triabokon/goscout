package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"goscout/internal/crawler"
	"goscout/internal/parser"
	"goscout/internal/sitemap"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "goscout",
		Aliases:      []string{"gs"},
		Short:        "GoScout is a simple web-crawler tool.",
		SilenceUsage: true,
	}

	var config Config
	cmd.Flags().AddFlagSet(config.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		ctx := context.Background()
		client := &http.Client{Timeout: config.HTTPTimeout}
		c := crawler.New(config.Crawler, parser.New(client))

		fmt.Printf("Start crawler with %d workers\n", config.Crawler.WorkerCount)
		c.Start(ctx)

		fmt.Printf("Crawling website %s\n", config.SiteURL)
		started := time.Now()
		if err = c.Crawl(ctx, config.SiteURL, 0); err != nil {
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
		c.Wait()

		// log errors, because we need to write urls that we managed to find
		if len(c.Errors()) != 0 {
			fmt.Println("Following errors occurred during website crawling: ")
			for _, e := range c.Errors() {
				fmt.Println(e)
			}
			fmt.Println()
		}

		seenURLs := c.SeenURLs()
		fmt.Printf("Crawler found %d urls in %s time\n", len(seenURLs), elapsedTime)

		fmt.Println("Generating sitemap ...")
		s := sitemap.New(config.Sitemap)
		s.GenerateSitemap(seenURLs, config.SiteURL)

		fmt.Printf("Writing sitemap to %s ...\n", config.FileName)
		if wErr := s.WriteToFile(config.FileName); wErr != nil {
			return fmt.Errorf("failed to write sitemap: %w", wErr)
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
