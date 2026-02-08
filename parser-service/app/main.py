from __future__ import annotations

import logging
import os
from urllib.parse import urlparse

from fastapi import FastAPI, HTTPException
from playwright.async_api import TimeoutError as PlaywrightTimeoutError

from app.models import ParseRequest, ParseResponse
from app.playwright_runner import close_session, open_page
from app.util import detect_site, CloudflareChallengeError
from app.sites import kks101, shuba69, tadu

app = FastAPI(title="novels parser-service", version="0.1.0")
log = logging.getLogger("parser-service")


def _default_storage_path(site: str, url: str) -> str | None:
    storage_dir = os.environ.get("STORAGE_DIR", "/data")
    host = (urlparse(url).hostname or "").lower()
    if not host:
        return None
    if site == "unknown":
        return os.path.join(storage_dir, f"{host}_storage.json")
    return os.path.join(storage_dir, f"{site}_storage.json")


@app.get("/health")
def health() -> dict:
    return {"ok": True}


@app.post("/parse", response_model=ParseResponse)
async def parse(req: ParseRequest) -> ParseResponse:
    site = req.site or detect_site(req.url)
    if site not in ("69shuba", "101kks", "tadu"):
        raise HTTPException(status_code=400, detail=f"unsupported site: {site}")

    storage_path = req.storage_state_path or _default_storage_path(site, req.url)
    # Print progress to stdout so `docker-compose logs -f` shows activity immediately.
    print(f"[parser-service] start site={site} url={req.url} limit={req.chapters_limit} storage={storage_path}", flush=True)
    if req.cookie_header:
        print(f"[parser-service] using cookie_header (length={len(req.cookie_header)})", flush=True)
    else:
        print(f"[parser-service] no cookie_header provided", flush=True)

    print(f"[parser-service] opening browser session...", flush=True)
    sess = await open_page(
        user_agent=req.user_agent,
        storage_state_path=storage_path,
        cookie_header=req.cookie_header,
        humanize=req.humanize,
        locale=req.locale,
        timezone_id=req.timezone_id,
        viewport_width=req.viewport_width,
        viewport_height=req.viewport_height,
        debug_http=req.debug_http,
        debug_http_body_max_chars=req.debug_http_body_max_chars,
    )
    try:
        # per-request navigation timeout
        try:
            sess.page.set_default_navigation_timeout(req.navigation_timeout_ms)
            sess.page.set_default_timeout(req.navigation_timeout_ms)
        except Exception:
            pass

        print(f"[parser-service] browser session ready, starting parse for {site}", flush=True)

        if site == "101kks":
            print(f"[parser-service] parsing book page: {req.url}", flush=True)
            try:
                book = await kks101.parse_book(
                    sess.page,
                    req.url,
                    referer=(req.referer or "").strip() or None,
                    humanize=req.humanize,
                    human_delay_ms_min=req.human_delay_ms_min,
                    human_delay_ms_max=req.human_delay_ms_max,
                    cloudflare_wait_ms=req.cloudflare_wait_ms,
                )
            except CloudflareChallengeError as e:
                print(f"[parser-service] blocked phase={e.phase} url={e.requested_url} page_url={e.page_url} title={e.title}", flush=True)
                raise HTTPException(
                    status_code=403,
                    detail=(
                        "blocked by Cloudflare (Turnstile) on 101kks; "
                        "update /data/101kks_storage.json (cf_clearance) or pass cookie_header"
                    ),
                ) from e
            except PlaywrightTimeoutError as e:
                print(f"[parser-service] timeout phase=book url={req.url} err={e}", flush=True)
                raise HTTPException(status_code=504, detail=f"timeout loading book page url={req.url}: {e}") from e
            if not book.catalog_url:
                raise HTTPException(status_code=422, detail="catalog_url not found on book page")
            try:
                refs = await kks101.parse_catalog(
                    sess.page,
                    book.catalog_url,
                    humanize=req.humanize,
                    human_delay_ms_min=req.human_delay_ms_min,
                    human_delay_ms_max=req.human_delay_ms_max,
                    cloudflare_wait_ms=req.cloudflare_wait_ms,
                )
                print(f"[parser-service] catalog parsed: found {len(refs)} chapter refs from {book.catalog_url}", flush=True)
            except CloudflareChallengeError as e:
                print(f"[parser-service] blocked phase={e.phase} url={e.requested_url} page_url={e.page_url} title={e.title}", flush=True)
                raise HTTPException(
                    status_code=403,
                    detail=(
                        "blocked by Cloudflare (Turnstile) on 101kks; "
                        "update /data/101kks_storage.json (cf_clearance) or pass cookie_header"
                    ),
                ) from e
            except PlaywrightTimeoutError as e:
                print(f"[parser-service] timeout phase=catalog url={book.catalog_url} err={e}", flush=True)
                raise HTTPException(status_code=504, detail=f"timeout loading catalog page url={book.catalog_url}: {e}") from e
            if req.chapters_limit and req.chapters_limit > 0:
                refs = refs[: req.chapters_limit]
                print(f"[parser-service] limited chapters to {len(refs)} (limit={req.chapters_limit})", flush=True)
            book.chapters = refs
            print(f"[parser-service] processing {len(refs)} chapters", flush=True)
            chapters = []
            for i, ref in enumerate(refs, start=1):
                print(f"[parser-service] 101kks chapter {i}/{len(refs)} url={ref.url}", flush=True)
                try:
                    chapters.append(
                        await kks101.parse_chapter(
                            sess.page,
                            ref.url,
                            humanize=req.humanize,
                            human_delay_ms_min=req.human_delay_ms_min,
                            human_delay_ms_max=req.human_delay_ms_max,
                            cloudflare_wait_ms=req.cloudflare_wait_ms,
                        )
                    )
                except CloudflareChallengeError as e:
                    print(f"[parser-service] blocked phase={e.phase} url={e.requested_url} page_url={e.page_url} title={e.title}", flush=True)
                    raise HTTPException(
                        status_code=403,
                        detail=(
                            "blocked by Cloudflare (Turnstile) on 101kks; "
                            "update /data/101kks_storage.json (cf_clearance) or pass cookie_header"
                        ),
                    ) from e
                except PlaywrightTimeoutError as e:
                    raise HTTPException(status_code=504, detail=f"timeout loading chapter {i}/{len(refs)}: {ref.url}: {e}") from e
            return ParseResponse(site="101kks", book=book, chapters=chapters, debug={"storage_state_path": storage_path})

        if site == "69shuba":
            try:
                book = await shuba69.parse_book(sess.page, req.url)
            except PlaywrightTimeoutError as e:
                print(f"[parser-service] timeout phase=book url={req.url} err={e}", flush=True)
                raise HTTPException(status_code=504, detail=f"timeout loading book page url={req.url}: {e}") from e
            if not book.catalog_url:
                raise HTTPException(status_code=422, detail="catalog_url not found on book page")
            try:
                refs = await shuba69.parse_catalog(sess.page, book.catalog_url)
            except PlaywrightTimeoutError as e:
                print(f"[parser-service] timeout phase=catalog url={book.catalog_url} err={e}", flush=True)
                raise HTTPException(status_code=504, detail=f"timeout loading catalog page url={book.catalog_url}: {e}") from e
            if req.chapters_limit and req.chapters_limit > 0:
                refs = refs[: req.chapters_limit]
            book.chapters = refs
            chapters = []
            for i, ref in enumerate(refs, start=1):
                print(f"[parser-service] 69shuba chapter {i}/{len(refs)} url={ref.url}", flush=True)
                try:
                    chapters.append(await shuba69.parse_chapter(sess.page, ref.url))
                except PlaywrightTimeoutError as e:
                    raise HTTPException(status_code=504, detail=f"timeout loading chapter {i}/{len(refs)}: {ref.url}: {e}") from e
            return ParseResponse(site="69shuba", book=book, chapters=chapters, debug={"storage_state_path": storage_path})

        if site == "tadu":
            try:
                book = await tadu.parse_book(sess.page, req.url)
            except PlaywrightTimeoutError as e:
                print(f"[parser-service] timeout phase=book url={req.url} err={e}", flush=True)
                raise HTTPException(status_code=504, detail=f"timeout loading book page url={req.url}: {e}") from e
            if not book.catalog_url:
                raise HTTPException(status_code=422, detail="catalog_url not found on book page")
            try:
                refs = await tadu.parse_catalog(sess.page, book.catalog_url)
            except PlaywrightTimeoutError as e:
                print(f"[parser-service] timeout phase=catalog url={book.catalog_url} err={e}", flush=True)
                raise HTTPException(status_code=504, detail=f"timeout loading catalog page url={book.catalog_url}: {e}") from e
            if req.chapters_limit and req.chapters_limit > 0:
                refs = refs[: req.chapters_limit]
            book.chapters = refs
            chapters = []
            for i, ref in enumerate(refs, start=1):
                print(f"[parser-service] tadu chapter {i}/{len(refs)} url={ref.url}", flush=True)
                try:
                    chapters.append(await tadu.parse_chapter(sess.page, ref.url))
                except PlaywrightTimeoutError as e:
                    raise HTTPException(status_code=504, detail=f"timeout loading chapter {i}/{len(refs)}: {ref.url}: {e}") from e
            return ParseResponse(site="tadu", book=book, chapters=chapters, debug={"storage_state_path": storage_path})

        raise HTTPException(status_code=400, detail=f"unsupported site: {site}")
    finally:
        await close_session(sess, save_storage_state_to=(storage_path if req.save_storage_state else None))

