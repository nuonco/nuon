Returns an app by name or ID only when live v3 grants and both credential ceilings permit `app.read` at that app.
Authorization and detail loading share a read-only snapshot. Discovering a parent app through install-scoped access
does not authorize this endpoint. Missing and inaccessible apps return not found.
