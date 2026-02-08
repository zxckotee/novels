from __future__ import annotations

import random
from urllib.parse import urljoin

from playwright.async_api import Page

from app.models import Book, Chapter, ChapterRef
from app.util import normalize_text, CloudflareChallengeError, looks_like_cloudflare_challenge


async def _human_pause(page: Page, *, human_delay_ms_min: int, human_delay_ms_max: int) -> None:
    lo = max(0, int(human_delay_ms_min))
    hi = max(lo, int(human_delay_ms_max))
    try:
        await page.wait_for_timeout(random.randint(lo, hi))
    except Exception:
        pass


async def _human_scroll(page: Page) -> None:
    try:
        await page.evaluate(
            """() => {
              const h = Math.max(document.body.scrollHeight, document.documentElement.scrollHeight);
              const y = Math.min(h, 600 + Math.floor(Math.random()*800));
              window.scrollTo(0, y);
            }"""
        )
    except Exception:
        pass


async def _ensure_not_cloudflare(
    page: Page,
    *,
    phase: str,
    requested_url: str,
    cloudflare_wait_ms: int = 0,
) -> None:
    # Give Cloudflare a brief chance to auto-pass (some setups are JS-only).
    page_url = ""
    title = ""
    try:
        page_url = page.url or ""
    except Exception:
        page_url = ""
    try:
        title = (await page.title()) or ""
    except Exception:
        title = ""

    if looks_like_cloudflare_challenge(page_url=page_url, title=title):
        # Optionally wait a bit for auto-pass (JS challenges) before failing.
        if cloudflare_wait_ms and cloudflare_wait_ms > 0:
            try:
                await page.wait_for_timeout(int(cloudflare_wait_ms))
            except Exception:
                pass
            try:
                page_url = page.url or ""
            except Exception:
                page_url = page_url
            try:
                title = (await page.title()) or ""
            except Exception:
                title = title
            if not looks_like_cloudflare_challenge(page_url=page_url, title=title):
                return
        raise CloudflareChallengeError(phase=phase, requested_url=requested_url, page_url=page_url, title=title)

    # Quick DOM probes without pulling full HTML.
    try:
        if await page.locator("text=Just a moment").count() > 0:
            raise CloudflareChallengeError(phase=phase, requested_url=requested_url, page_url=page_url, title=title)
        if await page.locator("text=Один момент").count() > 0:
            raise CloudflareChallengeError(phase=phase, requested_url=requested_url, page_url=page_url, title=title)
        if await page.locator("iframe[src*='turnstile']").count() > 0:
            raise CloudflareChallengeError(phase=phase, requested_url=requested_url, page_url=page_url, title=title)
        if await page.locator("form#challenge-form").count() > 0:
            raise CloudflareChallengeError(phase=phase, requested_url=requested_url, page_url=page_url, title=title)
    except CloudflareChallengeError:
        raise
    except Exception:
        # If selectors fail, fall back to content snippet (best-effort).
        try:
            html = await page.content()
            snip = (html or "")[:2000]
            if looks_like_cloudflare_challenge(page_url=page_url, title=title, html_snippet=snip):
                raise CloudflareChallengeError(phase=phase, requested_url=requested_url, page_url=page_url, title=title)
        except CloudflareChallengeError:
            raise
        except Exception:
            pass


async def parse_book(
    page: Page,
    url: str,
    *,
    referer: str | None = None,
    humanize: bool = False,
    human_delay_ms_min: int = 150,
    human_delay_ms_max: int = 650,
    cloudflare_wait_ms: int = 0,
) -> Book:
    print(f"[parser-service] parse_book: navigating to {url}", flush=True)
    goto_kwargs = {"wait_until": "domcontentloaded"}
    if referer:
        goto_kwargs["referer"] = referer
    await page.goto(url, **goto_kwargs)
    print(f"[parser-service] parse_book: page loaded, url={page.url}", flush=True)
    if humanize:
        await _human_pause(page, human_delay_ms_min=human_delay_ms_min, human_delay_ms_max=human_delay_ms_max)
        await _human_scroll(page)
    await _ensure_not_cloudflare(page, phase="book", requested_url=url, cloudflare_wait_ms=cloudflare_wait_ms)

    title = await _meta_property(page, "og:title") or ""
    cover = await _meta_property(page, "og:image") or ""
    desc = await _meta_property(page, "og:description") or ""
    # Author: sometimes not present in og:* tags; prefer DOM "作者：" line.
    author = await _meta_property(page, "og:novel:author") or ""
    if not author.strip():
        try:
            author_dom = await page.evaluate(
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
                  return '';
                }"""
            )
            if isinstance(author_dom, str) and author_dom.strip():
                author = author_dom
        except Exception:
            pass
    category = await _meta_property(page, "og:novel:category") or ""
    catalog_url = await _meta_property(page, "og:novel:read_url") or ""

    tags = []
    try:
        raw = await page.evaluate("() => (typeof bookinfo !== 'undefined' && bookinfo && bookinfo.tags) ? bookinfo.tags : ''")
        if isinstance(raw, str):
            tags = [t.strip() for t in raw.split(",") if t.strip()]
    except Exception:
        tags = []

    return Book(
        title=title.strip(),
        cover_url=cover.strip(),
        description=normalize_text(desc),
        author=author.strip(),
        category=category.strip(),
        tags=_dedup(tags),
        catalog_url=catalog_url.strip(),
        chapters=[],
    )


async def parse_catalog(
    page: Page,
    catalog_url: str,
    *,
    humanize: bool = False,
    human_delay_ms_min: int = 150,
    human_delay_ms_max: int = 650,
    cloudflare_wait_ms: int = 0,
) -> list[ChapterRef]:
    print(f"[parser-service] parse_catalog: navigating to {catalog_url}", flush=True)
    await page.goto(catalog_url, wait_until="domcontentloaded")
    if humanize:
        await _human_pause(page, human_delay_ms_min=human_delay_ms_min, human_delay_ms_max=human_delay_ms_max)
        await _human_scroll(page)
    await _ensure_not_cloudflare(page, phase="catalog", requested_url=catalog_url, cloudflare_wait_ms=cloudflare_wait_ms)

    # Wait for chapter list to load (may be dynamic)
    print(f"[parser-service] parse_catalog: waiting for #tab_chapters selector", flush=True)
    try:
        await page.wait_for_selector("#tab_chapters li a[href*='/txt/']", timeout=10000)
        print(f"[parser-service] parse_catalog: found #tab_chapters selector", flush=True)
    except Exception as e:
        print(f"[parser-service] parse_catalog: selector timeout, waiting 2s fallback. err={e}", flush=True)
        # Fallback: wait a bit for any dynamic content
        await page.wait_for_timeout(2000)

    # Collect all /txt/{book}/{chapter}.html links with titles.
    # Try multiple selectors to catch different page structures.
    print(f"[parser-service] parse_catalog: extracting chapter links", flush=True)
    hrefs_with_titles = await page.evaluate(
        """() => {
          const links = [];
          // Try #tab_chapters first (most common)
          const tabChapters = document.querySelector('#tab_chapters');
          if (tabChapters) {
            tabChapters.querySelectorAll('li a[href*="/txt/"]').forEach(a => {
              const href = a.getAttribute('href') || '';
              if (href.includes('/txt/') && href.endsWith('.html')) {
                const span = a.querySelector('span');
                const title = span ? (span.textContent || '').trim() : (a.textContent || '').trim();
                links.push({ href, title });
              }
            });
          }
          // Fallback: search all links
          if (links.length === 0) {
            document.querySelectorAll('a[href*="/txt/"]').forEach(a => {
              const href = a.getAttribute('href') || '';
              if (href.includes('/txt/') && href.endsWith('.html')) {
                const title = (a.textContent || '').trim();
                links.push({ href, title });
              }
            });
          }
          return links;
        }"""
    )
    
    print(f"[parser-service] parse_catalog: found {len(hrefs_with_titles) if hrefs_with_titles else 0} links from evaluate", flush=True)
    
    out: list[ChapterRef] = []
    seen = set()
    for item in hrefs_with_titles or []:
        if not isinstance(item, dict):
            continue
        h = item.get("href") or ""
        if not h or not isinstance(h, str):
            continue
        absu = urljoin(catalog_url, h)
        if absu in seen:
            continue
        seen.add(absu)
        title = (item.get("title") or "").strip() if isinstance(item.get("title"), str) else ""
        out.append(ChapterRef(url=absu, title=title))
    
    print(f"[parser-service] parse_catalog: returning {len(out)} unique chapter refs", flush=True)
    if len(out) == 0:
        # Debug: log page URL and title to see what we got
        try:
            page_url = page.url
            page_title = await page.title()
            print(f"[parser-service] parse_catalog: WARNING - no chapters found! page_url={page_url} page_title={page_title}", flush=True)
            # Log a snippet of the page HTML to debug
            html_snippet = await page.evaluate("() => document.querySelector('#tab_chapters')?.innerHTML?.substring(0, 500) || 'NOT FOUND'")
            print(f"[parser-service] parse_catalog: #tab_chapters HTML snippet: {html_snippet}", flush=True)
        except Exception as e:
            print(f"[parser-service] parse_catalog: failed to get debug info: {e}", flush=True)
    
    return out


async def parse_chapter(
    page: Page,
    url: str,
    *,
    humanize: bool = False,
    human_delay_ms_min: int = 150,
    human_delay_ms_max: int = 650,
    cloudflare_wait_ms: int = 0,
) -> Chapter:
    await page.goto(url, wait_until="domcontentloaded")
    if humanize:
        await _human_pause(page, human_delay_ms_min=human_delay_ms_min, human_delay_ms_max=human_delay_ms_max)
        await _human_scroll(page)
    await _ensure_not_cloudflare(page, phase="chapter", requested_url=url, cloudflare_wait_ms=cloudflare_wait_ms)

    title = ""
    try:
        title = (await page.text_content("div.txtnav h1")) or ""
    except Exception:
        title = ""

    # Extract readable text from #txtcontent, removing obvious ads/scripts.
    content = await page.evaluate(
        """() => {
          const nav = document.querySelector('div.txtnav');
          if (!nav) return '';
          const root = nav.cloneNode(true);
          for (const sel of ['script', '.txtad', '.txtcenter', '.page1', '.tools', '.setbox', '#pageheadermenu', '#pagefootermenu']) {
            root.querySelectorAll(sel).forEach(n => n.remove());
          }
          const tc = root.querySelector('#txtcontent') || root;
          return (tc.innerText || '').trim();
        }"""
    )
    return Chapter(url=url, title=(title or "").strip(), content=normalize_text(str(content or "")))


async def _meta_property(page: Page, prop: str) -> str | None:
    try:
        return await page.get_attribute(f'meta[property="{prop}"]', "content")
    except Exception:
        return None


def _dedup(items: list[str]) -> list[str]:
    seen = set()
    out = []
    for x in items:
        if x not in seen:
            seen.add(x)
            out.append(x)
    return out

