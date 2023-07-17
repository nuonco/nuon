package external

//go:generate buf generate
//go:generate buf generate --template buf.gen.tag.yaml
//go:generate buf generate --include-imports --template buf.gen.node.yaml
//go:generate buf generate --include-imports --template buf.gen.local.yaml
