package crawler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/triabokon/goscout/internal/crawler"
	"github.com/triabokon/goscout/internal/crawler/mocks"
)

var gfi = gofakeit.New(1)

func TestCrawler_Crawl(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		startURL := gfi.URL()
		for name, tc := range map[string]struct {
			tuneMock func(p *mocks.MockParser)
			errorMsg string
		}{
			"url extraction error": {
				errorMsg: "failed to extract url from web page",
				tuneMock: func(p *mocks.MockParser) {
					p.EXPECT().ExtractURLs(startURL).Return(nil, nil, fmt.Errorf("url extraction error"))
				},
			},
			"web url filtering error": {
				errorMsg: "failed to filter web urls",
				tuneMock: func(p *mocks.MockParser) {
					p.EXPECT().ExtractURLs(startURL).Return([]string{":"}, nil, nil)
				},
			},
			"static url filtering error": {
				errorMsg: "failed to filter static urls",
				tuneMock: func(p *mocks.MockParser) {
					p.EXPECT().ExtractURLs(startURL).Return(nil, []string{":"}, nil)
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				ctx := context.Background()
				parser := mocks.NewMockParser(ctrl)

				tc.tuneMock(parser)

				c := crawler.New(crawler.Config{Depth: 3}, parser)
				err := c.Crawl(ctx, startURL, 1)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			})
		}
	})

	t.Run("exceeds depth", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		parser := mocks.NewMockParser(ctrl)

		c := crawler.New(crawler.Config{Depth: 1}, parser)
		err := c.Crawl(ctx, gfi.URL(), 2)
		assert.Error(t, err)
		assert.Equal(t, crawler.ErrExceedsDepth, err)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		parser := mocks.NewMockParser(ctrl)

		startURL := gfi.URL()
		staticUrl := "https://example.com/image.jpeg"

		parser.EXPECT().ExtractURLs(startURL).Return([]string{}, []string{staticUrl}, nil)

		c := crawler.New(crawler.Config{Depth: 3}, parser)
		err := c.Crawl(ctx, startURL, 1)
		assert.NoError(t, err)
		assert.Equal(t, map[string][]string{startURL: {staticUrl}}, c.SeenURLs())
	})

	t.Run("url already seen", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		parser := mocks.NewMockParser(ctrl)

		startURL := gfi.URL()
		parser.EXPECT().ExtractURLs(startURL).Return([]string{}, []string{}, nil).Times(1)

		c := crawler.New(crawler.Config{Depth: 3}, parser)
		err := c.Crawl(ctx, startURL, 2)
		assert.NoError(t, err)

		err = c.Crawl(ctx, startURL, 2)
		assert.NoError(t, err)
	})
}
