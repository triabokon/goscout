package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// todo: introduce xml ns https://www.sitemaps.org/schemas/sitemap/0.9/

type SitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	XMLNS   string   `xml:"xmlns,attr"`
	Sitemap *Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string     `xml:"loc"`
	Sitemap []*Sitemap `xml:"sitemap"`
}

func NewSitemap(xmlName, xmlNS string) SitemapIndex {
	return SitemapIndex{
		XMLName: xml.Name{Local: xmlName},
		XMLNS:   xmlNS,
	}
}

func generateSitemap(data map[string][]string, value string) *Sitemap {
	node := &Sitemap{Loc: value}
	if children, ok := data[value]; ok {
		for _, childValue := range children {
			child := generateSitemap(data, childValue)
			node.Sitemap = append(node.Sitemap, child)
		}
	}
	return node
}

func (s *SitemapIndex) SetSitemap(data map[string][]string, baseURL string) {
	s.Sitemap = generateSitemap(data, baseURL)
}

func (s *SitemapIndex) WriteToFile(filename string, indent int) error {
	// Generate the xmlSitemap XML
	xmlSitemap, err := xml.MarshalIndent(s, "", strings.Repeat(" ", indent))
	if err != nil {
		fmt.Println("Error generating xmlSitemap:", err)
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(xml.Header))
	if err != nil {
		return err
	}

	_, err = file.Write(xmlSitemap)
	if err != nil {
		return err
	}

	return nil
}
