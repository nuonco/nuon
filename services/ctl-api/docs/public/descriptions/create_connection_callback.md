Create a VCS connection after the GitHub App install redirect. The caller must be authenticated and have write access to `org_id`.

`code` is the user-authorization code GitHub issues when the installing user completes the app's OAuth web flow. The API exchanges it and only attaches `github_install_id` if that user can see the installation; otherwise it responds `403`. The installation id on its own is not trusted because GitHub lets it be spoofed.
