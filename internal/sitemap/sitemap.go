package sitemap

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// todo: introduce xml ns https://www.sitemaps.org/schemas/sitemap/0.9/

const indentSymbol = " "

type SiteMap struct {
	config Config
	index  *Index
}

type Index struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URL     *URL     `xml:"url"`
}

type URL struct {
	Loc  string `xml:"loc"`
	URLs []*URL `xml:"url"`
}

func New(config Config) *SiteMap {
	return &SiteMap{
		config: config,
		index: &Index{
			XMLNS: config.XMLNS,
		},
	}
}

func (s *SiteMap) GenerateSitemap(data map[string][]string, rootValue string) {
	s.index.URL = generateSitemap(data, rootValue)
}

func (s *SiteMap) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	if _, err = file.WriteString(xml.Header); err != nil {
		return fmt.Errorf("failed to write xml header to file: %w", err)
	}
	xmlSitemap, err := xml.MarshalIndent(s.index, "", strings.Repeat(indentSymbol, s.config.Indent))
	if err != nil {
		return fmt.Errorf("failed to marshal sitemap: %w", err)
	}
	if _, err = file.Write(xmlSitemap); err != nil {
		return fmt.Errorf("failed to write sitemap to file: %w", err)
	}
	if cErr := file.Close(); cErr != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

type stackItem struct {
	value string
	node  *URL
}

func generateSitemap(data map[string][]string, rootValue string) *URL {
	visited := make(map[string]bool, len(data))
	rootNode := &URL{Loc: rootValue}
	stack := []*stackItem{
		{
			value: rootValue,
			node:  rootNode,
		},
	}

	for len(stack) > 0 {
		// Pop an item from the stack
		lastIdx := len(stack) - 1
		item := stack[lastIdx]
		stack = stack[:lastIdx]
		// check if visited before.
		if _, ok := visited[item.value]; ok {
			continue
		}
		// Push children to the stack
		if children, ok := data[item.value]; ok {
			for _, childValue := range children {
				child := &URL{Loc: childValue}
				item.node.URLs = append(item.node.URLs, child)
				stack = append(stack, &stackItem{
					value: childValue,
					node:  child,
				})
			}
		}
		visited[item.value] = true
	}
	return rootNode
}
