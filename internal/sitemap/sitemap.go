package sitemap

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

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

// WriteToFile writes the xml site map to a file with filename.
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

// generateSitemap builds sitemap as a tree structure from given map.
func generateSitemap(data map[string][]string, rootValue string) *URL {
	visited := make(map[string]bool, len(data))
	rootNode := &URL{Loc: rootValue}
	// stack is used to process the urls in a depth-first search manner
	stack := []*stackItem{{value: rootValue, node: rootNode}}
	for len(stack) > 0 {
		// pop item from the stack
		lastIdx := len(stack) - 1
		item := stack[lastIdx]
		stack = stack[:lastIdx]
		// check if the url has already been visited
		if _, ok := visited[item.value]; ok {
			continue
		}
		// if the url has children, add them to the node and stack
		if children, ok := data[item.value]; ok {
			for _, childValue := range children {
				child := &URL{Loc: childValue}
				item.node.URLs = append(item.node.URLs, child)
				stack = append(stack, &stackItem{value: childValue, node: child})
			}
		}
		visited[item.value] = true
	}
	return rootNode
}
