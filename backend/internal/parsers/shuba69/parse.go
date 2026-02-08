package shuba69

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var reLooksLikeChapterHref = regexp.MustCompile(`(?i)/(txt/\\d+/\\d+|txt/\\d+/\\d+/|txt/\\d+/\\d+\\?.*|txt/\\d+/\\d+#.*)$`)

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func ParseBookPage(pageURL string, htmlBytes []byte) (*Book, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	book := &Book{}

	// Parse within main container if present: <div class="mybox">
	scope := findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "mybox") })
	if scope == nil {
		scope = root
	}

	// Title (best-effort): og:title -> <h1 class="muluh1"> -> h1
	if meta := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" &&
			strings.EqualFold(getAttr(n, "property"), "og:title") &&
			strings.TrimSpace(getAttr(n, "content")) != ""
	}); meta != nil {
		book.Title = strings.TrimSpace(getAttr(meta, "content"))
	}
	if book.Title == "" {
		if h1 := findFirst(scope, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "h1" && hasClass(n, "muluh1")
		}); h1 != nil {
			book.Title = strings.TrimSpace(nodeText(h1))
		}
	}
	if book.Title == "" {
		if h1 := findFirst(scope, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "h1" }); h1 != nil {
			book.Title = strings.TrimSpace(nodeText(h1))
		}
	}

	// Cover
	if meta := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" &&
			strings.EqualFold(getAttr(n, "property"), "og:image") &&
			strings.TrimSpace(getAttr(n, "content")) != ""
	}); meta != nil {
		book.CoverURL = strings.TrimSpace(getAttr(meta, "content"))
	}

	// Tags: <ul id="tagul"> <a>...</a> ... </ul>
	if tagul := findFirst(scope, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "ul" && strings.EqualFold(getAttr(n, "id"), "tagul")
	}); tagul != nil {
		seen := map[string]bool{}
		for _, a := range findAll(tagul, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "a" }) {
			t := strings.TrimSpace(nodeText(a))
			if t == "" || seen[t] {
				continue
			}
			seen[t] = true
			book.Tags = append(book.Tags, t)
		}
	}

	// Description: <div class="navtxt"><p>...<br>...</p>...</div> (best-effort)
	if meta := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" &&
			strings.EqualFold(getAttr(n, "property"), "og:description") &&
			strings.TrimSpace(getAttr(n, "content")) != ""
	}); meta != nil {
		book.Description = strings.TrimSpace(getAttr(meta, "content"))
	}
	if book.Description == "" {
		if navtxt := findFirst(scope, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "navtxt") }); navtxt != nil {
			// take first paragraph as description; second paragraph often "关键词"
			ps := findAll(navtxt, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "p" })
			if len(ps) > 0 {
				book.Description = strings.TrimSpace(nodeText(ps[0]))
			} else {
				book.Description = strings.TrimSpace(nodeText(navtxt))
			}
		}
	}

	// Catalog URL: <a class="btn more-btn" href="https://www.69shuba.com/book/90488/">完整目录</a>
	if a := findFirst(scope, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "more-btn") && strings.TrimSpace(getAttr(n, "href")) != ""
	}); a != nil {
		base, _ := url.Parse(pageURL)
		href := strings.TrimSpace(getAttr(a, "href"))
		if u, err := url.Parse(href); err == nil && base != nil {
			book.CatalogURL = base.ResolveReference(u).String()
		}
	}

	// This page usually only contains "latest chapters" (qustime). We don't require chapters here.
	if strings.TrimSpace(book.Title) == "" {
		book.Title = "69shuba Novel"
	}
	book.Description = normalizeText(book.Description)

	return book, nil
}

// ParseCatalogPage parses https://www.69shuba.com/book/{bookID}/ (full chapter list)
// Per parsers.md: <div class="catalog" id="catalog"> <ul><li data-num="1"><a href="...">title</a></li>...</ul>
func ParseCatalogPage(pageURL string, htmlBytes []byte) (*Book, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	book := &Book{CatalogURL: pageURL}
	scope := findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "mybox") })
	if scope == nil {
		scope = root
	}

	// Title in <h1 class="muluh1"><a ...>XXX最新章节</a></h1>
	if h1 := findFirst(scope, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "h1" && hasClass(n, "muluh1") }); h1 != nil {
		book.Title = strings.TrimSpace(nodeText(h1))
		// strip suffix if present
		book.Title = strings.TrimSuffix(book.Title, "最新章节")
		book.Title = strings.TrimSpace(book.Title)
	}

	catalog := findFirst(scope, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && strings.EqualFold(getAttr(n, "id"), "catalog")
	})
	if catalog == nil {
		return nil, fmt.Errorf("catalog not found")
	}

	base, _ := url.Parse(pageURL)
	seen := map[string]bool{}
	for _, li := range findAll(catalog, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "li" }) {
		a := findFirst(li, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "a" && strings.TrimSpace(getAttr(n, "href")) != "" })
		if a == nil {
			continue
		}
		href := strings.TrimSpace(getAttr(a, "href"))
		if href == "" {
			continue
		}
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		abs := href
		if base != nil {
			abs = base.ResolveReference(u).String()
		}
		if !reLooksLikeChapterHref.MatchString(abs) {
			// still accept: it is normally /txt/...
		}
		if abs == "" || seen[abs] {
			continue
		}
		seen[abs] = true
		title := strings.TrimSpace(nodeText(a))
		num := 0
		if v := strings.TrimSpace(getAttr(li, "data-num")); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				num = n
			}
		}
		book.Chapters = append(book.Chapters, ChapterRef{Number: num, URL: abs, Title: title})
	}

	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("no chapters found in catalog")
	}
	return book, nil
}

func ParseChapterPage(chURL string, htmlBytes []byte) (*Chapter, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Per parsers.md: content lives under <div class="txtnav"> and title under <h1 class="hide720"> inside it.
	txtnav := findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "txtnav") })
	if txtnav == nil {
		return nil, fmt.Errorf("txtnav not found")
	}

	title := ""
	if h1 := findFirst(txtnav, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "h1" }); h1 != nil {
		title = strings.TrimSpace(nodeText(h1))
	}

	content := extractTxtnavContent(txtnav)
	content = normalizeText(content)
	if title != "" {
		// Drop duplicated first line if it equals title
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

func extractTxtnavContent(txtnav *html.Node) string {
	var b strings.Builder

	skipNode := func(n *html.Node) bool {
		if n == nil || n.Type != html.ElementNode {
			return false
		}
		// skip scripts/ads and known non-content blocks
		if n.Data == "script" {
			return true
		}
		if hasClass(n, "yueduad1") || hasClass(n, "contentadv") || hasClass(n, "bottom-ad") || hasClass(n, "bottom-ad2") {
			return true
		}
		if hasClass(n, "tools") || hasClass(n, "page1") || hasClass(n, "baocuo") || hasClass(n, "txtinfo") {
			return true
		}
		if strings.EqualFold(getAttr(n, "id"), "txtright") {
			return true
		}
		return false
	}

	var dfs func(*html.Node)
	dfs = func(n *html.Node) {
		if n == nil {
			return
		}
		if skipNode(n) {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
			if c.Type == html.ElementNode && c.Data == "br" {
				b.WriteString("\n")
			}
		}
	}

	dfs(txtnav)

	// normalize odd fullwidth/ideographic spaces used by site (e.g. "  ")
	out := strings.ReplaceAll(b.String(), "\u00a0", " ")
	out = strings.ReplaceAll(out, "\u3000", " ")
	out = strings.ReplaceAll(out, "\u2003", " ")
	return out
}
