package main

import (
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type Parser struct {
	client *http.Client
}

func NewParser(c *http.Client) *Parser {
	return &Parser{c}
}

// Fetch and return the body of the webpage
func (p *Parser) fetchWebpage(urlStr string) (*html.Tokenizer, func() error, error) {
	resp, err := p.client.Get(urlStr)
	if err != nil {
		return nil, nil, err
	}
	tokenizer := html.NewTokenizer(resp.Body)
	return tokenizer, resp.Body.Close, nil
}

// Parse the HTML and return a slice of outgoing links
func (p *Parser) parseHTML(
	tokenizer *html.Tokenizer,
	baseURL *url.URL,
) (outgoingLinks []string, staticLinks []string, err error) {
	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			return outgoingLinks, staticLinks, nil
		case tt == html.StartTagToken, tt == html.SelfClosingTagToken:
			token := tokenizer.Token()
			switch token.DataAtom.String() {
			case "a", "link":
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						outgoingLink := p.resolveLink(attr.Val, baseURL)
						if outgoingLink != "" {
							outgoingLinks = append(outgoingLinks, outgoingLink)
						}
					}
				}
			case "img", "script", "image":
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						staticLink := p.resolveLink(attr.Val, baseURL)
						if staticLink != "" {
							staticLinks = append(staticLinks, staticLink)
						}
					}
				}
			}
		}
	}
}

// Resolve the link with respect to the base URL
func (p *Parser) resolveLink(link string, baseURL *url.URL) string {
	outgoingURL, err := url.Parse(link)
	if err != nil {
		return ""
	}
	outgoingURL = baseURL.ResolveReference(outgoingURL)
	if outgoingURL.Scheme != "https" {
		return ""
	}
	if outgoingURL.Hostname() != baseURL.Hostname() {
		return ""
	}
	return outgoingURL.String()
}

func (p *Parser) ExtractAllLinks(urlStr string) ([]string, []string, error) {
	tokenizer, closer, err := p.fetchWebpage(urlStr)
	if err != nil {
		return nil, nil, err
	}
	defer closer()

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}

	return p.parseHTML(tokenizer, baseURL)
}
