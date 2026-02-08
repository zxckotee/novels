package kks101

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var reChapterHref = regexp.MustCompile(`(?i)^/txt/\d+/\d+\.html$`)
var reBookinfoTags = regexp.MustCompile(`tags\s*:\s*'([^']*)'`)

func ParseBookPage(pageURL string, htmlBytes []byte) (*Book, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	book := &Book{}

	book.Title = metaProp(root, "og:title")
	book.CoverURL = metaProp(root, "og:image")
	book.Description = metaProp(root, "og:description")
	book.Author = metaProp(root, "og:novel:author")
	book.Category = metaProp(root, "og:novel:category")
	book.CatalogURL = metaProp(root, "og:novel:read_url") // https://101kks.com/book/{id}/index.html

	// Tags from JS `var bookinfo = { tags: 'a,b,c,' }`
	for _, s := range findAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "script" }) {
		t := nodeText(s)
		if t == "" {
			continue
		}
		m := reBookinfoTags.FindStringSubmatch(t)
		if len(m) != 2 {
			continue
		}
		raw := strings.TrimSpace(m[1])
		if raw == "" {
			continue
		}
		var tags []string
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				tags = append(tags, p)
			}
		}
		book.Tags = dedupKeep(tags)
		break
	}

	// Fallback: the page has a "完整目錄" link too.
	if strings.TrimSpace(book.CatalogURL) == "" {
		if a := findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "more-btn") && strings.TrimSpace(getAttr(n, "href")) != ""
		}); a != nil {
			base, _ := url.Parse(pageURL)
			href := strings.TrimSpace(getAttr(a, "href"))
			if u, err := url.Parse(href); err == nil && base != nil {
				book.CatalogURL = base.ResolveReference(u).String()
			}
		}
	}

	if strings.TrimSpace(book.Title) == "" {
		return nil, fmt.Errorf("book title not found")
	}

	book.Description = normalizeText(book.Description)
	return book, nil
}

func ParseCatalogPage(pageURL string, htmlBytes []byte) (*Book, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	book := &Book{CatalogURL: pageURL}

	// Title may exist in og:title on index page too; not guaranteed.
	book.Title = metaProp(root, "og:title")
	if strings.TrimSpace(book.Title) == "" {
		if h1 := findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "h1" }); h1 != nil {
			book.Title = strings.TrimSpace(nodeText(h1))
		}
	}

	base, _ := url.Parse(pageURL)

	// Best-effort: collect all /txt/{book}/{chapter}.html links.
	seen := map[string]bool{}
	for _, a := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && strings.TrimSpace(getAttr(n, "href")) != ""
	}) {
		href := strings.TrimSpace(getAttr(a, "href"))
		if href == "" {
			continue
		}

		// normalize to absolute
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		abs := href
		if base != nil {
			abs = base.ResolveReference(u).String()
		}
		// path match
		pu, err := url.Parse(abs)
		if err != nil {
			continue
		}
		if !reChapterHref.MatchString(pu.Path) {
			continue
		}

		if seen[abs] {
			continue
		}
		seen[abs] = true

		title := strings.TrimSpace(nodeText(a))
		// try parse chapter number from "第123章" prefix if present
		num := parseChapterNumber(title)
		book.Chapters = append(book.Chapters, ChapterRef{Number: num, URL: abs, Title: title})
	}

	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("no chapters found on catalog page")
	}
	return book, nil
}

func ParseChapterPage(chURL string, htmlBytes []byte) (*Chapter, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	txtnav := findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "txtnav") })
	if txtnav == nil {
		return nil, fmt.Errorf("txtnav not found")
	}

	title := ""
	if h1 := findFirst(txtnav, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "h1" }); h1 != nil {
		title = strings.TrimSpace(nodeText(h1))
	}

	txtcontent := findFirst(txtnav, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && strings.EqualFold(getAttr(n, "id"), "txtcontent")
	})
	if txtcontent == nil {
		// fallback: sometimes content container might be directly in txtnav
		txtcontent = txtnav
	}

	content := extractContent(txtcontent)
	content = normalizeText(content)

	if title != "" {
		lines := strings.Split(content, "\n")
		if len(lines) > 0 && strings.TrimSpace(lines[0]) == strings.TrimSpace(title) {
			content = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		}
	}

	if content == "" {
		return nil, fmt.Errorf("empty chapter content")
	}
	return &Chapter{URL: chURL, Title: title, Content: content}, nil
}

func extractContent(root *html.Node) string {
	var b strings.Builder

	skip := func(n *html.Node) bool {
		if n == nil || n.Type != html.ElementNode {
			return false
		}
		if n.Data == "script" {
			return true
		}
		if hasClass(n, "txtad") || hasClass(n, "txtcenter") || hasClass(n, "bottom-ad") || hasClass(n, "page1") {
			return true
		}
		return false
	}

	var dfs func(*html.Node)
	dfs = func(n *html.Node) {
		if n == nil {
			return
		}
		if skip(n) {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
			if c.Type == html.ElementNode && (c.Data == "br") {
				b.WriteString("\n")
			}
		}
	}
	dfs(root)
	return b.String()
}

func metaProp(root *html.Node, prop string) string {
	if prop == "" {
		return ""
	}
	if meta := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && strings.EqualFold(getAttr(n, "property"), prop) && strings.TrimSpace(getAttr(n, "content")) != ""
	}); meta != nil {
		return strings.TrimSpace(getAttr(meta, "content"))
	}
	return ""
}

func parseChapterNumber(title string) int {
	// Very small heuristic: find "第" ... "章" and parse digits in-between.
	title = strings.TrimSpace(title)
	i := strings.Index(title, "第")
	j := strings.Index(title, "章")
	if i == -1 || j == -1 || j <= i {
		return 0
	}
	mid := title[i+len("第") : j]
	mid = strings.TrimSpace(mid)
	n, err := strconv.Atoi(mid)
	if err != nil {
		return 0
	}
	return n
}

func dedupKeep(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

