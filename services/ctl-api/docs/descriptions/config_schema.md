Return jsonschemas for Nuon configs. These can be used in frontmatter in most editors that have a TOML LSP (such as
[Taplo](https://taplo.tamasfe.dev/) configured.

```toml
#:schema https://api.nuon.co/v1/general/config-schema?source=input

description = "description"
```

You can pass in a valid source argument to render within a specific source file:

- input
- installer
- sandbox
- runner
- docker_build
- container_image
- helm
- terraform
- job

By default, the config expects that you are using multiple files and sources. If you are _not_, then pass the
`?flat=true` param.
