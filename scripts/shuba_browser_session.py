import argparse
import json
import os
import pathlib
import sys
from datetime import datetime, timezone

from playwright.sync_api import sync_playwright


def _now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat()

def _cookie_summary(cookies: list[dict]) -> str:
    # return a short one-line summary (names only) for debugging
    names = []
    for c in cookies:
        n = (c.get("name") or "").strip()
        if n:
            names.append(n)
    names = sorted(set(names))
    return ", ".join(names)


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Interactive browser session (Playwright). "
            "Use it to manually authenticate/complete checks in a real browser, "
            "then export cookies/storage state for later HTTP scraping."
        )
    )
    parser.add_argument("--url", required=True, help="Start URL to open (e.g. https://www.69shuba.com/book/90474.htm)")
    parser.add_argument(
        "--out",
        default="/data/shuba_storage.json",
        help="Where to save Playwright storage state (cookies + localStorage)",
    )
    parser.add_argument(
        "--user-data-dir",
        default="/data/pw-user-data",
        help="Persistent browser profile dir (keeps cache/storage across runs)",
    )
    parser.add_argument(
        "--chromium-channel",
        default="chromium",
        help="Playwright channel: chromium|chrome|msedge (default: chromium)",
    )
    parser.add_argument("--headed", action="store_true", help="Run headed (needed for interactive checks)")
    parser.add_argument(
        "--wait",
        default="manual",
        choices=["manual", "networkidle"],
        help="How to wait before prompting to export state",
    )

    args = parser.parse_args()

    out_path = pathlib.Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    user_data_dir = pathlib.Path(args.user_data_dir)
    user_data_dir.mkdir(parents=True, exist_ok=True)

    with sync_playwright() as p:
        context = p.chromium.launch_persistent_context(
            user_data_dir=str(user_data_dir),
            headless=not args.headed,
            viewport={"width": 1200, "height": 800},
            locale="en-US",
            user_agent=os.environ.get(
                "PLAYWRIGHT_UA",
                "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
            ),
        )
        page = context.new_page()

        print(f"[{_now_iso()}] Opening: {args.url}", flush=True)
        page.goto(args.url, wait_until="domcontentloaded")

        if args.wait == "networkidle":
            try:
                page.wait_for_load_state("networkidle", timeout=60_000)
            except Exception:
                # Networkidle can be flaky on JS-heavy pages; ignore.
                pass

        print(
            "\nBrowser is running.\n"
            "- Complete any interactive steps in the visible browser (if --headed).\n"
            "- When you're done, press ENTER here to export storage state.\n",
            flush=True,
        )
        try:
            input()
        except KeyboardInterrupt:
            print("Interrupted; not saving.", flush=True)
            context.close()
            return 130

        storage = context.storage_state()
        out_path.write_text(json.dumps(storage, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"[{_now_iso()}] Saved storage state: {out_path}", flush=True)

        # Debug hint: whether Cloudflare clearance cookie exists (if a challenge was shown).
        cookies = storage.get("cookies", []) if isinstance(storage, dict) else []
        has_clearance = any((c.get("name") == "cf_clearance") for c in cookies if isinstance(c, dict))
        print(f"[{_now_iso()}] Cookies saved ({len(cookies)}): {_cookie_summary(cookies)}", flush=True)
        if has_clearance:
            print(f"[{_now_iso()}] OK: cf_clearance is present.", flush=True)
        else:
            print(
                f"[{_now_iso()}] NOTE: cf_clearance is NOT present. "
                "If import still gets 403/Just a moment, repeat and ensure the page fully loads after any check.",
                flush=True,
            )

        context.close()
        return 0


if __name__ == "__main__":
    raise SystemExit(main())

