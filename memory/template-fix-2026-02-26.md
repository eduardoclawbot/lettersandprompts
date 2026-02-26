# Template Fix - 2026-02-26

## Problem
The chat page (`/chat/`) was rendering with the wrong template - using `_default/single.html` (about page layout) instead of the chat-specific layout.

## Root Cause
Hugo doesn't recognize `layout = "chat"` in front matter the way we tried to use it. The `layout` parameter in Hugo front matter works differently than expected.

## Solution
1. Changed front matter from `layout = "chat"` to `type = "chat"` in `content/chat.md`
2. Created `layouts/chat/` directory
3. Moved `layouts/_default/chat.html` to `layouts/chat/single.html`

This follows Hugo's convention where `type = "chat"` tells Hugo to look for templates in `layouts/chat/` directory.

## Testing
✅ Tested locally with `hugo server` on port 1314
✅ Verified correct template loads with all chat interface elements:
  - ASCII banner
  - Chat message container
  - Input box with ">" prompt
  - User sidebar
  - Handle picker modal

## Deployment
- Commit: 0db8185 "Fix chat page template resolution (type=chat, layouts/chat/single.html)"
- Pushed to GitHub main branch
- Cloud Build in progress: tag `template-fix`
- Will deploy to Cloud Run after build completes

## Hugo Template Lookup Order Reference
When Hugo sees `type = "chat"` in front matter, it looks for templates in this order:
1. `layouts/chat/single.html` ✅ (what we created)
2. `layouts/chat/single.*.html`
3. `layouts/_default/single.html`

This is the proper way to create custom layouts for specific content types in Hugo.
