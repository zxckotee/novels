package shuba69

type Book struct {
	Title       string
	CoverURL    string
	Description string
	Tags        []string
	CatalogURL  string
	Chapters    []ChapterRef
}

type ChapterRef struct {
	Number int
	URL   string
	Title string
}

type Chapter struct {
	URL     string
	Title   string
	Content string
}

