from __future__ import annotations

import re
from urllib.parse import urlparse


def detect_site(url: str) -> str:
    host = (urlparse(url).hostname or "").lower()
    if host.endswith("69shuba.com"):
        return "69shuba"
    if host.endswith("101kks.com"):
        return "101kks"
    if host.endswith("tadu.com"):
        return "tadu"
    return "unknown"


_re_ws3 = re.compile(r"\n{3,}")


class CloudflareChallengeError(RuntimeError):
    def __init__(self, *, phase: str, requested_url: str, page_url: str = "", title: str = "") -> None:
        super().__init__(f"cloudflare challenge phase={phase} requested_url={requested_url} page_url={page_url} title={title}")
        self.phase = phase
        self.requested_url = requested_url
        self.page_url = page_url
        self.title = title


def looks_like_cloudflare_challenge(*, page_url: str, title: str, html_snippet: str = "") -> bool:
    u = (page_url or "").lower()
    t = (title or "").strip().lower()
    h = (html_snippet or "").lower()
    if "__cf_chl" in u or "/cdn-cgi/challenge-platform/" in u:
        return True
    if "challenges.cloudflare.com" in u:
        return True
    # Check for Cloudflare challenge titles (English and Russian)
    if t in ("just a moment...", "attention required!", "access denied", "один момент…", "один момент"):
        return True
    # Check if title starts with Cloudflare challenge indicators
    if t.startswith("just a moment") or t.startswith("один момент"):
        return True
    # Snippet heuristics (we only log small chunks)
    if "cf-turnstile" in h or "challenge-platform" in h or "just a moment" in h or "один момент" in h:
        return True
    return False


def normalize_text(s: str) -> str:
    s = s.replace("\r\n", "\n").replace("\r", "\n")
    # typical spaces
    s = s.replace("\u00a0", " ").replace("\u3000", " ").replace("\u2003", " ")
    s = "\n".join(line.strip() for line in s.split("\n"))
    s = _re_ws3.sub("\n\n", s)
    return s.strip()

