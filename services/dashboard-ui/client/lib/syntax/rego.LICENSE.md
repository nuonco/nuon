# rego.tmLanguage.json

Converted from `syntaxes/Rego.tmLanguage` in
[open-policy-agent/vscode-opa](https://github.com/open-policy-agent/vscode-opa),
which ships it as an XML property list. Converted to the JSON shape Shiki
expects with Python's `plistlib`; the only edit is `name`, lowercased to `rego`
so it matches the language id we register.

Licensed under the Apache License 2.0, the licence of the source repository.
Shiki does not bundle a Rego grammar, which is why this is vendored.
