# How to build link-preview (OG) pages and dynamic preview images

A hand-off guide for an implementation AI. It describes the pattern as
built and proven in this repo (`internal/shareweb/` — Tikos videos
`/t/:handle/:id`, Spaces `/s/:id`, contributions `/c/:token`, with
dynamic cards at `/og/s/:id.png` and `/og/t/:id.png`). Everything here
is portable to any stack; the Go specifics are named so you can copy
them or substitute equivalents.

## The mental model

When someone pastes a link into WhatsApp/X/Telegram/iMessage/Slack, a
**scraper bot** (not a browser) fetches the URL once and reads ONLY the
HTML `<head>` meta tags. Two consequences drive the whole design:

1. **Scrapers never run JavaScript.** A React SPA shows them an empty
   shell. The share URL must be **server-rendered HTML** with the meta
   tags already in the markup. (You do not need SSR for your whole app
   — just for the share routes.)
2. **The preview image is just another URL** in an `og:image` tag. It
   can be a static brand image, or an endpoint that renders a
   per-item PNG on demand — which is what makes previews feel alive
   ("LIVE · 1.2K listening", raised-amount progress, the video's
   poster frame).

So a share surface = two endpoints:

```
GET /x/:id          → HTML page: og/twitter meta + a human landing page
GET /og/x/:id.png   → 1200×630 PNG rendered per item, cached
```

## Part 1 — the share page

### Routes and rendering

Mount share routes at the ROOT of the domain (short URLs matter:
`tulaafrica.com/c/AbC123`). Server-render with plain HTML templates
(Go `html/template` here, embedded via `go:embed`); each page type has
its own template plus one shared `<head>` partial for fonts/CSS/store
metadata. The page must be a REAL landing page too — humans click
these links: show the item, then a "Get the app" CTA (and
`apple-itunes-app` meta so iOS shows the App Store smart banner).

### The meta tags that matter (all of them, every page)

```html
<link rel="canonical" href="{absolute page URL}">
<meta property="og:site_name" content="Tula">
<meta property="og:type"      content="website">
<meta property="og:title"     content="Support {title} on Tula">
<meta property="og:description" content="{one dense human sentence}">
<meta property="og:image"     content="{ABSOLUTE https URL to the card}">
<meta property="og:url"       content="{absolute page URL}">
<meta name="twitter:card"     content="summary_large_image">
<meta name="twitter:title"    content="…">        <!-- repeat: X reads its own tags -->
<meta name="twitter:description" content="…">
<meta name="twitter:image"    content="…">
<meta name="description"      content="…">        <!-- plain SEO fallback -->
```

Rules learned the hard way:

- **Absolute URLs only** in `og:image`/`og:url` — relative paths break
  most scrapers. Build them from a configured public base URL, never
  from the request Host header (proxies lie).
- `twitter:card: summary_large_image` is what makes X/iMessage show
  the big card instead of a thumbnail.
- **Pack the description**: it is the only body text most previews
  show. Compose it from real data — "KES 12,500 raised of KES 50,000 —
  support Umoja Savings on Tula. Monthly chama for…". One sentence,
  data first.
- The title should carry the ACTION, not just the name ("You're
  invited: {title} — a Tula Space", "Support {name} on Tula").
- Unknown/expired/revoked items must render a graceful branded 404
  page (again server-side), not a blank error.
- Private items: render a locked card with NO details and add
  `<meta name="robots" content="noindex">`.

### Privacy: one safe card per URL

The scraper fetches with **no auth and no viewer context**, and the
same preview is then shown to everyone the link reaches. Therefore the
page AND the image must expose only what the least-privileged viewer
may see. Concretely from this repo: a private Space's page/card never
carries title or host (unless it is a paid/ticketed one the host is
deliberately promoting); the contribution page exposes progress
numbers, never member names or individual contributions; bearer-token
URLs (`/c/:token`) ARE the access control — treat token mint/revoke as
a real permissioned API with audit, and make revocation kill the page
instantly.

## Part 2 — the dynamic og:image

### Why render on demand (and not pre-generate)

A card is fetched roughly once per share per platform. Rendering takes
single-digit milliseconds and produces a 100–200 KB PNG, so: render on
request, keep a small in-memory cache, set `Cache-Control`, done. No
queue, no storage, no invalidation pipeline. (`/og/s/:id.png` here:
10-minute in-memory TTL + `Cache-Control: public, max-age=600`; live
Spaces drop to 60s because the card carries the listener count; a
512-entry cap on the cache guards against id-scanning memory abuse —
on overflow just reset the map, cards are cheap.)

### Canvas rules

- **1200×630** pixels — the universal OG size (1.91:1). Design for
  legibility at 500px wide (that's how WhatsApp shows it).
- Serve `image/png` with an explicit `Content-Length`. Name the route
  with a `.png` suffix — some scrapers sniff extensions.
- **Embed the fonts in the binary** (`go:embed` + TTF parsing here —
  Plus Jakarta Sans, same family as the app, OFL-licensed). Never
  depend on system fonts on the server.

### Drawing (Go: `fogleman/gg`; equivalents: node-canvas, Sharp+SVG, Skia)

The card is drawn like a tiny poster, ~150 lines per card type:

1. **Background**: brand gradient (`gg.NewLinearGradient` dark purple
   → violet) + 2–3 huge low-alpha circles as decorative blobs.
2. **Brand row**: logo mark + wordmark, top-left, small.
3. **State pill**: rounded rect + label ("LIVE" red, "UPCOMING" amber,
   "CONTRIBUTION" green, price "KES 500" ticket pill). Pills are what
   make the card scannable in a chat list.
4. **Title**: bold, ~64px, **manually word-wrapped** to the content
   width, max 2–3 lines with "…" truncation (measure with the font
   face; never let text overflow the canvas).
5. **Context line**: host name / creator handle / raised-of-target —
   regular weight, dimmed.
6. **Faces (optional but high-impact)**: fetch avatar images
   best-effort (2s timeout, size-capped body, failures fall back to an
   initials circle in a deterministic brand color). Draw circular:
   `DrawCircle → Clip() → DrawImage → ResetClip()`, ring highlight for
   the host. Cap at ~5 faces.
7. **Photo/poster compositing** (the Tikos card): fetch the item's
   poster/thumbnail (via a presigned URL if the bucket is private),
   decode (register jpeg/png/gif/webp decoders), **cover-crop** it into
   a rect: compute the aspect-preserving crop of the source, then
   scale with a good kernel (`golang.org/x/image/draw` CatmullRom).
   Lay a dark scrim gradient over the photo edge so text stays
   readable. Poster fetch failure must degrade to the branded
   no-poster layout — never to an HTTP error.
8. Encode PNG into a buffer, cache, serve.

### The endpoint shape

```
GET /og/{type}/{id}.png
  → parse id (strip .png), load PUBLIC metadata from DB
  → apply the SAME privacy rules as the landing page
  → cache key = id + every field that appears on the card
    (state|title|price|count …) so a state change naturally makes a
    new key instead of serving a stale card
  → cache hit? serve : render → cache → serve
```

And the landing page points at it:

```go
if cfg.OGImage == "" {                       // ops can pin a static override
    page.OGImage = baseURL + "/og/s/" + id + ".png"
}
```

## Wiring checklist for a new share surface (what I did for /c/:token)

1. DB: whatever the page needs must be readable WITHOUT auth — write a
   dedicated "public view" query/method that returns ONLY safe fields.
2. Route `GET /c/:token` on the root router + template with the full
   meta set (copy an existing template; swap pills/fields).
3. Compose `og:title`/`og:description` from real numbers.
4. (Optional) `/og/c/:token.png` card — same renderer skeleton, new
   layout. Until it exists, fall back to the static brand OG image so
   previews are never image-less.
5. Graceful 404 for unknown/revoked; `noindex` where private.
6. Tests: assert the rendered HTML contains the expected meta values
   for public/private/unknown cases (this repo does golden-sample
   PNG tests for the cards too — `og_samples_test.go`).
7. Verify with real scrapers once deployed: X Card Validator,
   Facebook Sharing Debugger, and paste-into-WhatsApp. Remember
   platforms cache aggressively — the debuggers have "re-scrape"
   buttons; a changed card may take hours to refresh organically.

## Gotchas index (the things that silently break previews)

- SPA share URLs (scrapers see an empty div) → server-render.
- Relative `og:image` → absolute https URLs from config.
- Missing `twitter:*` duplicates → tiny thumbnail on X.
- Per-viewer content on the OG endpoint → leaks; one safe card per URL.
- Fonts not embedded → tofu boxes on the server render.
- Poster fetch errors 500-ing the card → always degrade to branded.
- Unbounded per-id caches → cap + cheap reset.
- Long titles overflowing the canvas → measure, wrap, truncate.
- Forgetting `Cache-Control` → every paste re-renders.
