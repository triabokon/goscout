package parser

import "fmt"

var (
	ErrURLHasDifferentHost = fmt.Errorf("url has different host")
	ErrURLHasInvalidSchema = fmt.Errorf("url has invalid schema")
)

const HTTPSSchema = "https"

type HTMLElementType string

const (
	HTMLElementTypeA    HTMLElementType = "a"
	HTMLElementTypeLink HTMLElementType = "link"
	HTMLElementTypeBase HTMLElementType = "base"

	HTMLElementTypeIFrame HTMLElementType = "iframe"
	HTMLElementTypeEmbed  HTMLElementType = "embed"
	HTMLElementTypeImg    HTMLElementType = "img"
	HTMLElementTypeImage  HTMLElementType = "image"
	HTMLElementTypeScript HTMLElementType = "script"
	HTMLElementTypeSource HTMLElementType = "source"
)

type HTMLAttributeType string

const (
	HTMLAttributeTypeHref HTMLAttributeType = "href"
	HTMLAttributeTypeSrc  HTMLAttributeType = "src"
)
