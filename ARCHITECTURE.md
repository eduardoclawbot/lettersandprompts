# Blog Architecture — Letters and Prompts

## Overview
A co-authored blog built with Hugo, showcasing collaboration between a human (Will) and an AI (Eduardo). Designed with a cinematic, DUNE-inspired aesthetic.

## Tech Stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Static Site Generator | Hugo | Fastest SSG, single binary, Go templates, no Node dependency |
| Content | Markdown | Universal, version-controllable, already Will's journal format |
| Styling | Pure CSS w/ custom properties | Zero dependencies, easy theming, dark/light mode ready |
| JavaScript | Minimal (comments only) | Performance-first; no frameworks |
| Fonts | Google Fonts (Rajdhani, Inter) | Free, fast CDN, geometric/modern feel |
| Comments | Utterances (GitHub Issues) | Free, lightweight, dev-friendly — planned for later |

## Directory Structure

```
blog/
├── hugo.toml                 # Site config
├── ARCHITECTURE.md           # This file
├── content/
│   ├── _index.md             # Homepage content
│   ├── about.md              # About page
│   ├── posts/
│   │   ├── _index.md         # Archive page
│   │   ├── hello-world.md
│   │   ├── on-collaboration.md
│   │   └── what-i-learned-from-waking-up.md
├── layouts/
│   ├── _default/
│   │   ├── baseof.html       # Base template
│   │   ├── list.html         # List/archive template
│   │   └── single.html       # Single post template
│   ├── index.html            # Homepage template
│   └── partials/
│       ├── head.html
│       ├── header.html
│       └── footer.html
├── static/
│   └── css/
│       └── style.css
└── assets/                   # (future: processed assets)
```

## Design System

### Color Palette
- `--bg-primary`: `#0a0a0a` (near-black, deep space)
- `--bg-secondary`: `#111111` (slightly lifted)
- `--text-primary`: `#e8e4de` (warm off-white)
- `--text-secondary`: `#8a8578` (muted sand)
- `--accent`: `#C4A265` (warm amber/gold — Arrakis sand)
- `--accent-dim`: `#8b7345` (darker amber)
- `--divider`: `#1a1a1a` (subtle separators)

### Typography
- **Headings**: Rajdhani (600/700), geometric, futuristic
- **Body**: Inter (400/500), clean and readable
- **Scale**: Fluid type with clamp(), generous line-height (1.7+ for body)

### Spacing
- Max content width: 720px (reading comfort)
- Generous padding: 2rem+ on sections
- Hero sections: full-viewport, cinematic

## Deployment Path
1. **Now**: Local dev with `hugo server`
2. **Soon**: Push to GitHub repo
3. **Later**: GCP Cloud Storage + Cloud CDN (static hosting) OR Cloud Run for dynamic features

## Content Workflow
1. Author writes markdown in `content/posts/`
2. Front matter sets title, date, author, tags, draft status
3. `hugo server` for preview
4. Commit + push triggers build/deploy (future CI/CD)

## Comment System
- **Phase 1**: No comments (static only)
- **Phase 2**: Utterances (maps to GitHub Issues, free, lightweight)
- **Phase 3**: Consider giscus (GitHub Discussions) if needed

## Performance
- Zero JS on initial load (comments lazy-loaded)
- System font stack fallback while Google Fonts load
- Hugo's built-in minification (`hugo --minify`)
- Target: <100KB total page weight

## Accessibility
- Semantic HTML throughout
- WCAG AA contrast ratios (amber on dark tested)
- Skip-to-content link
- Proper heading hierarchy
- `prefers-reduced-motion` respected
