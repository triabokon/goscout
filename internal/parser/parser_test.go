package parser

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"

	"github.com/triabokon/goscout/internal/parser/mocks"
)

var gfi = gofakeit.New(1)

func TestParser_GetPageTokenizer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)

	mockClient.EXPECT().Get(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("<html><body>Test</body></html>")),
	}, nil)

	p := New(mockClient)

	tokenizer, err := p.getPageTokenizer(gfi.URL())
	assert.NoError(t, err)

	assert.Equal(t, html.TextToken, tokenizer.Next())
	assert.Equal(t, html.StartTagToken, tokenizer.Next())

	tokenName, _ := tokenizer.TagName()
	assert.Equal(t, "html", string(tokenName))
}

func TestParser_ParseWebPage(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "link elements",
			html:     `<html><body><a href="/page">Link</a><link href="/css/style.css"></body></html>`,
			expected: []string{"https://example.com/page", "https://example.com/css/style.css"},
		},
		{
			name:     "img elements",
			html:     `<html><body><img src="/img/image.jpg"></body></html>`,
			expected: []string{"https://example.com/img/image.jpg"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := New(nil)
			baseURL, pErr := url.Parse("https://example.com")
			assert.NoError(t, pErr)
			tokenizer := html.NewTokenizer(strings.NewReader(tc.html))

			webUrls, staticUrls, err := parser.parseWebPage(tokenizer, baseURL)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, append(webUrls, staticUrls...))
		})
	}
}

func TestParser_HandleToken(t *testing.T) {
	testCases := []struct {
		name       string
		token      html.Token
		attrType   HTMLAttributeType
		expectUrls []string
	}{
		{
			name:     "find href attribute",
			attrType: HTMLAttributeTypeHref,
			token: html.Token{
				Type: html.StartTagToken,
				Data: string(HTMLElementTypeA),
				Attr: []html.Attribute{
					{Key: string(HTMLAttributeTypeHref), Val: "https://example.com/page"},
				},
			},
			expectUrls: []string{"https://example.com/page"},
		},
		{
			name:     "find src attribute",
			attrType: HTMLAttributeTypeSrc,
			token: html.Token{
				Type: html.StartTagToken,
				Data: string(HTMLElementTypeImg),
				Attr: []html.Attribute{
					{Key: string(HTMLAttributeTypeSrc), Val: "/img/image.jpg"},
				},
			},
			expectUrls: []string{"https://example.com/img/image.jpg"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := New(nil)
			baseURL, pErr := url.Parse("https://example.com")
			assert.NoError(t, pErr)
			urls, err := parser.handleToken(tc.token, baseURL, tc.attrType)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectUrls, urls)
		})
	}
}

func TestParser_ResolveURL(t *testing.T) {
	testCases := []struct {
		name             string
		urlStr           string
		expectURL        string
		expectedErrorMsg string
	}{
		{
			name:      "valid relative url",
			urlStr:    "/page",
			expectURL: "https://example.com/page",
		},
		{
			name:      "valid absolute url",
			urlStr:    "https://example.com/page",
			expectURL: "https://example.com/page",
		},
		{
			name:             "invalid url",
			urlStr:           ":",
			expectURL:        "",
			expectedErrorMsg: "failed to parse url: parse \":\": missing protocol scheme",
		},
		{
			name:             "different host",
			urlStr:           "https://other.com/page",
			expectURL:        "",
			expectedErrorMsg: ErrURLHasDifferentHost.Error(),
		},
		{
			name:             "invalid schema",
			urlStr:           "htp://example.com/page",
			expectURL:        "",
			expectedErrorMsg: ErrURLHasInvalidSchema.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := New(nil)
			baseURL, pErr := url.Parse("https://example.com")
			assert.NoError(t, pErr)

			resolvedURL, err := parser.resolveURL(tc.urlStr, baseURL)
			if err != nil {
				assert.Equal(t, err.Error(), tc.expectedErrorMsg)
			}
			assert.Equal(t, tc.expectURL, resolvedURL)
		})
	}
}

func TestParser_ExtractURLs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	u := "https://example.com"
	expectedWebURLs := []string{"https://example.com/link1", "https://example.com/link2"}
	expectedStaticURLs := []string{"https://example.com/image1", "https://example.com/image2"}

	mockResponse := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(`
            <html>
            <body>
                <a href="/link1">Link 1</a>
                <a href="/link2">Link 2</a>
                <img src="/image1" />
                <img src="/image2" />
            </body>
            </html>
        `)),
	}
	mockClient := mocks.NewMockHTTPClient(mockCtrl)
	mockClient.EXPECT().Get(u).Return(mockResponse, nil).Times(1)

	p := New(mockClient)
	webURLs, staticURLs, err := p.ExtractURLs(u)
	assert.Nil(t, err)
	assert.Equal(t, expectedWebURLs, webURLs)
	assert.Equal(t, expectedStaticURLs, staticURLs)
}
