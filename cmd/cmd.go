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
		startURL := "https://monzo.com/"
		client := &http.Client{ // todo: config
			Timeout: time.Second * 30, // Increase the timeout to give slow servers a chance to respond
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse // stop after 10 redirects
				}
				return nil
			},
		}

		queue := make(chan string, 100)
		g, gctx := errgroup.WithContext(ctx)
		c := crawler.New(parser.New(client), g, queue)

		c.Start(gctx)
		if err = c.Crawl(gctx, startURL); err != nil {
			return fmt.Errorf("failed to crawl web page: %w", err)
		}

		running := true
		for running {
			<-time.Tick(1 * time.Second) // todo: config
			if len(queue)+c.ActiveWorkersCount() == 0 {
				fmt.Println("There are no active workers, nor any pending tasks.")
				close(queue)
				running = false
			}
		}

		if gErr := c.Wait(); gErr != nil {
			return gErr
		}

		s := sitemap.New("sitemap", "https://www.sitemaps.org/schemas/sitemap/0.9/")
		s.GenerateSitemap(c.SeenURLs(), startURL)
		if wErr := s.WriteToFile("sitemap.xml"); wErr != nil { // todo: config
			return fmt.Errorf("failed to write sitemap: %w", err)
		}
		return nil
	}
	return cmd
}

func Execute() {
	if err := Cmd().Execute(); err != nil {
		os.Exit(1)
	}
}
