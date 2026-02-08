package shuba69

import (
	"strings"

	"golang.org/x/net/html"
)

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, className string) bool {
	if n == nil {
		return false
	}
	class := getAttr(n, "class")
	if class == "" {
		return false
	}
	for _, c := range strings.Fields(class) {
		if c == className {
			return true
		}
	}
	return false
}

func findFirst(root *html.Node, pred func(*html.Node) bool) *html.Node {
	var dfs func(*html.Node) *html.Node
	dfs = func(n *html.Node) *html.Node {
		if n == nil {
			return nil
		}
		if pred(n) {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if got := dfs(c); got != nil {
				return got
			}
		}
		return nil
	}
	return dfs(root)
}

func findAll(root *html.Node, pred func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var dfs func(*html.Node)
	dfs = func(n *html.Node) {
		if n == nil {
			return
		}
		if pred(n) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(root)
	return out
}

func nodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var dfs func(*html.Node)
	dfs = func(x *html.Node) {
		if x == nil {
			return
		}
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
			// treat <br> as newline separator
			if c.Type == html.ElementNode && c.Data == "br" {
				b.WriteString("\n")
			}
		}
	}
	dfs(n)
	return strings.TrimSpace(b.String())
}

