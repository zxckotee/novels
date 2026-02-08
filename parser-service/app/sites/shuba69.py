from __future__ import annotations

import re
from urllib.parse import urljoin

from playwright.async_api import Page

from app.models import Book, Chapter, ChapterRef
from app.util import normalize_text


_re_catalog_href = re.compile(r"/book/\d+/all\.html$", re.I)


async def parse_book(page: Page, url: str) -> Book:
    await page.goto(url, wait_until="domcontentloaded")

    # Title: first h1 is usually the book title.
    title = (await _safe_text(page, "h1")) or ""

    # Cover image container differs across templates: bookimg2 vs bookimg.
    cover = (
        (await _safe_attr(page, ".bookimg2 img", "src"))
        or (await _safe_attr(page, ".bookimg img", "src"))
        or ""
    )
    desc = await _safe_text(page, ".jianjie") or ""
    author = ""
    try:
        author = await page.evaluate(
            """() => {
              const root = document.querySelector('div.booknav2') || document;
              const ps = Array.from(root.querySelectorAll('p'));
              for (const p of ps) {
                const t = (p.textContent || '').replace(/\\s+/g,' ').trim();
                if (t.startsWith('作者：') || t.startsWith('作者:')) {
                  const a = p.querySelector('a');
                  return (a ? (a.textContent || '') : t.replace(/^作者[:：]/,'')).trim();
                }
              }
              // Some pages use "作 者" spacing; handle loosely.
              const m = (document.body.innerText || '').match(/作者[:：]\\s*([^\\n\\r]+)/);
              return m ? (m[1] || '').trim() : '';
            }"""
        )
        author = (author or "").strip()
    except Exception:
        author = ""

    # Find "完整目录" / all.html link
    catalog_url = ""
    try:
        hrefs = await page.eval_on_selector_all(
            "a[href]",
            "(els) => els.map(a => ({h: a.getAttribute('href')||'', t:(a.textContent||'').trim()}))",
        )
        for it in hrefs or []:
            h = (it.get("h") or "").strip()
            t = (it.get("t") or "").strip()
            if not h:
                continue
            absu = urljoin(url, h)
            if "完整目录" in t or _re_catalog_href.search(absu):
                catalog_url = absu
                break
    except Exception:
        catalog_url = ""

    return Book(
        title=title.strip(),
        cover_url=cover.strip(),
        description=normalize_text(desc),
        author=author,
        category="",
        tags=[],
        catalog_url=catalog_url,
        chapters=[],
    )


async def parse_catalog(page: Page, catalog_url: str) -> list[ChapterRef]:
    await page.goto(catalog_url, wait_until="domcontentloaded")
    # 69shuba catalog usually has links under #catalog or similar.
    hrefs = await page.eval_on_selector_all(
        "a[href]",
        """(els) => els
          .map(a => ({h: a.getAttribute('href')||'', t:(a.textContent||'').trim(), num: a.getAttribute('data-num')||''}))
          .filter(x => x.h && x.h.includes('/txt/') && x.h.endsWith('.html'))""",
    )
    out: list[ChapterRef] = []
    seen = set()
    for it in hrefs or []:
        h = (it.get("h") or "").strip()
        if not h:
            continue
        absu = urljoin(catalog_url, h)
        if absu in seen:
            continue
        seen.add(absu)
        title = (it.get("t") or "").strip()
        num_raw = (it.get("num") or "").strip()
        num = int(num_raw) if num_raw.isdigit() else None
        out.append(ChapterRef(url=absu, title=title, number=num))
    return out


async def parse_chapter(page: Page, url: str) -> Chapter:
    await page.goto(url, wait_until="domcontentloaded")

    title = (await _safe_text(page, "div.txtnav h1")) or ""

    content = await page.evaluate(
        """() => {
          const nav = document.querySelector('div.txtnav');
          if (!nav) return '';
          const root = nav.cloneNode(true);
          for (const sel of ['script', 'style', '.txtinfo', '.txtad', '.txtcenter', '.page1', '#txtright']) {
            root.querySelectorAll(sel).forEach(n => n.remove());
          }
          return (root.innerText || '').trim();
        }"""
    )
    return Chapter(url=url, title=title.strip(), content=normalize_text(str(content or "")))


async def _safe_text(page: Page, sel: str) -> str | None:
    try:
        return await page.text_content(sel)
    except Exception:
        return None


async def _safe_attr(page: Page, sel: str, attr: str) -> str | None:
    try:
        return await page.get_attribute(sel, attr)
    except Exception:
        return None

