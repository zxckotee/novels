package fanqie

type Book struct {
	Title       string
	CoverURL    string
	Description string
	Chapters    []ChapterRef
}

type ChapterRef struct {
	URL   string
	Title string
}

type Chapter struct {
	URL     string
	Title   string
	Content string
}

