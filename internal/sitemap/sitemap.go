package sitemap

import (
	"encoding/xml"
	"fmt"
	"os"
)

// todo: introduce xml ns https://www.sitemaps.org/schemas/sitemap/0.9/

type Index struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	Sitemap *Sitemap `xml:"url"`
}

type Sitemap struct {
	Loc     string     `xml:"loc"`
	Sitemap []*Sitemap `xml:"url"`
}

func New(xmlName, xmlNS string) Index {
	return Index{
		XMLName: xml.Name{Local: xmlName},
		XMLNS:   xmlNS,
	}
}

func (s *Index) GenerateSitemap(data map[string][]string, rootValue string) {
	s.Sitemap = generateSitemap(data, rootValue)
}

func (s *Index) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	if _, err = file.WriteString(xml.Header); err != nil {
		return fmt.Errorf("failed to write xml header to file: %w", err)
	}
	xmlSitemap, err := xml.MarshalIndent(s, "", " ")
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
	node  *Sitemap
}

func generateSitemap(data map[string][]string, rootValue string) *Sitemap {
	visited := make(map[string]bool, len(data))
	rootNode := &Sitemap{Loc: rootValue}
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
				child := &Sitemap{Loc: childValue}
				item.node.Sitemap = append(item.node.Sitemap, child)
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
