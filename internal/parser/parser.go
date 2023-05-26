package parser

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

//go:generate mockgen -destination=./mocks/http_mock.go -package=mocks github.com/triabokon/goscout/internal/parser HTTPClient
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Parser struct {
	client HTTPClient
}

func New(c HTTPClient) *Parser {
	return &Parser{client: c}
}

// ExtractURLs fetches web page by url and extracts all links from it.
func (p *Parser) ExtractURLs(u string) (webURLs, staticURLs []string, err error) {
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

// getPageTokenizer fetch the web page and get tokenizer to parse it.
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

// parseWebPage tokenizes the web page, collect and sorts the urls into web urls and static urls.
func (p *Parser) parseWebPage(tokenizer *html.Tokenizer, baseURL *url.URL) (wu, su []string, err error) {
	for {
		tt := tokenizer.Next()
		switch {
		// if the token type is an ErrorToken, we've reached the end of the document
		case tt == html.ErrorToken:
			return wu, su, nil
		case tt == html.StartTagToken, tt == html.SelfClosingTagToken:
			token := tokenizer.Token()
			switch HTMLElementType(token.DataAtom.String()) {
			// if element is a link or base element, add its urls to the web urls
			case HTMLElementTypeA, HTMLElementTypeLink, HTMLElementTypeBase:
				urls, tErr := p.handleToken(token, baseURL, HTMLAttributeTypeHref)
				if tErr != nil {
					return nil, nil, fmt.Errorf("failed to handle token: %w", tErr)
				}
				wu = append(wu, urls...)
			// if element is an image, script, source, embed, or iframe, add its urls to the static urls
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

// handleToken processes html token and extracts urls by the specified attribute type.
func (p *Parser) handleToken(token html.Token, baseURL *url.URL, attrType HTMLAttributeType) ([]string, error) {
	urls := make([]string, 0, len(token.Attr))
	for _, attr := range token.Attr {
		if HTMLAttributeType(attr.Key) == attrType {
			u, err := p.resolveURL(attr.Val, baseURL)
			switch err {
			case nil:
				urls = append(urls, u)
			case ErrURLHasInvalidSchema, ErrURLHasDifferentHost:
				// if url has invalid schema or different host, ignore it
			default:
				return nil, fmt.Errorf("failed to resolve url: %w", err)
			}
		}
	}
	return urls, nil
}

// resolveURL parse url string, resolve it relative to a baseURL, and validate it.
func (p *Parser) resolveURL(u string, baseURL *url.URL) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	parsedURL = baseURL.ResolveReference(parsedURL)
	if parsedURL.Scheme != HTTPSSchema {
		return "", ErrURLHasInvalidSchema
	}
	if parsedURL.Hostname() != baseURL.Hostname() {
		return "", ErrURLHasDifferentHost
	}
	return parsedURL.String(), nil
}
