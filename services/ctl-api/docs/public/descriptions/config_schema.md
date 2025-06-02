Return jsonschemas for Nuon configs. These can be used in frontmatter in most editors that have a TOML LSP (such as
[Taplo](https://taplo.tamasfe.dev/) configured.

```toml
#:schema https://api.nuon.co/v1/general/config-schema?source=inputs

description = "description"
```

You can pass in a valid source argument to render within a specific config file:

- input
- input-group
- installer
- sandbox
- runner
- docker_build
- container_image
- helm
- terraform
- job
