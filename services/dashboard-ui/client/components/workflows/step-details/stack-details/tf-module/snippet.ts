// Terraform snippet fragments shared by every published stack module's "TF Module"
// tab. Nothing here is cloud-specific: the module's `inputs` and `secrets` maps and
// the root-level `variable` blocks have the same shape whatever the target.

const padTo = (name: string, width: number) => name.padEnd(width, ' ')

// `inputs = { ... }` for the module block. Only keys the app declares are accepted,
// so the snippet lists exactly the customer-facing ones.
export const buildInputsBlock = (
  inputs: Array<{ name?: string; default?: string }>
): string => {
  if (inputs.length === 0) return ''

  const nameWidth = Math.max(
    ...inputs.map((input) => (input.name ?? '').length)
  )
  const lines = inputs.map(
    (input) =>
      `    ${padTo(input.name ?? '', nameWidth)} = "${input.default ?? ''}"`
  )

  return `\n\n  inputs = {\n${lines.join('\n')}\n  }`
}

// `secrets = { ... }`. Auto-generated secrets are the stack's, so only
// customer-supplied ones belong here. Values come from the root `variable` blocks.
export const buildSecretsBlock = (
  secrets: Array<{ name?: string }>
): string => {
  if (secrets.length === 0) return ''

  const width = Math.max(...secrets.map((secret) => (secret.name ?? '').length))
  const lines = secrets.map(
    (secret) =>
      `    ${padTo(secret.name ?? '', width)} = { value = var.${secret.name ?? ''} }`
  )

  return `\n\n  secrets = {\n${lines.join('\n')}\n  }`
}

// Root-level `variable` blocks for each customer secret. Terraform fills them from
// `TF_VAR_<name>`, so no real value is ever written to main.tf.
export const buildSecretVariablesBlock = (
  secrets: Array<{ name?: string; description?: string }>
): string => {
  if (secrets.length === 0) return ''

  const blocks = secrets.map((secret) => {
    const description = secret.description?.trim()
    // Widths match what `terraform fmt` would produce for the attributes present.
    const width = description ? 'description'.length : 'sensitive'.length
    const attrs = [
      `  ${padTo('type', width)} = string`,
      `  ${padTo('sensitive', width)} = true`,
      ...(description
        ? [`  ${padTo('description', width)} = "${description}"`]
        : []),
    ]
    return `variable "${secret.name ?? ''}" {\n${attrs.join('\n')}\n}`
  })

  return `\n\n${blocks.join('\n\n')}`
}

// `export TF_VAR_<name>='<placeholder>'` lines, shown alongside both auth methods
// since secrets are needed regardless of how the module authenticates.
export const buildSecretExports = (secrets: Array<{ name?: string }>): string =>
  secrets
    .map(
      (secret) =>
        `export TF_VAR_${secret.name ?? ''}='<${(secret.name ?? '').replace(/_/g, '-')}-value>'`
    )
    .join('\n')
