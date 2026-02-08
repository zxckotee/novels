from __future__ import annotations

import asyncio
import os
import random
from dataclasses import dataclass
from typing import Optional

from playwright.async_api import async_playwright, Browser, BrowserContext, Page
from playwright.async_api import Response, Request


@dataclass
class BrowserSession:
    playwright: any
    browser: Browser
    context: BrowserContext
    page: Page


async def open_page(
    *,
    user_agent: Optional[str],
    storage_state_path: Optional[str],
    cookie_header: Optional[str] = None,
    humanize: bool = False,
    locale: Optional[str] = None,
    timezone_id: Optional[str] = None,
    viewport_width: int = 1280,
    viewport_height: int = 720,
    debug_http: bool = False,
    debug_http_body_max_chars: int = 800,
) -> BrowserSession:
    pw = await async_playwright().start()
    launch_args = [
        # Hide the most obvious automation flag (not a silver bullet).
        "--disable-blink-features=AutomationControlled",
    ]
    browser = await pw.chromium.launch(headless=True, args=launch_args)

    context_kwargs = {}
    ua = (user_agent or "").strip()
    if ua:
        context_kwargs["user_agent"] = ua

    ch = (cookie_header or "").strip()
    if ch:
        # Useful for Cloudflare clearance cookies when storage_state is missing/outdated.
        context_kwargs["extra_http_headers"] = {"cookie": ch}

    if humanize:
        # More realistic defaults.
        context_kwargs["viewport"] = {"width": viewport_width, "height": viewport_height}
        # Using UA locale headers alone is not enough; locale/timezone help reduce obvious mismatches.
        if locale and locale.strip():
            context_kwargs["locale"] = locale.strip()
        if timezone_id and timezone_id.strip():
            context_kwargs["timezone_id"] = timezone_id.strip()

    if storage_state_path:
        # only load if exists; otherwise run with a fresh context
        if os.path.exists(storage_state_path):
            context_kwargs["storage_state"] = storage_state_path

    context = await browser.new_context(**context_kwargs)
    page = await context.new_page()
    if humanize:
        await _apply_stealth(context, page)
    # Keep timeouts bounded but not too aggressive; can be overridden per-request.
    page.set_default_navigation_timeout(120_000)
    page.set_default_timeout(120_000)
    if debug_http:
        _attach_http_logging(page, body_max_chars=debug_http_body_max_chars)
    return BrowserSession(playwright=pw, browser=browser, context=context, page=page)


async def close_session(sess: BrowserSession, *, save_storage_state_to: Optional[str]) -> None:
    try:
        if save_storage_state_to:
            await sess.context.storage_state(path=save_storage_state_to)
    finally:
        try:
            await sess.context.close()
        finally:
            await sess.browser.close()
            try:
                await sess.playwright.stop()
            except Exception:
                pass


def _attach_http_logging(page: Page, *, body_max_chars: int) -> None:
    # Print to stdout so docker logs show it immediately.
    def on_request(req: Request) -> None:
        try:
            rt = req.resource_type
        except Exception:
            rt = "?"
        print(f"[pw] --> {req.method} {req.url} rt={rt}", flush=True)

    async def _log_response(resp: Response) -> None:
        try:
            req = resp.request
            rt = req.resource_type
        except Exception:
            rt = "?"
        try:
            ct = (resp.headers or {}).get("content-type", "")
        except Exception:
            ct = ""
        print(f"[pw] <-- {resp.status} {resp.url} rt={rt} ct={ct}", flush=True)

        # For errors, include a small snippet of body to quickly see Cloudflare/blocks.
        if body_max_chars and resp.status >= 400:
            try:
                # Avoid logging binary responses
                if ct and ("text/" in ct or "json" in ct or "html" in ct or "xml" in ct):
                    txt = await resp.text()
                    txt = (txt or "").replace("\r\n", "\n")
                    snippet = txt[:body_max_chars]
                    print(f"[pw] body[{resp.status}] {resp.url}\n{snippet}\n---", flush=True)
            except Exception:
                pass

    def on_response(resp: Response) -> None:
        asyncio.create_task(_log_response(resp))

    def on_nav(frame) -> None:
        try:
            if frame == page.main_frame:
                print(f"[pw] nav -> {frame.url}", flush=True)
        except Exception:
            pass

    page.on("request", on_request)
    page.on("response", on_response)
    page.on("framenavigated", on_nav)


async def _apply_stealth(context: BrowserContext, page: Page) -> None:
    """
    Light-weight stealth patches to reduce the most trivial automation fingerprints.
    Not guaranteed to bypass Cloudflare/Turnstile.
    """
    # Do not advertise "HeadlessChrome" in UA.
    try:
        ua = await page.evaluate("() => navigator.userAgent")
        if isinstance(ua, str) and "HeadlessChrome" in ua:
            await context.set_extra_http_headers({"user-agent": ua.replace("HeadlessChrome", "Chrome")})
    except Exception:
        pass

    # Patch navigator.webdriver and a few common surface-level signals.
    stealth_js = r"""
(() => {
  try {
    Object.defineProperty(navigator, 'webdriver', {get: () => undefined});
  } catch (e) {}
  try {
    window.chrome = window.chrome || { runtime: {} };
  } catch (e) {}
  try {
    Object.defineProperty(navigator, 'languages', {get: () => navigator.languages && navigator.languages.length ? navigator.languages : ['en-US','en']});
  } catch (e) {}
  try {
    Object.defineProperty(navigator, 'plugins', {get: () => navigator.plugins && navigator.plugins.length ? navigator.plugins : [1,2,3,4,5]});
  } catch (e) {}
})();
"""
    try:
        await context.add_init_script(stealth_js)
    except Exception:
        pass

    # Tiny jitter to avoid "robotic" immediacy for first paint.
    try:
        await page.wait_for_timeout(random.randint(120, 380))
    except Exception:
        pass

