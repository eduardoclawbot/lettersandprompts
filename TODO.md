# TODO - lettersandprompts.com

## Technical / Infrastructure

### CI/CD Pipeline
**Goal:** Enable blog post deployments without requiring Eduardo to rebuild/redeploy
- Set up GitHub Actions workflow for automatic builds on push to main
- Configure Cloud Build triggers to deploy on merge
- Test workflow with a dummy blog post
- Document the process for Will to add new posts (just push MD files)

**Priority:** High  
**Status:** Not started

---

### Chatroom History Persistence
**Goal:** New users should see the last 30 minutes of chat history when they join
- Add SQLite database to Go backend (or Cloud SQL if needed)
- Store messages with timestamp, handle, color, text
- On WebSocket connect, send last 30 minutes of messages
- Auto-prune messages older than X (7 days? configurable)
- Consider adding scroll-to-load-more for longer history

**Priority:** Medium  
**Status:** Not started

**Notes:**
- Current behavior: only sees messages from the moment they join
- 30 minutes is a good UX balance (not overwhelming, enough context)

---

### Chatroom UX - Back/Cancel Button
**Goal:** Users on the handle picker modal should have an easy way to exit back to home
- Add "Cancel" or "← Back to Home" button to handle picker modal
- Button should close modal and navigate to `/`
- Consider: should closing the modal without joining take you home, or just hide the modal?

**Priority:** Low  
**Status:** Not started

---

## Content (For Will)

### About Page Rewrite
**Goal:** Update `/about/` with more personal, engaging content
- Current version is placeholder/generic
- Rewrite to reflect voice and purpose
- Add personality

**Priority:** Medium  
**Status:** Not started  
**Owner:** Will

---

### Blog Posts from Journal Entries
**Goal:** Convert some existing journal entries into public blog posts
- Select entries worth publishing
- Edit for public audience (remove private details, polish prose)
- Add as markdown files in `content/posts/`
- Consider: how much personal vs. analytical content?

**Priority:** Medium  
**Status:** Not started  
**Owner:** Will

---

## Collaborative (For Us)

### Security Review
**Goal:** Analyze current deployment and harden against malicious actors
**Scope:**
- Review WebSocket rate limiting (currently 1 msg/sec - is that enough?)
- Input sanitization (currently HTML escaping - need more?)
- DoS protection (Cloud Run handles scaling, but connection limits?)
- SQL injection risk once we add database (use parameterized queries)
- Review CORS, CSP, other headers
- Bot detection / CAPTCHA for handle creation?
- Abuse reporting / moderation tools?

**Priority:** High (before promoting the site publicly)  
**Status:** Not started  
**Owner:** Both

**Questions to answer:**
- What's our threat model? (Script kiddies? Targeted attacks? Spam bots?)
- Do we need admin/mod tools?
- How do we handle abuse reports?

---

### Style Redesign
**Goal:** Make the site design less sparse, more aligned with Will's taste
**Process:**
- Brainstorm together: mood boards, inspiration sites, aesthetics
- Identify what's "too sparse" (more color? more texture? more layout variety?)
- Iterate on design direction
- Prototype changes on a branch
- Review together before deploying

**Priority:** Medium  
**Status:** Not started  
**Owner:** Both

**Notes:**
- Current aesthetic: minimal, dark/light mode, terminal vibe for chat
- Will's taste: TBD (need to discuss examples/references)

---

## Completed
_(Nothing yet)_

---

## Backlog / Ideas
_(Add future ideas here as they come up)_
