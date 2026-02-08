package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"novels-backend/internal/config"
	"novels-backend/internal/database"
	"novels-backend/internal/importer"
)

func main() {
	var pageURL string
	var chaptersLimit int
	var storageState string
	var cookie string
	var userAgent string

	flag.StringVar(&pageURL, "url", "", "69shuba book URL (e.g. https://www.69shuba.com/book/90474.htm)")
	flag.IntVar(&chaptersLimit, "chapters-limit", 0, "Limit number of chapters to import (0 = all found on page)")
	flag.StringVar(&storageState, "storage-state", "", "Path to Playwright storage_state JSON (exported from tools shuba-browser). If set, cookies will be derived from it.")
	flag.StringVar(&cookie, "cookie", "", "Raw Cookie header value (overrides --storage-state if set)")
	flag.StringVar(&userAgent, "user-agent", "", "User-Agent header value")
	flag.Parse()

	if pageURL == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --url is required")
		os.Exit(2)
	}

	cfg := config.Load()
	db, err := database.Connect(cfg.Database)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: connect db:", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	res, err := importer.Import69Shuba(ctx, db, importer.Import69ShubaOptions{
		PageURL:          pageURL,
		ChaptersLimit:    chaptersLimit,
		UploadDir:        cfg.UploadsDir,
		Cookie:           cookie,
		UserAgent:        userAgent,
		StorageStatePath: storageState,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	fmt.Printf("OK: novel_id=%s slug=%s chapters_saved=%d chapters_total=%d", res.NovelID.String(), res.Slug, res.ChaptersSaved, res.ChaptersTotal)
	if res.CoverKey != nil {
		fmt.Printf(" cover_key=%s", *res.CoverKey)
	}
	fmt.Println()
}

