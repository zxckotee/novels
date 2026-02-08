package fanqie

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ParseBookPage(pageURL string, htmlBytes []byte) (*Book, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	book := &Book{}

	// Title (best-effort)
	if meta := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && getAttr(n, "property") == "og:title" && getAttr(n, "content") != ""
	}); meta != nil {
		book.Title = strings.TrimSpace(getAttr(meta, "content"))
	}
	if book.Title == "" {
		if h1 := findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "h1" }); h1 != nil {
			book.Title = strings.TrimSpace(nodeText(h1))
		}
	}

	// Cover (best-effort)
	if meta := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && getAttr(n, "property") == "og:image" && getAttr(n, "content") != ""
	}); meta != nil {
		book.CoverURL = strings.TrimSpace(getAttr(meta, "content"))
	}

	// Description: <div class="page-abstract-content"><p>...</p></div>
	if desc := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "page-abstract-content")
	}); desc != nil {
		var parts []string
		for _, p := range findAll(desc, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "p" }) {
			t := nodeText(p)
			if t != "" {
				parts = append(parts, t)
			}
		}
		book.Description = strings.TrimSpace(strings.Join(parts, "\n\n"))
	}

	// Chapter links: <div class="chapter-item"><a href="/reader/..." class="chapter-item-title">...</a></div>
	base, _ := url.Parse(pageURL)
	for _, a := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "chapter-item-title") && getAttr(n, "href") != ""
	}) {
		href := strings.TrimSpace(getAttr(a, "href"))
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		abs := base.ResolveReference(u).String()
		title := strings.TrimSpace(nodeText(a))
		book.Chapters = append(book.Chapters, ChapterRef{URL: abs, Title: title})
	}

	if book.Title == "" {
		book.Title = "Fanqie Novel"
	}
	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("no chapters found on page")
	}

	return book, nil
}

func ParseChapterPage(chURL string, htmlBytes []byte) (*Chapter, error) {
	root, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Title: <h1 class="muye-reader-title">...</h1>
	title := ""
	if h1 := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1" && hasClass(n, "muye-reader-title")
	}); h1 != nil {
		title = strings.TrimSpace(nodeText(h1))
	}

	// Content container: <div class="muye-reader-content noselect"> ... <p>...</p> ...</div>
	contentNode := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "muye-reader-content")
	})
	if contentNode == nil {
		return nil, fmt.Errorf("muye-reader-content not found")
	}

	var paras []string
	for _, p := range findAll(contentNode, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "p" }) {
		t := nodeText(p)
		if t != "" {
			paras = append(paras, t)
		}
	}
	content := normalizeText(strings.Join(paras, "\n\n"))
	if content == "" {
		return nil, fmt.Errorf("empty chapter content")
	}

	return &Chapter{URL: chURL, Title: title, Content: content}, nil
}

