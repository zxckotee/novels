#!/usr/bin/env python3
"""
Import MangaLib projects parsed by newmaga-parser into Novels (as text novels),
creating stub chapter contents: "Скоро будет".

Reads:  newmaga-parser SQLite DB (mangalib_projects, mangalib_chapters)
Copies covers from: newmaga-parser/storage/icons/mangalib/project_<id>.(jpg|png|webp)
Writes to Postgres: novels + novel_localizations + chapters + chapter_contents

Usage example:
  python3 import_mangalib_stub.py \
    --sqlite /home/fsociety/newmaga-parser/newmanga_local.db \
    --covers /home/fsociety/newmaga-parser/storage/icons/mangalib \
    --uploads /home/fsociety/novels/uploads \
    --pg "postgres://novels:novels_dev_password@localhost:5432/novels?sslmode=disable" \
    --limit 20
"""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import sqlite3
import sys
import uuid
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Iterable, Optional


def slugify(value: str) -> str:
    value = value.strip().lower()
    value = value.replace("—", "-").replace("–", "-")
    value = re.sub(r"[^a-z0-9\u0400-\u04ff\s_-]+", "", value, flags=re.IGNORECASE)
    value = re.sub(r"[\s_]+", "-", value).strip("-")
    value = re.sub(r"-{2,}", "-", value)
    return value


@dataclass(frozen=True)
class Project:
    mangalib_project_id: int
    slug_url: Optional[str]
    slug: Optional[str]
    rus_name: Optional[str]
    name: Optional[str]
    summary: Optional[str]
    other_names: Optional[str]


@dataclass(frozen=True)
class Chapter:
    chapter_id: int
    volume: str
    number: str
    title: Optional[str]


def first_non_empty(*vals: Optional[str]) -> str:
    for v in vals:
        if v is None:
            continue
        s = v.strip()
        if s:
            return s
    return ""


def parse_other_names(other_names_json: Optional[str]) -> list[str]:
    if not other_names_json:
        return []
    try:
        data = json.loads(other_names_json)
        if not isinstance(data, list):
            return []
        out: list[str] = []
        seen = set()
        for x in data:
            if not isinstance(x, str):
                continue
            s = x.strip()
            if not s or s in seen:
                continue
            seen.add(s)
            out.append(s)
        return out
    except Exception:
        return []


def find_cover_path(covers_dir: str, mangalib_project_id: int) -> Optional[str]:
    base = f"project_{mangalib_project_id}"
    for ext in (".jpg", ".jpeg", ".png", ".webp"):
        p = os.path.join(covers_dir, base + ext)
        if os.path.isfile(p):
            return p
    # fallback: any extension
    try:
        for name in os.listdir(covers_dir):
            if name.startswith(base + "."):
                p = os.path.join(covers_dir, name)
                if os.path.isfile(p):
                    return p
    except FileNotFoundError:
        return None
    return None


def ensure_dir(path: str) -> None:
    os.makedirs(path, exist_ok=True)


def now_utc() -> datetime:
    return datetime.now(timezone.utc)


def build_novel_slug(p: Project) -> str:
    base = first_non_empty(p.slug_url, p.slug, p.rus_name, p.name, str(p.mangalib_project_id))
    base = slugify(base)
    if not base:
        base = f"project-{p.mangalib_project_id}"
    s = f"mangalib-{p.mangalib_project_id}-{base}"
    return s[:255].rstrip("-")


def load_projects(conn: sqlite3.Connection, limit: int, offset: int) -> list[Project]:
    cur = conn.cursor()
    cur.execute(
        """
        SELECT
            mangalib_project_id,
            slug_url,
            slug,
            rus_name,
            name,
            summary,
            other_names
        FROM mangalib_projects
        ORDER BY COALESCE(updated_at, created_at) DESC, mangalib_project_id DESC
        LIMIT ? OFFSET ?
        """,
        (limit, offset),
    )
    out: list[Project] = []
    for row in cur.fetchall():
        out.append(
            Project(
                mangalib_project_id=int(row[0]),
                slug_url=row[1],
                slug=row[2],
                rus_name=row[3],
                name=row[4],
                summary=row[5],
                other_names=row[6],
            )
        )
    return out


def load_chapters(conn: sqlite3.Connection, mangalib_project_id: int) -> list[Chapter]:
    cur = conn.cursor()
    cur.execute(
        """
        SELECT chapter_id, volume, number, title
        FROM mangalib_chapters
        WHERE project_id = ?
        ORDER BY CAST(volume AS REAL) ASC, CAST(number AS REAL) ASC, chapter_id ASC
        """,
        (mangalib_project_id,),
    )
    out: list[Chapter] = []
    for row in cur.fetchall():
        out.append(
            Chapter(
                chapter_id=int(row[0]),
                volume=str(row[1]),
                number=str(row[2]),
                title=row[3],
            )
        )
    return out


def connect_pg(pg_dsn: str):
    try:
        import psycopg
    except Exception as e:
        raise RuntimeError(
            "psycopg is required. Install with: python3 -m pip install 'psycopg[binary]'"
        ) from e
    return psycopg.connect(pg_dsn)


def pg_slug_exists(pg_conn, slug: str) -> bool:
    with pg_conn.cursor() as cur:
        cur.execute("SELECT EXISTS(SELECT 1 FROM novels WHERE slug = %s)", (slug,))
        return bool(cur.fetchone()[0])


def copy_cover_to_uploads(
    *,
    uploads_dir: str,
    covers_dir: str,
    mangalib_project_id: int,
    novel_id: uuid.UUID,
) -> Optional[str]:
    src = find_cover_path(covers_dir, mangalib_project_id)
    if not src:
        return None
    ext = os.path.splitext(src)[1].lower()
    if ext not in (".jpg", ".jpeg", ".png", ".webp"):
        ext = ".jpg"
    dst_dir = os.path.join(uploads_dir, "covers")
    ensure_dir(dst_dir)
    dst_name = f"{novel_id}{ext}"
    dst = os.path.join(dst_dir, dst_name)
    shutil.copyfile(src, dst)
    return f"covers/{dst_name}"


def import_project(
    *,
    pg_conn,
    project: Project,
    chapters: list[Chapter],
    lang: str,
    placeholder: str,
    uploads_dir: str,
    covers_dir: str,
    dry_run: bool,
) -> tuple[bool, str]:
    title = first_non_empty(project.rus_name, project.name, project.slug_url, project.slug)
    if not title:
        return False, "no title"

    novel_slug = build_novel_slug(project)
    if pg_slug_exists(pg_conn, novel_slug):
        return False, "exists"

    if dry_run:
        return True, f"[dry-run] would import slug={novel_slug} title={title!r} chapters={len(chapters)}"

    novel_id = uuid.uuid4()
    ts = now_utc()

    alt_titles = parse_other_names(project.other_names)
    description = (project.summary or "").strip() or None

    cover_key = copy_cover_to_uploads(
        uploads_dir=uploads_dir,
        covers_dir=covers_dir,
        mangalib_project_id=project.mangalib_project_id,
        novel_id=novel_id,
    )

    with pg_conn.cursor() as cur:
        # novels
        cur.execute(
            """
            INSERT INTO novels
              (id, slug, cover_image_key, translation_status, original_chapters_count, release_year, author, created_at, updated_at)
            VALUES
              (%s, %s, %s, 'ongoing', %s, NULL, NULL, %s, %s)
            """,
            (str(novel_id), novel_slug, cover_key, len(chapters), ts, ts),
        )

        # novel_localizations
        cur.execute(
            """
            INSERT INTO novel_localizations
              (novel_id, lang, title, description, alt_titles, created_at, updated_at)
            VALUES
              (%s, %s, %s, %s, %s, %s, %s)
            """,
            (str(novel_id), lang, title, description, alt_titles, ts, ts),
        )

        # chapters + chapter_contents
        for idx, ch in enumerate(chapters, start=1):
            chapter_id = uuid.uuid4()
            number = float(idx)  # stable unique numbering
            ch_title = (ch.title or "").strip() or None

            cur.execute(
                """
                INSERT INTO chapters
                  (id, novel_id, number, slug, title, views, published_at, created_at, updated_at)
                VALUES
                  (%s, %s, %s, NULL, %s, 0, %s, %s, %s)
                """,
                (str(chapter_id), str(novel_id), number, ch_title, ts, ts, ts),
            )

            cur.execute(
                """
                INSERT INTO chapter_contents
                  (chapter_id, lang, content, word_count, source, updated_at)
                VALUES
                  (%s, %s, %s, 0, 'import', %s)
                """,
                (str(chapter_id), lang, placeholder, ts),
            )

    return True, f"imported novel_id={novel_id} slug={novel_slug} chapters={len(chapters)} cover={'yes' if cover_key else 'no'}"


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--sqlite", required=True, help="Path to newmaga-parser SQLite DB (newmanga_local.db)")
    ap.add_argument("--covers", required=True, help="Path to covers dir (storage/icons/mangalib)")
    ap.add_argument("--uploads", required=True, help="Path to novels uploads dir (will write covers/...)")
    ap.add_argument("--pg", required=True, help="Postgres DSN, e.g. postgres://user:pass@localhost:5432/db?sslmode=disable")
    ap.add_argument("--lang", default="ru")
    ap.add_argument("--limit", type=int, default=50)
    ap.add_argument("--offset", type=int, default=0)
    ap.add_argument("--placeholder", default="Скоро будет")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args(argv)

    if not os.path.isfile(args.sqlite):
        print(f"SQLite DB not found: {args.sqlite}", file=sys.stderr)
        return 2

    ensure_dir(args.uploads)

    sqlite_conn = sqlite3.connect(args.sqlite)
    sqlite_conn.row_factory = sqlite3.Row

    projects = load_projects(sqlite_conn, args.limit, args.offset)
    print(f"Found {len(projects)} projects to process (limit={args.limit} offset={args.offset})")

    pg_conn = connect_pg(args.pg)
    pg_conn.autocommit = False

    imported = 0
    skipped = 0
    try:
        for p in projects:
            chs = load_chapters(sqlite_conn, p.mangalib_project_id)
            ok, msg = import_project(
                pg_conn=pg_conn,
                project=p,
                chapters=chs,
                lang=args.lang,
                placeholder=args.placeholder,
                uploads_dir=args.uploads,
                covers_dir=args.covers,
                dry_run=args.dry_run,
            )

            if args.dry_run:
                print(msg)
                continue

            if ok:
                pg_conn.commit()
                imported += 1
                print("OK  ", msg)
            else:
                pg_conn.rollback()
                skipped += 1
                print("SKIP", f"id={p.mangalib_project_id}", msg)
    finally:
        try:
            pg_conn.close()
        except Exception:
            pass
        sqlite_conn.close()

    print(f"Done. imported={imported} skipped={skipped}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))

