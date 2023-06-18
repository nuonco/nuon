# Plugins

## catalog

This package exposes an interface for fetching information about a plugin.

```go
pc, err := catalog.New(c.v, catalog.WithCredentials(&credentials.Config{
  UseDefault: true,
}))

plugin, err := pc.GetLatest(ctx, catalog.PluginTypeTerraform)
if err != nil {
  return fmt.Errorf("unable to get latest: %w", err)
}
```

## configs

This package contains configs for plugins, as well as shared tooling for building plugins.

### Migration Path

When we originally set out, it was faster/easier to debug the plugin tooling by just rendering go templates. This was always a short term solution, and we will be migrating the existing plugin configs to this package.
