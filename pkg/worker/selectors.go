package worker

import (
	"bytes"
	"io"
	"strings"
	"sync"

	"github.com/andybalholm/cascadia"
	"github.com/antchfx/htmlquery"
	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/net/html"
)

// cache for css selectors.
var cssCache sync.Map

type Selectors []*html.Node

func NewSelector(r io.Reader) (Selectors, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return NewSelectorFromBytes(data)
}

// NewSelectorFromBytes - creates a selector from byte slice.
func NewSelectorFromBytes(data []byte) (Selectors, error) {
	root, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return Selectors([]*html.Node{root}), nil
}

// Css selector - select element by id, class, nodename etc.
func (nodes Selectors) Css(selector string) core.ISelector {
	var sel cascadia.Sel
	if cached, ok := cssCache.Load(selector); ok {
		sel = cached.(cascadia.Sel)
	} else {
		parsed, err := cascadia.ParseWithPseudoElement(selector)
		if err != nil {
			return Selectors([]*html.Node{})
		}
		sel = parsed
		cssCache.Store(selector, sel)
	}

	selected := make(Selectors, 0, len(nodes))
	for _, node := range nodes {
		selected = append(selected, cascadia.QueryAll(node, sel)...)
	}

	return selected
}

// Xpath selector - select element using an xpath expression.
func (nodes Selectors) Xpath(xpath string) core.ISelector {
	selected := make(Selectors, 0, len(nodes))
	for _, node := range nodes {
		matches, err := htmlquery.QueryAll(node, xpath)
		if err != nil {
			continue
		}
		selected = append(selected, matches...)
	}
	return selected
}

// Extracts all the text of a node and its descendants.
func (nodes Selectors) Text(def ...string) []string {
	texts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		text := strings.TrimSpace(innerText(node))
		if text == "" && len(def) > 0 {
			texts = append(texts, def[0])
			continue
		}
		texts = append(texts, text)
	}
	return texts
}

// Extracts text from a node.
func innerText(n *html.Node) string {
	var b strings.Builder
	collectText(n, &b)
	return b.String()
}

func collectText(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		b.WriteString(n.Data)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, b)
	}
}

// Extracts attribute values
func (nodes Selectors) Attr(attrName string) []string {
	attrs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		for _, attr := range node.Attr {
			if attr.Key == attrName {
				attrs = append(attrs, attr.Val)
			}
		}
	}
	return attrs
}

// Get the first matched node
func (nodes Selectors) Get() *html.Node {
	if len(nodes) <= 0 {
		return nil
	}
	return nodes[0]
}

// Gets all the matched nodes
func (nodes Selectors) GetAll() []*html.Node {
	return nodes
}
