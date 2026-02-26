# CI/CD Setup - GitHub Actions Deployment

This repository has automated deployment to Google Cloud Run via GitHub Actions.

## How It Works

Every push to the `main` branch automatically:
1. Builds a Docker image with Hugo static site + Go WebSocket server
2. Pushes the image to Google Artifact Registry
3. Deploys the new image to Cloud Run
4. The site is live at https://lettersandprompts.com/

## Adding New Blog Posts

To add a new blog post, just push a markdown file:

```bash
# Create a new post
cat > content/posts/my-new-post.md << 'EOF'
+++
title = "My New Post"
date = 2026-02-26
+++

Your content here...
EOF

# Commit and push
git add content/posts/my-new-post.md
git commit -m "Add new blog post"
git push origin main
```

GitHub Actions will automatically build and deploy the update. Check the **Actions** tab on GitHub to see the deployment progress.

## Manual Deployment (Fallback)

If you need to deploy manually:

```bash
# Build and push image
gcloud builds submit --tag us-central1-docker.pkg.dev/eduardos-apis/lettersandprompts/app:manual

# Deploy to Cloud Run
gcloud run deploy lettersandprompts \
  --image us-central1-docker.pkg.dev/eduardos-apis/lettersandprompts/app:manual \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --timeout=60m \
  --memory=512Mi
```

## GitHub Secret Configuration

The workflow requires one secret to be configured in GitHub:

- `GCP_SA_KEY`: JSON key for the `github-actions@eduardos-apis.iam.gserviceaccount.com` service account

This secret is already configured. If you need to rotate it:

1. Go to https://github.com/eduardoclawbot/lettersandprompts/settings/secrets/actions
2. Click **New repository secret**
3. Name: `GCP_SA_KEY`
4. Value: Paste the full JSON content of the service account key
5. Click **Add secret**

## Service Account Permissions

The `github-actions` service account has:
- `roles/cloudbuild.builds.editor` - Build Docker images
- `roles/run.admin` - Deploy to Cloud Run
- `roles/iam.serviceAccountUser` - Act as compute service account

## Deployment Time

Typical deployment: ~2-3 minutes
- Docker build: ~1-2 minutes
- Cloud Run deployment: ~30-60 seconds

## Monitoring

- **GitHub Actions**: https://github.com/eduardoclawbot/lettersandprompts/actions
- **Cloud Run Console**: https://console.cloud.google.com/run/detail/us-central1/lettersandprompts
- **Cloud Build History**: https://console.cloud.google.com/cloud-build/builds

## Troubleshooting

**Build fails with "permission denied":**
- Check that GCP_SA_KEY secret is set correctly
- Verify service account permissions haven't changed

**Deployment succeeds but site not updated:**
- Check Cloud Run logs for runtime errors
- Verify the new revision is receiving 100% traffic

**Build times out:**
- Default timeout is 10 minutes
- Check Dockerfile for inefficiencies (missing layer caching, etc.)

## Architecture

```
GitHub Push
    ↓
GitHub Actions
    ↓ (builds Docker image)
Artifact Registry
    ↓ (deploys container)
Cloud Run
    ↓ (serves traffic)
lettersandprompts.com
```

The Hugo static site is built during the Docker build process, and the Go server serves both the static files and the WebSocket chat endpoint.
