package sitemap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/triabokon/goscout/internal/sitemap"
)

func TestSitemap_GenerateSitemap(t *testing.T) {
	data := map[string][]string{
		"https://example.com":        {"https://example.com/child1", "https://example.com/child2"},
		"https://example.com/child1": {"https://example.com/grandchild1"},
	}
	rootValue := "https://example.com"
	expectedSitemap := &sitemap.URL{
		Loc: "https://example.com",
		URLs: []*sitemap.URL{
			{
				Loc: "https://example.com/child1",
				URLs: []*sitemap.URL{
					{
						Loc: "https://example.com/grandchild1",
					},
				},
			},
			{
				Loc: "https://example.com/child2",
			},
		},
	}

	s := sitemap.New(sitemap.Config{})
	s.GenerateSitemap(data, rootValue)
	assert.Equal(t, expectedSitemap, s.Index().URL)
}
