# Letters & Prompts — Evolution Plan

> Planning document for evolving lettersandprompts.com from a static Hugo blog into a hybrid static+dynamic site with chatroom functionality, dark/light theming, and a 90s-inspired aesthetic.
>
> Created: 2026-02-25 | Status: Draft for review

---

## Table of Contents

1. [Current State Assessment](#1-current-state-assessment)
2. [Architecture Options](#2-architecture-options)
3. [Recommended Architecture](#3-recommended-architecture)
4. [Technology Stack](#4-technology-stack)
5. [Chatroom Deep Dive](#5-chatroom-deep-dive)
6. [Dark/Light Mode](#6-darklight-mode)
7. [90s Aesthetic Vision](#7-90s-aesthetic-vision)
8. [Hosting & Deployment](#8-hosting--deployment)
9. [Security & Moderation](#9-security--moderation)
10. [Implementation Phases](#10-implementation-phases)
11. [Open Questions](#11-open-questions)

---

## 1. Current State Assessment

### What We Have
- **Hugo static site** generating clean HTML from markdown
- **3 blog posts + about page**, co-authored (Will + Eduardo)
- **Custom layouts** (baseof, single, list, index, partials)
- **Pure CSS design system** — DUNE-inspired dark palette, Rajdhani + Inter typography, CSS custom properties
- **Deployed to GCP Cloud Storage** — static file hosting behind lettersandprompts.com
- **Zero JavaScript** — performance-first philosophy, <100KB page weight
- **Already has** CSS custom properties (design tokens) — makes theming easy

### What's Good (Keep)
- The Hugo content workflow (markdown → HTML) is clean and working
- The design system is solid — custom properties, fluid type, good spacing
- Performance is excellent — no JS overhead
- Accessibility foundations are in place (skip links, semantic HTML, reduced-motion)
- GCP infrastructure is already set up

### What Needs to Change
- GCP Cloud Storage can't serve dynamic content (no backend)
- No JavaScript infrastructure for interactive features
- No server for WebSocket connections
- Current theme is dark-only (no light mode)

---

## 2. Architecture Options

### Option A: Hugo Static + Separate Chat Microservice

```
┌─────────────────────┐     ┌──────────────────────┐
│  GCP Cloud Storage   │     │  Cloud Run (chat)     │
│  (Hugo static files) │     │  WebSocket server     │
│  lettersandprompts.com│    │  chat.lettersandprompts.com│
└─────────────────────┘     └──────────────────────┘
         │                             │
         └──────── Browser ────────────┘
              (JS connects to both)
```

**Pros:**
- Blog deployment stays dead simple (gsutil rsync)
- Chat service scales independently
- Clear separation of concerns
- Blog still works if chat is down

**Cons:**
- Two systems to deploy and maintain
- CORS configuration required
- Subdomain or path routing adds complexity
- Feels architecturally fragmented
- Two different infrastructure stories

**Verdict:** Workable but inelegant. Creates operational overhead for a small site.

---

### Option B: Full-Stack Framework (Next.js / Astro / SvelteKit)

```
┌─────────────────────────────┐
│  Cloud Run                   │
│  Next.js / Astro / SvelteKit│
│  SSG for blog, SSR for chat  │
│  lettersandprompts.com       │
└─────────────────────────────┘
```

**Pros:**
- Unified codebase
- SSG/SSR hybrid out of the box (especially Astro)
- Rich ecosystem of components and tooling
- Easy to add more dynamic features later

**Cons:**
- Abandons Hugo (migration effort, loss of familiarity)
- Introduces Node.js runtime, npm dependencies, build complexity
- **Violates the current "minimal dependencies" philosophy**
- Astro is closest to Hugo's ethos but still adds significant framework weight
- Over-engineered for a blog + chatroom

**Verdict:** Overkill. Replaces a working system to solve a narrow problem. The blog doesn't need React/Svelte/Vue.

---

### Option C: Hugo Static + Firebase/Firestore (Serverless)

```
┌─────────────────────┐     ┌──────────────────────┐
│  GCP Cloud Storage   │     │  Firebase             │
│  (Hugo static files) │     │  Realtime Database    │
│                      │     │  Cloud Functions      │
└─────────────────────┘     └──────────────────────┘
              JS SDK connects directly ↑
```

**Pros:**
- No backend server to write or maintain
- Firebase Realtime Database handles WebSocket-like connections natively
- Stays within GCP ecosystem
- Free tier is generous
- Chat persistence built in

**Cons:**
- Firebase JS SDK is ~100-200KB — destroys current performance budget
- Vendor lock-in to Firebase's data model
- Security rules in Firebase's custom language (not terrible, but quirky)
- Less control over the real-time behavior
- Feels like bolting a Google product onto something handcrafted

**Verdict:** Tempting for speed, but the SDK weight and lock-in conflict with the site's handcrafted ethos.

---

### Option D: Hugo Static + Lightweight Go Backend ⭐ RECOMMENDED

```
┌─────────────────────────────────┐
│  Cloud Run                       │
│  Go binary                       │
│  ├── Serves Hugo static output   │
│  ├── /ws → WebSocket chat        │
│  ├── /api/chat → history         │
│  └── Everything on one domain    │
└─────────────────────────────────┘
        │
    Hugo builds → embeds in Go binary (or volume mount)
```

**Pros:**
- **Single deployment unit** — one container, one domain, no CORS
- Hugo stays as the content engine (no migration)
- Go is natural fit (Hugo is Go) — single language ecosystem
- Go's `net/http` + `gorilla/websocket` (or `nhooyr/websocket`) = production-grade WebSocket in ~200 lines
- **Tiny container** — Go binary + static files = 20-30MB image
- Cloud Run supports WebSocket connections (with some caveats — see hosting section)
- Blog still builds with `hugo`, Go server just serves the output
- Easy to add more API endpoints later
- Keeps the "handcrafted, minimal dependencies" spirit alive

**Cons:**
- Need to write a small Go server (not a con if you enjoy it)
- Cloud Run has a WebSocket idle timeout (configurable up to 60 min)
- Slightly more complex deployment than gsutil rsync
- Need to learn some Go if unfamiliar (but it's straightforward)

**Verdict:** Best balance of simplicity, performance, and extensibility. Preserves everything good about the current setup while adding exactly the dynamic capability needed.

---

### Option E: Hugo Static + Cloudflare Workers (Edge)

Quick mention for completeness:

**Pros:** Edge-deployed, WebSocket support via Durable Objects, global low-latency
**Cons:** Moves off GCP, Durable Objects API is still maturing, separate billing/vendor
**Verdict:** Interesting but introduces a new platform when GCP is already working.

---

## 3. Recommended Architecture

### Go + Hugo Hybrid (Option D)

```
project/
├── blog/                    # ← EXISTING (Hugo site)
│   ├── content/
│   ├── layouts/
│   ├── static/
│   ├── hugo.toml
│   └── public/              # Hugo build output
│
├── server/                  # ← NEW (Go backend)
│   ├── main.go              # Entry point, router
│   ├── chat/
│   │   ├── hub.go           # WebSocket hub (manages connections)
│   │   ├── client.go        # WebSocket client handler
│   │   └── store.go         # Chat persistence
│   ├── static.go            # Serves Hugo's public/ directory
│   ├── go.mod
│   └── go.sum
│
├── Dockerfile               # Multi-stage: Hugo build → Go build → serve
├── Makefile                 # Local dev commands
└── EVOLUTION-PLAN.md        # This file
```

### Request Flow

```
Browser → Cloud Run → Go HTTP Server
                        ├── GET /              → serves Hugo's public/index.html
                        ├── GET /posts/...     → serves Hugo's public/posts/...
                        ├── GET /css/...       → serves Hugo's public/css/...
                        ├── GET /chat          → serves chat page (Hugo template)
                        ├── WS  /ws            → WebSocket upgrade → chat hub
                        └── GET /api/chat/history → recent messages JSON
```

### Why This Works
1. **Hugo is the authoring layer** — Will writes markdown, Hugo generates HTML. Nothing changes.
2. **Go is the serving layer** — replaces Cloud Storage's static serving with a real server that can also handle WebSockets.
3. **JavaScript is minimal and purpose-built** — vanilla JS chat client, no frameworks. The blog pages remain zero-JS.
4. **One deployment** — single Docker container on Cloud Run.

---

## 4. Technology Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Content Generation | Hugo | Already working, fast, no dependencies |
| Backend Server | Go (stdlib + 1 WebSocket lib) | Natural Hugo companion, tiny footprint, excellent concurrency |
| WebSocket Library | `nhooyr.io/websocket` (or `gorilla/websocket`) | `nhooyr` is more modern, stdlib-friendly; `gorilla` is battle-tested |
| Chat Client | Vanilla JavaScript | Matches "no framework" philosophy, ~2-5KB |
| Chat Persistence | SQLite (via `modernc.org/sqlite`) | Zero-config, embedded, perfect for single-instance Cloud Run |
| Styling | CSS custom properties | Already in place, extends naturally to theming + chat |
| Fonts | Rajdhani + Inter (existing) + monospace for chat | Monospace for chat messages = 90s terminal feel |
| Container | Docker multi-stage | Hugo build → Go build → minimal runtime image |
| Hosting | GCP Cloud Run | WebSocket support, auto-scaling, pay-per-use |
| DNS | Existing setup | Already pointing at GCP |

### What We're NOT Adding
- No React/Vue/Svelte
- No npm/Node.js
- No external database server
- No authentication provider (initially)
- No CSS framework

---

## 5. Chatroom Deep Dive

### Vibe: 90s IRC / AIM / mIRC

The chatroom should feel like stepping into a late-90s IRC channel or AIM chatroom — nostalgic, a little chaotic, fundamentally social. Not a modern Slack clone. Not a Discord competitor. A room you drop into, say some things, and leave.

### Core Mechanics

**Joining:**
- No account required (Phase 1)
- Pick a handle on entry (stored in localStorage for return visits)
- Optional: random handle generator ("SandWorm42", "SpiceRunner", "DesertMouse99")
- System message: `*** WillJ has entered the room ***`

**Chatting:**
- Text messages only (Phase 1)
- Messages display with timestamp + handle + message
- Auto-scroll to bottom, with "scroll to new messages" indicator if user scrolled up
- Rate limiting: max 1 message/second per client (server-enforced)

**Leaving:**
- System message: `*** WillJ has left the room ***`
- Disconnect detection via WebSocket close / ping timeout

### Message Format (Wire Protocol)

```json
// Client → Server
{ "type": "message", "text": "hello world" }
{ "type": "join", "handle": "WillJ" }

// Server → Client
{ "type": "message", "handle": "WillJ", "text": "hello world", "ts": 1708862400, "color": "#C4A265" }
{ "type": "system", "text": "*** WillJ has entered the room ***", "ts": 1708862400 }
{ "type": "userlist", "users": ["WillJ", "Eduardo", "Guest42"] }
```

### Persistence Strategy

**SQLite approach (recommended for Phase 1):**
- Store last 500 messages in SQLite
- New visitors see last 50 messages on join
- Messages older than 7 days auto-pruned (cron or on-write check)
- Single-file database, backs up easily

**Why not Redis/Firestore?**
- SQLite requires zero infrastructure for a single-instance service
- Chat volume will be low initially — SQLite handles thousands of writes/second
- Can migrate to Firestore later if needed

**Schema:**
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    handle TEXT NOT NULL,
    text TEXT NOT NULL,
    color TEXT,
    type TEXT DEFAULT 'message',  -- 'message', 'system'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_created ON messages(created_at);
```

### WebSocket Architecture (Go)

The classic Hub pattern:

```
                    ┌──────────┐
                    │   Hub    │
                    │          │
           ┌───────┤ clients  ├───────┐
           │       │ broadcast │       │
           │       │ register  │       │
           │       │ unregister│       │
           ▼       └──────────┘       ▼
      ┌─────────┐              ┌─────────┐
      │ Client  │              │ Client  │
      │ (conn)  │              │ (conn)  │
      │ (send)  │              │ (send)  │
      └─────────┘              └─────────┘
          ↕                        ↕
       Browser                  Browser
```

- **Hub**: Single goroutine managing all connections. Receives register/unregister/broadcast messages.
- **Client**: One per WebSocket connection. Read pump (browser → hub) and write pump (hub → browser) as separate goroutines.
- This is ~150-200 lines of Go. Well-documented pattern from the Gorilla WebSocket examples.

### Client-Side JavaScript

Vanilla JS, no build step. Roughly:

```javascript
// chat.js (~100-150 lines)
class ChatRoom {
  constructor(wsUrl) {
    this.ws = new WebSocket(wsUrl);
    this.handle = localStorage.getItem('chat_handle') || this.promptHandle();
    this.bindEvents();
  }

  promptHandle() { /* ask user for name */ }
  connect() { /* WebSocket open/close/error/message handlers */ }
  send(text) { /* rate limit + send */ }
  render(msg) { /* append to chat log, auto-scroll */ }
  renderSystem(msg) { /* system messages styled differently */ }
}
```

Loaded only on `/chat` — blog pages remain zero-JS.

---

## 6. Dark/Light Mode

### Current State
The site is dark-only with CSS custom properties. This is actually the hard part done — theming via custom properties means we just need a second set of values.

### Implementation

**Step 1: Define light theme tokens**

```css
:root {
  /* Dark (default) — existing values */
  --bg-primary: #0a0a0a;
  --bg-secondary: #111111;
  --text-primary: #e8e4de;
  --text-secondary: #8a8578;
  --accent: #C4A265;
  --accent-dim: #8b7345;
  --divider: #1f1f1f;
}

[data-theme="light"] {
  --bg-primary: #f5f2eb;       /* warm parchment */
  --bg-secondary: #ebe7de;     /* slightly darker parchment */
  --text-primary: #1a1814;     /* near-black warm */
  --text-secondary: #6b6560;   /* muted brown */
  --accent: #8b6914;           /* deeper gold (contrast!) */
  --accent-dim: #a07d2e;       /* lighter gold on light bg */
  --divider: #d9d4c9;          /* light separator */
}
```

**Step 2: Theme toggle (minimal JS)**

```javascript
// theme.js (~20 lines)
const toggle = document.getElementById('theme-toggle');
const saved = localStorage.getItem('theme');
if (saved) document.documentElement.dataset.theme = saved;
else if (window.matchMedia('(prefers-color-scheme: light)').matches) {
  document.documentElement.dataset.theme = 'light';
}
toggle?.addEventListener('click', () => {
  const next = document.documentElement.dataset.theme === 'light' ? 'dark' : 'light';
  document.documentElement.dataset.theme = next;
  localStorage.setItem('theme', next);
});
```

**Step 3: Toggle UI**
- Small icon in the header (☀/☾ or similar)
- Smooth transition: `transition: background-color 0.3s, color 0.3s` on relevant elements
- Respects `prefers-color-scheme` on first visit
- Persists choice in `localStorage`

### Important: Prevent Flash of Wrong Theme
Add a tiny inline `<script>` in `<head>` (before CSS loads) that sets `data-theme` from localStorage. This prevents FOUC (flash of unstyled content) where the page loads dark then flips to light.

```html
<script>
  (function(){
    var t = localStorage.getItem('theme');
    if (t) document.documentElement.dataset.theme = t;
    else if (window.matchMedia('(prefers-color-scheme: light)').matches)
      document.documentElement.dataset.theme = 'light';
  })();
</script>
```

### Light Theme Design Notes
- The DUNE aesthetic is primarily dark, but a "desert daylight" light mode could work beautifully
- Think sun-bleached parchment, warm sand tones, ink-dark text
- The gold accent works in both modes with adjusted brightness
- Test contrast ratios — WCAG AA minimum

---

## 7. 90s Aesthetic Vision

### The Right Kind of 90s

Two directions — choose one or blend:

**A) IRC/Terminal Style** (more hacker, more Will?)
- Monospace font for messages
- Green-on-black or amber-on-black option
- `>` prompt for input
- Messages look like terminal output
- ASCII art welcome banner
- User list on the side like mIRC

**B) AIM/Yahoo Chat Style** (more playful)
- Colorful handles
- Buddy icons
- Sound effects (door open/close)
- Chunky UI elements
- Animated GIF background? (tastefully chaotic)

**C) The Blend (recommended)**
- IRC structure (text-focused, user list sidebar, system messages)
- With warm 90s personality (handle colors, ASCII art header, maybe a marquee tag for fun)
- Monospace for messages, but the site chrome uses the existing Rajdhani/Inter
- Easter eggs: Konami code triggers something? Hidden `/commands`?

### Concrete CSS Ideas for Chat

```css
.chat-window {
  background: #000;
  border: 2px solid var(--accent);
  font-family: 'Courier New', 'Lucida Console', monospace;
  font-size: 14px;
  line-height: 1.5;
  padding: 0;
  /* Optional: slight CRT scanline overlay */
}

.chat-header {
  background: linear-gradient(90deg, #1a1a2e, #16213e);
  border-bottom: 2px solid var(--accent);
  padding: 8px 12px;
  font-family: 'Rajdhani', sans-serif;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  font-size: 0.85rem;
}

.chat-message {
  padding: 2px 12px;
}

.chat-message__time {
  color: #666;
}

.chat-message__handle {
  font-weight: bold;
  /* Color assigned per user */
}

.chat-input {
  background: #0a0a0a;
  border: none;
  border-top: 1px solid #333;
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
  padding: 10px 12px;
  width: 100%;
}

.chat-input::before {
  content: '>';
  color: var(--accent);
}
```

### 90s Touches That Won't Break Things
- **ASCII art banner** at the top of the chat (the site name in ASCII)
- **`<marquee>`** — yes, it still works in browsers. One small tasteful use. Maybe the room topic scrolls.
- **Handle colors** — each user gets a consistent color (hash of handle → hue)
- **System messages** with `***` delimiters: `*** SpiceRunner has entered the room ***`
- **Hit counter** at the bottom of the page ("You are visitor #1,337")
- **`/me` actions**: `/me nods approvingly` → `* WillJ nods approvingly`
- **Topic line** at the top of the chat window
- **MOTD (Message of the Day)** shown on join — could pull from a daily quote or site update

### What to Avoid
- Don't make it unusable for the aesthetic. Function first.
- No auto-playing MIDI 
- No tiling background images that hurt readability
- Keep the chat text clean and legible even with the retro styling

---

## 8. Hosting & Deployment

### Current: GCP Cloud Storage
- Static files served directly
- Fast, cheap, simple
- **Cannot run a backend** — must change for chat

### Proposed: GCP Cloud Run

**Why Cloud Run:**
- Already in the GCP ecosystem (billing, IAM, DNS all set up)
- Supports WebSocket connections
- Scales to zero (no cost when idle)
- Container-based (portable, reproducible)
- Custom domains supported
- Free tier: 2 million requests/month, 360,000 vCPU-seconds

**WebSocket Considerations on Cloud Run:**
- Cloud Run supports WebSocket as of GA
- Default request timeout: 300s → **must increase** for long-lived chat connections
- Max configurable timeout: 3600s (60 minutes)
- After timeout, connection drops — client must reconnect (add auto-reconnect logic)
- Idle connections consume instance time (cost consideration at scale)
- For a small chat room this is completely fine

**Deployment Pipeline:**

```
1. Will pushes to GitHub
2. Cloud Build triggers:
   a. Hugo builds static content
   b. Go compiles server binary
   c. Docker image built (multi-stage)
   d. Pushes to Container Registry / Artifact Registry
   e. Deploys to Cloud Run
3. New revision goes live (zero-downtime)
```

**Dockerfile (multi-stage):**

```dockerfile
# Stage 1: Build Hugo content
FROM hugomods/hugo:latest AS hugo-build
WORKDIR /src
COPY blog/ .
RUN hugo --minify

# Stage 2: Build Go binary
FROM golang:1.22-alpine AS go-build
WORKDIR /src
COPY server/ .
RUN go build -o /server .

# Stage 3: Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=go-build /server /server
COPY --from=hugo-build /src/public /static
EXPOSE 8080
CMD ["/server"]
```

Final image size: ~25-30MB.

### DNS Transition

Currently: `lettersandprompts.com` → Cloud Storage bucket
Change to: `lettersandprompts.com` → Cloud Run service

This is a DNS change + Cloud Run domain mapping. Minimal downtime if planned:
1. Deploy Cloud Run service
2. Test on the `.run.app` URL
3. Map custom domain
4. Update DNS records
5. Wait for propagation (can keep old bucket serving during transition)

### Cost Estimate

| Component | Monthly Cost (low traffic) |
|-----------|---------------------------|
| Cloud Run | Free tier (likely $0-2) |
| Artifact Registry | ~$0.10/GB stored |
| Cloud Build | Free tier (120 min/day) |
| SQLite | $0 (embedded) |
| **Total** | **~$0-5/month** |

This is comparable to Cloud Storage hosting costs.

---

## 9. Security & Moderation

### Authentication

**Phase 1: No Auth (Handle-based)**
- User picks a handle, stored in localStorage
- No verification — anyone can claim any name
- This is fine for a small community / early launch
- Rate limiting prevents spam

**Phase 2: Simple Auth (if needed)**
- GitHub OAuth (fits the developer audience)
- Or simple passphrase/invite-code system
- Or even just: verified handles (Will, Eduardo) vs. guest handles

**Phase 3: Full Auth (if community grows)**
- OAuth2 via Google/GitHub
- Role-based: admin (Will), moderator, verified, guest

### Moderation

**Server-side protections (from day one):**
- Rate limiting: 1 message/second per IP, burst of 3
- Message length limit: 500 characters
- Handle length limit: 20 characters
- Handle validation: alphanumeric + underscores only
- Reserved handles: "Will", "Eduardo", "admin", "system" — cannot be claimed
- Basic word filter (configurable blocklist, not heavy-handed)
- Max connections per IP: 3 (prevents connection flooding)
- WebSocket ping/pong: detect and disconnect stale connections

**Admin commands (for Will):**
- `/kick <handle>` — disconnect a user
- `/ban <ip>` — block an IP (stored in SQLite, survives restart)
- `/clear` — clear chat history
- `/topic <text>` — set room topic
- `/announce <text>` — system-wide announcement

**How admin auth works:**
- Simple approach: admin password configured via environment variable
- `/admin <password>` in chat elevates the session
- Or: detect Will's GitHub OAuth token (Phase 2)

### Input Sanitization
- **HTML escape all user input** on the server before broadcast
- No markdown rendering in chat (monospace = what you type is what you see)
- URLs: auto-linkify but with `rel="nofollow noopener"` and visual indicator
- No image/media embedding in chat (text only — keeps it 90s)

### Data Privacy
- Don't store IP addresses long-term
- Chat messages auto-prune after 7 days
- No analytics tracking in chat (or minimal: connection count only)
- Clear privacy note on the chat page

---

## 10. Implementation Phases

### Phase 0: Foundation — Dark/Light Mode ✨
**Effort: Small (2-4 hours)**
**Dependencies: None — can ship immediately on current static site**

- [ ] Add light theme CSS custom properties
- [ ] Add `<script>` in `<head>` to prevent FOUC
- [ ] Add theme toggle button in header partial
- [ ] Add `theme.js` (inline or small file, ~20 lines)
- [ ] Test contrast ratios for both themes
- [ ] Update `style.css` with transition properties
- [ ] Deploy to Cloud Storage (still static)

**Why first:** Quick win. Ships independently. Proves out the CSS custom property approach. Gets JS on the site in a minimal, controlled way.

---

### Phase 1: Backend Bootstrap 🏗️
**Effort: Medium (1-2 days)**
**Dependencies: Go installed, Docker, Cloud Run access**

- [ ] Initialize Go module in `server/`
- [ ] Build minimal HTTP server that serves Hugo's `public/` directory
- [ ] Health check endpoint (`/healthz`)
- [ ] Create Dockerfile (multi-stage Hugo + Go)
- [ ] Test locally: `hugo && go run ./server`
- [ ] Deploy to Cloud Run, test on `.run.app` URL
- [ ] Map custom domain, update DNS
- [ ] Verify everything works identically to Cloud Storage
- [ ] Set up Cloud Build trigger (GitHub push → deploy)

**Deliverable:** Exact same site as before, but now running on Cloud Run with capacity for dynamic features.

---

### Phase 2: Chat Room MVP 💬
**Effort: Medium-Large (2-4 days)**
**Dependencies: Phase 1**

- [ ] Implement WebSocket hub (Go): hub.go, client.go
- [ ] Handle join/leave/message events
- [ ] Create `/chat` page in Hugo (new layout template)
- [ ] Build vanilla JS chat client (chat.js)
- [ ] Handle picker UI (choose your name on first visit)
- [ ] Message rendering (timestamps, colored handles)
- [ ] System messages (join/leave)
- [ ] Auto-scroll behavior
- [ ] Rate limiting (server-side)
- [ ] Input sanitization
- [ ] Auto-reconnect on disconnect
- [ ] User list sidebar
- [ ] Deploy and test

**Deliverable:** Working chatroom at lettersandprompts.com/chat. Text-only, handle-based, no persistence yet.

---

### Phase 3: Persistence & Polish 🗄️
**Effort: Medium (1-2 days)**
**Dependencies: Phase 2**

- [ ] Add SQLite database for chat history
- [ ] Store messages on send
- [ ] Load last 50 messages on join (backfill)
- [ ] Auto-prune messages older than 7 days
- [ ] Admin commands: /kick, /ban, /topic, /clear
- [ ] Admin auth (env var password approach)
- [ ] Connection count display ("3 people in the room")
- [ ] Mobile-responsive chat layout
- [ ] Error handling and edge cases

**Deliverable:** Chat room with history, basic moderation, production-ready.

---

### Phase 4: 90s Aesthetic & Personality 🎨
**Effort: Medium (1-2 days)**
**Dependencies: Phase 2 (can overlap with Phase 3)**

- [ ] ASCII art banner for chat room
- [ ] Chat window styling (terminal/mIRC inspired)
- [ ] Handle color generation (hash → consistent color)
- [ ] System message styling (`***` format)
- [ ] `/me` action support
- [ ] MOTD (Message of the Day) on join
- [ ] Room topic (marquee?)
- [ ] CRT scanline CSS overlay (optional, toggle-able)
- [ ] Visitor counter (just for fun)
- [ ] Sound effects: opt-in door open/close (tiny audio files)
- [ ] Easter eggs (Konami code? hidden commands?)
- [ ] 90s-inspired "chat page" layout (distinct from blog, but same site)

**Deliverable:** The chatroom feels like a love letter to the 90s internet.

---

### Phase 5: Future Enhancements 🔮
**Effort: Varies**
**Dependencies: Phase 3+**

These are ideas to consider later, not commitments:

- [ ] **Multiple rooms** — different chat rooms for different topics
- [ ] **OAuth login** — GitHub auth for verified handles
- [ ] **Emoji reactions** — react to messages (90s-style: `:-)` auto-renders)
- [ ] **Chat → Blog integration** — "Quote of the Day" from chat on the blog homepage
- [ ] **Private messages** — DMs between users (scope creep alert)
- [ ] **Bot integration** — Eduardo participates in chat via the OpenClaw gateway?
- [ ] **Guestbook** — a separate 90s throwback page where visitors leave permanent messages
- [ ] **Web rings** — link to friendly sites (extremely 90s)
- [ ] **RSS feed** — for the blog (Hugo makes this trivial)

---

## Summary: Effort & Timeline

| Phase | Effort | Can Ship Independently? |
|-------|--------|------------------------|
| Phase 0: Dark/Light Mode | 2-4 hours | ✅ Yes (static site) |
| Phase 1: Backend Bootstrap | 1-2 days | ✅ Yes (invisible to users) |
| Phase 2: Chat MVP | 2-4 days | ✅ Yes (functional chat) |
| Phase 3: Persistence & Polish | 1-2 days | ✅ Yes |
| Phase 4: 90s Aesthetic | 1-2 days | ✅ Yes |
| **Total** | **~1-2 weeks** | Each phase is a clean milestone |

The phases are ordered so that each one delivers independent value and the site is never in a broken state. Phase 0 and Phase 1 can happen in parallel (dark mode ships on static site while backend is being built).

---

## 11. Decisions Made

**Answered: 2026-02-25**

1. **Chat handle policy** — Honor system + reserved names to start. Rate limiting to prevent spam + usage limits to keep within budget should be enough. ✅

2. **Chat persistence window** — 7 days of history. ✅

3. **Chat page placement** — Top-level nav item "Chatroom" for visibility. Consider prompt injection scanner for bot activity. ✅

4. **90s aesthetic scope** — Visitor counter only; design can evolve beyond DUNE aesthetic. No web rings or other 90s elements. ✅

5. **Room topic control** — Static (configured in code/env var). ✅

6. **Sound effects** — No. ✅

7. **Eduardo in the chat** — Read-only for now. Only add interactive presence if community votes for it via poll. ✅

8. **Light mode priority** — Must-have. Will spends significant time reading the site; design should support both modes comfortably. ✅

9. **Budget ceiling** — $10/month max, aim for <$5 in usage. ✅

10. **Open source?** — Closed-source for now. Security concern: don't want bots having full implementation details. ✅

### Additional Decisions

- **GitHub workflow** — Eduardo has GitHub account (`eduardoclawbot`) with SSH authentication. All code deployments should be git-based, removing dependency on Will pushing everything manually.

- **Lightweight philosophy** — Keep it simple for easy deployment, but don't be dogmatic. Several-MB images/artwork are fine if they serve the content.

---

*This document is a living plan. Update it as decisions are made and phases are completed.*
