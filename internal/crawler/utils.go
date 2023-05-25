package crawler

import (
	"fmt"
	"mime"
	"net/url"
	"path"
	"strings"
	"sync"
)

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

func getExtension(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	return path.Ext(parsedURL.Path), nil
}

func isTextLink(u string) (bool, error) {
	extension, err := getExtension(u)
	if err != nil {
		return false, fmt.Errorf("failed to get extension: %w", err)
	}
	if extension == "" {
		return true, nil
	}
	mimeType := mime.TypeByExtension(extension)
	return strings.HasPrefix(mimeType, "text/"), nil
}

func filterStaticLinks(urls []string) ([]string, error) {
	filtered := make([]string, 0, len(urls))
	for _, l := range unique(urls) {
		textLink, err := isTextLink(l)
		if err != nil {
			return nil, fmt.Errorf("failed to check link type: %w", err)
		}
		if !textLink {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

func syncMapToMap(syncMap *sync.Map) map[string][]string {
	result := make(map[string][]string)
	syncMap.Range(func(key, value interface{}) bool {
		if strKey, ok := key.(string); ok {
			if strSlice, ok := value.([]string); ok {
				result[strKey] = strSlice
			}
		}
		return true
	})
	return result
}
