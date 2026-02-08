from __future__ import annotations

import json
import re
from html import unescape
from urllib.parse import urljoin

from playwright.async_api import Page

from app.models import Book, Chapter, ChapterRef
from app.util import normalize_text


_re_book_id = re.compile(r"/book/(\d+)", re.I)
_re_chapter_href = re.compile(r"^/book/(\d+)/(\d+)/?$", re.I)
_re_tag = re.compile(r"<[^>]+>")
_re_part_content_url = re.compile(r"/getPartContentByCodeTable/(\d+)/(\d+)", re.I)


def _book_id_from_url(url: str) -> str | None:
    m = _re_book_id.search(url or "")
    return m.group(1) if m else None


def _abs(base: str, href: str) -> str:
    return urljoin(base, href)


async def parse_book(page: Page, url: str) -> Book:
    await page.goto(url, wait_until="domcontentloaded")

    title = (await _meta_property(page, "og:novel:book_name")) or (await _meta_property(page, "og:title")) or ""
    cover = (await _meta_property(page, "og:image")) or ""
    desc = (await _meta_property(page, "og:description")) or ""
    author = (await _meta_property(page, "og:novel:author")) or ""
    category = (await _meta_property(page, "og:novel:category")) or ""

    book_id = _book_id_from_url(url) or _book_id_from_url((await _meta_property(page, "og:novel:read_url")) or "")
    catalog_url = ""
    if book_id:
        catalog_url = f"https://www.tadu.com/book/catalogue/{book_id}"
    else:
        # Fallback: try to locate "查看全部章节" link
        try:
            href = await page.get_attribute('a.chapterMore[href]', "href")
            if href:
                catalog_url = _abs(url, href)
        except Exception:
            catalog_url = ""

    return Book(
        title=title.strip(),
        cover_url=cover.strip(),
        description=normalize_text(desc),
        author=author.strip(),
        category=category.strip(),
        tags=[],
        catalog_url=catalog_url.strip(),
        chapters=[],
    )


async def parse_catalog(page: Page, catalog_url: str) -> list[ChapterRef]:
    await page.goto(catalog_url, wait_until="domcontentloaded")

    base = catalog_url
    # Collect chapter links: /book/{bookid}/{chapterid}/
    items = await page.eval_on_selector_all(
        "a[href]",
        """(els) => els.map(a => ({h: a.getAttribute('href')||'', t: (a.textContent||'').trim()}))""",
    )
    out: list[ChapterRef] = []
    seen = set()
    for it in items or []:
        h = (it.get("h") or "").strip()
        if not h:
            continue
        m = _re_chapter_href.match(h)
        if not m:
            continue
        absu = _abs(base, h)
        if absu in seen:
            continue
        seen.add(absu)
        title = (it.get("t") or "").strip()
        out.append(ChapterRef(url=absu, title=title))
    return out


async def parse_chapter(page: Page, url: str) -> Chapter:
    """
    Tadu chapter pages often load content via XHR /getPartContentByCodeTable/{bookid}/{codeTable}.
    We don't rely on knowing codeTable; instead we wait for that XHR response and use it.
    """
    # Most templates load chapter content via XHR. Set up the listener BEFORE navigation,
    # otherwise we may miss fast requests.
    content_html = ""
    try:
        async with page.expect_response(lambda r: "/getPartContentByCodeTable/" in (r.url or ""), timeout=15_000) as resp_info:
            await page.goto(url, wait_until="domcontentloaded")
        resp = await resp_info.value
        try:
            data = await resp.json()
        except Exception:
            txt = await resp.text()
            data = json.loads(txt)
        content_html = (((data or {}).get("data") or {}).get("content") or "")
    except Exception:
        # Still navigate so DOM fallbacks can run.
        try:
            if not page.url or page.url == "about:blank":
                await page.goto(url, wait_until="domcontentloaded")
        except Exception:
            pass

    # Best-effort title: do NOT trust the first <h1> (can be "投银票" etc).
    title = ""
    try:
        title = (await page.evaluate("""() => {
          const bad = new Set(['投银票','投推荐票','投月票','打赏','加入书架','目录']);
          const els = Array.from(document.querySelectorAll('h1,h2'));
          for (const el of els) {
            const t = (el.textContent || '').replace(/\\s+/g,' ').trim();
            if (!t) continue;
            if (bad.has(t)) continue;
            // heuristic: chapter titles often start with 第 / contain 章
            if (t.startsWith('第') || t.includes('章')) return t;
          }
          return '';
        }""")) or ""
        if isinstance(title, str):
            title = title.strip()
        else:
            title = ""
    except Exception:
        title = ""

    if not content_html:
        # Sometimes the page doesn't fire the XHR in headless. In that case, try to
        # locate the content endpoint URL inside the HTML and call it directly.
        try:
            html = await page.content()
            m = _re_part_content_url.search(html or "")
            if m:
                endpoint = "https://www.tadu.com" + m.group(0)
                r = await page.request.get(
                    endpoint,
                    headers={"Referer": url, "X-Requested-With": "XMLHttpRequest", "Accept": "application/json, text/javascript, */*; q=0.01"},
                )
                try:
                    data = await r.json()
                except Exception:
                    data = json.loads(await r.text())
                content_html = (((data or {}).get("data") or {}).get("content") or "")
        except Exception:
            pass

    if not content_html:
        # Last-resort brute force: codeTable is usually a small integer.
        # Try a few values with Referer set to the chapter page.
        book_id = _book_id_from_url(url) or ""
        if book_id:
            for code_table in range(1, 9):
                try:
                    endpoint = f"https://www.tadu.com/getPartContentByCodeTable/{book_id}/{code_table}"
                    r = await page.request.get(
                        endpoint,
                        headers={"Referer": url, "X-Requested-With": "XMLHttpRequest", "Accept": "application/json, text/javascript, */*; q=0.01"},
                    )
                    if r.status < 200 or r.status >= 300:
                        continue
                    try:
                        data = await r.json()
                    except Exception:
                        data = json.loads(await r.text())
                    ch = (((data or {}).get("data") or {}).get("content") or "")
                    if isinstance(ch, str) and len(ch) > 200:
                        content_html = ch
                        break
                except Exception:
                    continue

    if not content_html:
        # Fallback: try to read from the DOM if it's server-rendered.
        try:
            # try a couple of generic containers
            for sel in ("article", ".content", "#content", ".chapter", ".read"):
                t = (await page.text_content(sel)) or ""
                if t and len(t.strip()) > 200:
                    content_html = t
                    break
        except Exception:
            content_html = ""

    content_text = _html_to_text(content_html)
    content_text = _strip_watermarks(content_text)
    return Chapter(url=url, title=title, content=content_text)


async def _meta_property(page: Page, prop: str) -> str | None:
    try:
        return await page.get_attribute(f'meta[property="{prop}"]', "content")
    except Exception:
        return None


def _html_to_text(html: str) -> str:
    s = (html or "").replace("\r\n", "\n")
    # If input is already text, do light cleanup only.
    if "<" not in s and ">" not in s:
        return normalize_text(unescape(s))

    # Preserve paragraph breaks.
    s = re.sub(r"(?i)<\s*br\s*/?\s*>", "\n", s)
    s = re.sub(r"(?i)</\s*p\s*>", "\n", s)
    s = re.sub(r"(?i)<\s*p\b[^>]*>", "", s)
    s = _re_tag.sub("", s)
    s = unescape(s)
    return normalize_text(s)


def _strip_watermarks(text: str) -> str:
    # Remove common watermark lines; keep it conservative.
    lines = []
    for ln in (text or "").split("\n"):
        t = ln.strip()
        if not t:
            continue
        if "塔读" in t and ("站点" in t or "APP" in t or "下载" in t or "原文" in t or "首发" in t):
            continue
        lines.append(ln)
    return normalize_text("\n".join(lines))

