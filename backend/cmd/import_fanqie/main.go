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

	flag.StringVar(&pageURL, "url", "", "fanqienovel book URL (e.g. https://fanqienovel.com/page/7276384138653862966)")
	flag.IntVar(&chaptersLimit, "chapters-limit", 0, "Limit number of chapters to import (0 = all found on page)")
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	res, err := importer.ImportFanqie(ctx, db, importer.ImportFanqieOptions{
		PageURL:       pageURL,
		ChaptersLimit: chaptersLimit,
		UploadDir:     cfg.UploadsDir,
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

