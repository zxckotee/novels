from __future__ import annotations

from typing import Any, Literal, Optional

from pydantic import BaseModel, Field


Site = Literal["69shuba", "101kks", "tadu", "unknown"]


class ParseRequest(BaseModel):
    url: str = Field(..., description="Book page URL")
    site: Optional[Site] = Field(None, description="Optional explicit site override")
    chapters_limit: int = Field(0, ge=0, description="0 = no limit")
    user_agent: Optional[str] = None
    referer: Optional[str] = Field(None, description="Optional referer for initial navigation")
    cookie_header: Optional[str] = Field(
        None,
        description="Optional Cookie header string (e.g. 'cf_clearance=...; zh_choose=t') to help bypass Cloudflare.",
    )
    storage_state_path: Optional[str] = Field(
        None,
        description="Path inside parser-service container, e.g. /data/101kks_storage.json",
    )
    save_storage_state: bool = Field(
        True, description="If true, persist updated storage_state back to storage_state_path"
    )
    navigation_timeout_ms: int = Field(
        120_000, ge=5_000, le=600_000, description="Playwright navigation timeout in ms"
    )
    debug_http: bool = Field(
        False,
        description="If true, log Playwright request/response events (status/url/resourceType).",
    )
    debug_http_body_max_chars: int = Field(
        800,
        ge=0,
        le=10_000,
        description="Max chars of response body to log for error responses (>=400). 0 disables body logging.",
    )
    humanize: bool = Field(
        False,
        description="If true, apply light stealth patches + realistic context (viewport/locale/timezone) and small random delays.",
    )
    locale: Optional[str] = Field(
        None,
        description="Optional browser locale, e.g. 'ru-RU'. Used when humanize=true.",
    )
    timezone_id: Optional[str] = Field(
        None,
        description="Optional browser timezone id, e.g. 'Europe/Moscow'. Used when humanize=true.",
    )
    viewport_width: int = Field(
        1280, ge=320, le=4096, description="Viewport width. Used when humanize=true."
    )
    viewport_height: int = Field(
        720, ge=320, le=4096, description="Viewport height. Used when humanize=true."
    )
    human_delay_ms_min: int = Field(
        150,
        ge=0,
        le=10_000,
        description="Min delay between actions (ms) when humanize=true.",
    )
    human_delay_ms_max: int = Field(
        650,
        ge=0,
        le=10_000,
        description="Max delay between actions (ms) when humanize=true.",
    )
    cloudflare_wait_ms: int = Field(
        0,
        ge=0,
        le=120_000,
        description="If >0, when Cloudflare challenge is detected, wait up to this many ms for it to resolve before failing.",
    )


class ChapterRef(BaseModel):
    url: str
    title: str = ""
    number: Optional[int] = None


class Book(BaseModel):
    title: str
    cover_url: str = ""
    description: str = ""
    author: str = ""
    category: str = ""
    tags: list[str] = Field(default_factory=list)
    catalog_url: str = ""
    chapters: list[ChapterRef] = Field(default_factory=list)


class Chapter(BaseModel):
    url: str
    title: str = ""
    content: str


class ParseResponse(BaseModel):
    site: Site
    book: Book
    chapters: list[Chapter]
    debug: dict[str, Any] = Field(default_factory=dict)

