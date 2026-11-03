Returns app IDs and names in the selected org, filtered by live v3 grants and both credential ceilings before search
and pagination. An app is visible through `app.read` or as the parent of a readable install. Parent navigation does
not grant `app.read`; use the app detail endpoint for authorized app data. Results are ordered case-insensitively by
name, then ID; search matches a literal case-insensitive substring of name or ID.

This endpoint no longer returns configs or other full app fields. Clients must adopt the ID/name contract before cutover.
