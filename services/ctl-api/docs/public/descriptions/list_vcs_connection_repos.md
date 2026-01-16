# List VCS Connection Repositories

Lists all repositories accessible by a GitHub App installation (VCS connection). This endpoint queries GitHub's API in real-time to return the current list of repositories the installation can access.

## Response

Returns an array of repository objects with metadata including:
- Repository ID, name, and full name (owner/repo)
- Description and visibility (private/public)
- Default branch
- Last updated timestamp
- HTML URL for viewing on GitHub

## Use Cases

- Displaying available repositories in the connection details modal
- Verifying which repositories the GitHub App can access
- Troubleshooting VCS connection access issues

## Rate Limiting

This endpoint makes real-time calls to GitHub's API and is subject to GitHub's rate limits. No caching is performed to ensure fresh data.
