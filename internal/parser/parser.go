package parser

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Parser struct {
	client *http.Client
}

func New(c *http.Client) *Parser {
	return &Parser{client: c}
}

func (p *Parser) ExtractURLs(u string) (outgoingURLs, staticURLs []string, err error) {
	tokenizer, err := p.getPageTokenizer(u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch web page: %w", err)
	}
	baseURL, err := url.Parse(u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	return p.parseWebPage(tokenizer, baseURL)
}

// Fetch and return the body of the webpage.
func (p *Parser) getPageTokenizer(urlStr string) (*html.Tokenizer, error) {
	resp, err := p.client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get web page: %w", err)
	}
	var body bytes.Buffer
	if wErr := resp.Write(&body); wErr != nil {
		return nil, fmt.Errorf("failed to write response to buffer: %w", wErr)
	}
	defer resp.Body.Close()
	return html.NewTokenizer(&body), nil
}

// parseWebPage is now simplified, delegating the handling of different token types to the new functions.
func (p *Parser) parseWebPage(tokenizer *html.Tokenizer, baseURL *url.URL) (nu, su []string, err error) {
	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			return nu, su, nil
		case tt == html.StartTagToken, tt == html.SelfClosingTagToken:
			token := tokenizer.Token()
			switch HTMLElementType(token.DataAtom.String()) {
			case HTMLElementTypeA, HTMLElementTypeLink, HTMLElementTypeBase:
				urls, tErr := p.handleToken(token, baseURL, HTMLAttributeTypeHref)
				if tErr != nil {
					return nil, nil, fmt.Errorf("failed to handle token: %w", tErr)
				}
				nu = append(nu, urls...)
			case HTMLElementTypeImg, HTMLElementTypeImage, HTMLElementTypeScript,
				HTMLElementTypeSource, HTMLElementTypeEmbed, HTMLElementTypeIFrame:
				urls, tErr := p.handleToken(token, baseURL, HTMLAttributeTypeSrc)
				if tErr != nil {
					return nil, nil, fmt.Errorf("failed to handle token: %w", tErr)
				}
				su = append(su, urls...)
			}
		}
	}
}

// handleToken handles tokens attributes.
func (p *Parser) handleToken(token html.Token, baseURL *url.URL, attrType HTMLAttributeType) ([]string, error) {
	urls := make([]string, 0, len(token.Attr))
	for _, attr := range token.Attr {
		if HTMLAttributeType(attr.Key) == attrType {
			u, err := p.resolveURL(attr.Val, baseURL)
			switch err {
			case nil:
				urls = append(urls, u)
			case ErrURLHasInvalidSchema, ErrURLHasDifferentHost:
			default:
				return nil, fmt.Errorf("failed to resolve url: %w", err)
			}
		}
	}
	return urls, nil
}

// resolveURL resolves the url with respect to the base URL.
func (p *Parser) resolveURL(u string, baseURL *url.URL) (string, error) {
	parsedURL, err := url.Parse(strings.Trim(u, " "))
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	parsedURL = baseURL.ResolveReference(parsedURL)
	if parsedURL.Scheme != "https" {
		return "", ErrURLHasInvalidSchema
	}
	if parsedURL.Hostname() != baseURL.Hostname() {
		return "", ErrURLHasDifferentHost
	}
	return parsedURL.String(), nil
}
