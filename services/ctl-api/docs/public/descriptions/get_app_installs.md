Returns only readable installs for the requested app in the selected org. Live v3 grants and both credential ceilings
must permit `install.read` for each returned install; `app.read` is not required. Filtering precedes search and
pagination. Embedded parent-app data contains only ID and name, not app configs or settings.
