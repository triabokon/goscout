package crawler

import (
	"mime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawlerUtils_Unique(t *testing.T) {
	testCases := []struct {
		name           string
		input          []string
		expectedOutput []string
	}{
		{
			name:           "no duplicates",
			input:          []string{"apple", "banana", "cherry"},
			expectedOutput: []string{"apple", "banana", "cherry"},
		},
		{
			name:           "with duplicates",
			input:          []string{"apple", "banana", "cherry", "apple", "banana"},
			expectedOutput: []string{"apple", "banana", "cherry"},
		},
		{
			name:           "all duplicates",
			input:          []string{"apple", "apple", "apple", "apple", "apple"},
			expectedOutput: []string{"apple"},
		},
		{
			name:           "empty list",
			input:          []string{},
			expectedOutput: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, unique(tc.input), tc.expectedOutput)
		})
	}
}

func TestCrawlerUtils_GetExtension(t *testing.T) {
	testCases := []struct {
		name             string
		url              string
		expectedResult   string
		expectedErrorMsg string
	}{
		{
			name:             "valid url with extension",
			url:              "https://example.com/path/to/file.txt",
			expectedResult:   ".txt",
			expectedErrorMsg: "",
		},
		{
			name:             "valid url without extension",
			url:              "https://example.com/path/to/file",
			expectedResult:   "",
			expectedErrorMsg: "",
		},
		{
			name:             "invalid url",
			url:              ":",
			expectedResult:   "",
			expectedErrorMsg: "failed to parse url: parse \":\": missing protocol scheme",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getExtension(tc.url)
			if err != nil {
				assert.Equal(t, err.Error(), tc.expectedErrorMsg)
			}
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestCrawlerUtils_RegisterExtensionTypes(t *testing.T) {
	rErr := registerExtensionTypes()
	assert.NoError(t, rErr)

	allTypes := []string{
		WebPageExtensionTypeAspx, WebPageExtensionTypeAsp, WebPageExtensionTypePhp,
		WebPageExtensionTypeJsp, WebPageExtensionTypeErb,
	}
	for _, ext := range allTypes {
		mimeType, err := mime.ExtensionsByType("text/html")
		assert.NoError(t, err)
		assert.Contains(t, mimeType, ext)
	}
}

func TestCrawlerUtils_IsTextURL(t *testing.T) {
	testCases := []struct {
		name             string
		url              string
		expectedResult   bool
		expectedErrorMsg string
	}{
		{
			name:             "valid url with htm extension",
			url:              "https://example.com/path/to/file.htm",
			expectedResult:   true,
			expectedErrorMsg: "",
		},
		{
			name:             "valid url without extension",
			url:              "https://example.com/path/to/file",
			expectedResult:   true,
			expectedErrorMsg: "",
		},
		{
			name:             "valid url with non-text extension",
			url:              "https://example.com/path/to/file.jpg",
			expectedResult:   false,
			expectedErrorMsg: "",
		},
		{
			name:             "invalid url",
			url:              ":",
			expectedResult:   false,
			expectedErrorMsg: "failed to get extension: failed to parse url: parse \":\": missing protocol scheme",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := isTextURL(tc.url)
			if err != nil {
				assert.Equal(t, err.Error(), tc.expectedErrorMsg)
			}
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestCrawlerUtils_FilterWebURLs(t *testing.T) {
	testCases := []struct {
		name           string
		urls           []string
		seenURLs       func(s *sync.Map, url string)
		expectedResult []string
	}{
		{
			name:           "unique web urls",
			urls:           []string{"https://example.com/someurl", "https://example.com/someurl1.htm"},
			seenURLs:       func(s *sync.Map, url string) {},
			expectedResult: []string{"https://example.com/someurl", "https://example.com/someurl1.htm"},
		},
		{
			name:           "some static urls",
			urls:           []string{"https://example.com/script.js", "https://example.com/someurl"},
			seenURLs:       func(s *sync.Map, url string) {},
			expectedResult: []string{"https://example.com/someurl"},
		},
		{
			name:           "duplicated text urls",
			urls:           []string{"https://example.com/someurl", "https://example.com/someurl"},
			seenURLs:       func(s *sync.Map, url string) {},
			expectedResult: []string{"https://example.com/someurl"},
		},
		{
			name: "seen html urls",
			urls: []string{"https://example.com/file1.html", "https://example.com/file2.html"},
			seenURLs: func(s *sync.Map, url string) {
				s.Store(url, nil)
			},
			expectedResult: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &sync.Map{}
			for _, u := range tc.urls {
				tc.seenURLs(s, u)
			}
			result, err := filterWebURLs(tc.urls, s)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestCrawlerUtils_FilterStaticURLs(t *testing.T) {
	testCases := []struct {
		name           string
		urls           []string
		expectedResult []string
	}{
		{
			name:           "unique static urls",
			urls:           []string{"https://example.com/script.js", "https://example.com/image.png"},
			expectedResult: []string{"https://example.com/script.js", "https://example.com/image.png"},
		},
		{
			name:           "some web urls",
			urls:           []string{"https://example.com/someurl", "https://example.com/script.js"},
			expectedResult: []string{"https://example.com/script.js"},
		},
		{
			name:           "duplicated text urls",
			urls:           []string{"https://example.com/script.js", "https://example.com/script.js"},
			expectedResult: []string{"https://example.com/script.js"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filterStaticURLs(tc.urls)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
