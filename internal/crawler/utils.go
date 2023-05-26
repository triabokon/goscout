package crawler

import (
	"fmt"
	"mime"
	"net/url"
	"path"
	"strings"
	"sync"
)

type WebPageExtensionType string

const (
	WebPageExtensionTypeAspx = ".aspx"
	WebPageExtensionTypeAsp  = ".asp"
	WebPageExtensionTypePhp  = ".php"
	WebPageExtensionTypeJsp  = ".jsp"
	WebPageExtensionTypeErb  = ".erb"
)

// unique removes duplicates.
func unique(s []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// getExtension returns the file extension of the path in the url.
func getExtension(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	return path.Ext(parsedURL.Path), nil
}

// registerExtensionTypes registers all additional extensions that could have text/html web pages.
func registerExtensionTypes() error {
	allTypes := []string{
		WebPageExtensionTypeAspx, WebPageExtensionTypeAsp, WebPageExtensionTypePhp,
		WebPageExtensionTypeJsp, WebPageExtensionTypeErb,
	}
	for _, ext := range allTypes {
		if err := mime.AddExtensionType(ext, "text/html"); err != nil {
			return fmt.Errorf("failed to add extension type: %w", err)
		}
	}
	return nil
}

// isTextURL checks is url has extension and if so, whether it's mime type is text.
func isTextURL(u string) (bool, error) {
	extension, err := getExtension(u)
	if err != nil {
		return false, fmt.Errorf("failed to get extension: %w", err)
	}
	if extension == "" {
		return true, nil
	}
	if tErr := registerExtensionTypes(); tErr != nil {
		return false, fmt.Errorf("failed to register extensions: %w", tErr)
	}
	mimeType := mime.TypeByExtension(extension)
	return strings.HasPrefix(mimeType, "text/"), nil
}

// filterWebURLs filters visited web urls and urls that has wrong type.
func filterWebURLs(urls []string, seenURLs *sync.Map) ([]string, error) {
	filtered := make([]string, 0, len(urls))
	for _, u := range unique(urls) {
		if _, ok := seenURLs.Load(u); ok {
			continue
		}
		textLink, err := isTextURL(u)
		if err != nil {
			return nil, fmt.Errorf("failed to check url type: %w", err)
		}
		if textLink {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

// filterWebURLs filters static urls that has wrong type.
func filterStaticURLs(urls []string) ([]string, error) {
	filtered := make([]string, 0, len(urls))
	for _, u := range unique(urls) {
		textLink, err := isTextURL(u)
		if err != nil {
			return nil, fmt.Errorf("failed to check url type: %w", err)
		}
		if !textLink {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func seenURLsToMap(seenURLs *sync.Map) map[string][]string {
	result := make(map[string][]string)
	seenURLs.Range(func(key, value interface{}) bool {
		if strKey, ok := key.(string); ok {
			if strSlice, ok := value.([]string); ok {
				result[strKey] = strSlice
			}
		}
		return true
	})
	return result
}
