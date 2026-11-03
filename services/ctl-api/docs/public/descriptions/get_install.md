Returns an install by name or ID when live v3 grants and both credential ceilings permit `install.read` at that install.
The core detail load shares the authorization snapshot. Embedded parent-app data contains only ID and name; input
values use the existing redacted representation. Missing and inaccessible installs return not found.
