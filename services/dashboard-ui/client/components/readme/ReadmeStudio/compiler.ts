import type {
  TArraySource,
  TBlock,
  TCondition,
  TStateVariable,
  TValueKind,
} from './types'

const identifier = /^[A-Za-z_][A-Za-z0-9_-]*$/

const digExpr = (path: string, fallback: string, root: string) =>
  `(dig ${path
    .split('.')
    .map((part) => `"${part}"`)
    .join(' ')} ${fallback} ${root})`

const conditionExpr = (condition: TCondition) => {
  const value = digExpr(condition.path, '""', '.nuon')
  switch (condition.op) {
    case 'exists':
      return value
    case 'not-exists':
      return `not ${value}`
    case 'eq':
      return `eq (printf "%v" ${value}) "${condition.value ?? ''}"`
    case 'ne':
      return `ne (printf "%v" ${value}) "${condition.value ?? ''}"`
  }
}

const entityTags = {
  runbook: 'nuon-run-runbook',
  action: 'nuon-action-card',
  component: 'nuon-component-card',
} as const

const entityElement = (
  type: keyof typeof entityTags,
  id: string,
  name: string
) => {
  const tag = entityTags[type]
  const attrs = [
    id ? `id="${escapeHtml(id)}"` : '',
    name ? `name="${escapeHtml(name)}"` : '',
  ]
    .filter(Boolean)
    .join(' ')
  return `<${tag}${attrs ? ` ${attrs}` : ''}></${tag}>`
}

const valueElement = (kind: TValueKind, path: string, root: string) => {
  switch (kind) {
    case 'status':
      return `<nuon-status status="{{ ${digExpr(path, '""', root)} }}" variant="badge"></nuon-status>`
    case 'time':
      return `<nuon-time time="{{ ${digExpr(path, '""', root)} }}" format="relative"></nuon-time>`
    case 'text':
      return `{{ ${digExpr(path, '"—"', root)} }}`
  }
}

export function compileBlock(block: TBlock, index: number): string {
  switch (block.type) {
    case 'markdown':
      return block.content.trim()
    case 'raw':
      return block.content.trim()
    case 'runbook':
    case 'action':
    case 'component':
      if (!block.id && !block.name) return ''
      return entityElement(block.type, block.id, block.name)
    case 'banner': {
      const banner = [
        `<nuon-banner theme="${block.theme}">`,
        block.content.trim(),
        '</nuon-banner>',
      ].join('\n')
      if (!block.condition?.path.trim()) return banner
      return [
        `{{ if ${conditionExpr(block.condition)} }}`,
        banner,
        '{{ end }}',
      ].join('\n')
    }
    case 'status-row': {
      const items = block.items
        .filter((item) => item.path.trim())
        .map(
          (item) =>
            `<span><strong>${item.label ? `${item.label}: ` : ''}</strong>${valueElement(item.kind, item.path, '.nuon')}</span>`
        )
      return [
        '<div style="display:flex;align-items:center;gap:0.5rem 1.5rem;flex-wrap:wrap;">',
        ...items,
        '</div>',
      ].join('')
    }
    case 'table': {
      const rows = `$rows${index}`
      const row = `$row${index}`
      const lines = [
        `{{ ${rows} := (default (list) ${digExpr(block.sourcePath || 'unset', '(list)', '.nuon')}) }}`,
      ]
      if (block.limit && block.limit > 0) {
        lines.push(
          `{{ if gt (len ${rows}) ${block.limit} }}{{ ${rows} = slice ${rows} 0 ${block.limit} }}{{ end }}`
        )
      }
      lines.push(
        `{{ if eq (len ${rows}) 0 }}`,
        `_${block.emptyText || 'No data yet'}_`,
        '{{ else }}',
        '<table>',
        '  <thead>',
        `    <tr>${block.columns.map((column) => `<th>${column.header}</th>`).join('')}</tr>`,
        '  </thead>',
        '  <tbody>',
        `  {{ range ${row} := ${rows} }}`,
        `    <tr>${block.columns.map((column) => `<td>${valueElement(column.kind, column.path || 'unset', row)}</td>`).join('')}</tr>`,
        '  {{ end }}',
        '  </tbody>',
        '</table>',
        '{{ end }}'
      )
      return lines.join('\n')
    }
  }
}

export function compileTemplate(blocks: TBlock[]): string {
  const sections = blocks
    .map((block, index) => compileBlock(block, index))
    .filter((section) => section.trim())
  return `${['{{/* Generated with the Nuon README studio */}}', ...sections].join('\n\n')}\n`
}

export function resolvePath(state: unknown, path: string): unknown {
  let value: unknown = state
  for (const part of path.split('.')) {
    if (!value || typeof value !== 'object' || !(part in value))
      return undefined
    value = (value as Record<string, unknown>)[part]
  }
  return value
}

const escapeHtml = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (character) =>
      ({
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&#34;',
        "'": '&#39;',
      })[character] ?? character
  )

const scalar = (value: unknown) =>
  value !== null && ['string', 'number', 'boolean'].includes(typeof value)

export function substituteVariables(
  markdown: string,
  state?: Record<string, unknown>
) {
  if (!state) return markdown
  return markdown.replace(
    /{{\s*\.nuon\.([A-Za-z_][A-Za-z0-9_-]*(?:\.[A-Za-z_][A-Za-z0-9_-]*)*)\s*}}/g,
    (template, path: string) => {
      const value = resolvePath(state, path)
      return scalar(value) ? escapeHtml(String(value)) : template
    }
  )
}

function evalCondition(
  condition: TCondition,
  state?: Record<string, unknown>
): boolean {
  if (!condition.path.trim()) return true
  const value = state ? resolvePath(state, condition.path) : undefined
  switch (condition.op) {
    case 'exists':
      return !!value
    case 'not-exists':
      return !value
    case 'eq':
      return String(value ?? '') === (condition.value ?? '')
    case 'ne':
      return String(value ?? '') !== (condition.value ?? '')
  }
}

const previewValue = (kind: TValueKind, value: unknown): string => {
  switch (kind) {
    case 'status':
      return scalar(value)
        ? `<nuon-status status="${escapeHtml(String(value))}" variant="badge"></nuon-status>`
        : '—'
    case 'time':
      return scalar(value)
        ? `<nuon-time time="${escapeHtml(String(value))}" format="relative"></nuon-time>`
        : '—'
    case 'text':
      return scalar(value) ? escapeHtml(String(value)) : '—'
  }
}

const entityPreviewLabels = {
  runbook: {
    title: 'Run runbook',
    hint: 'Opens the runbook on the install page',
  },
  action: { title: 'Run action', hint: 'Triggers the action on the install' },
  component: {
    title: 'Deploy component',
    hint: 'Deploys the component to the install',
  },
} as const

const entityPreview = (
  type: keyof typeof entityPreviewLabels,
  name: string
) => {
  const { title, hint } = entityPreviewLabels[type]
  if (!name)
    return `<div style="border:1px dashed rgba(128,128,128,0.5);border-radius:0.5rem;padding:0.75rem 1rem;"><em>${title}: select a ${type} to finish this block</em></div>`
  return [
    '<div style="border:1px solid rgba(128,128,128,0.35);border-radius:0.5rem;padding:0.75rem 1rem;display:flex;align-items:center;gap:0.75rem;">',
    `<span style="font-size:1.1rem;">▶</span>`,
    `<span><strong>${title}: ${escapeHtml(name)}</strong><br/><span style="opacity:0.7;font-size:0.85em;">${hint}</span></span>`,
    '</div>',
  ].join('')
}

export function previewBlock(
  block: TBlock,
  state?: Record<string, unknown>
): string {
  switch (block.type) {
    case 'markdown':
      return substituteVariables(block.content.trim(), state)
    case 'raw':
      return block.content.trim()
    case 'runbook':
    case 'action':
    case 'component':
      return entityPreview(block.type, block.name)
    case 'banner': {
      if (block.condition && !evalCondition(block.condition, state)) return ''
      return [
        `<nuon-banner theme="${block.theme}">`,
        substituteVariables(block.content.trim(), state),
        '</nuon-banner>',
      ].join('\n')
    }
    case 'status-row': {
      const items = block.items
        .filter((item) => item.path.trim())
        .map((item) => {
          const value = state ? resolvePath(state, item.path) : undefined
          return `<span><strong>${item.label ? `${item.label}: ` : ''}</strong>${previewValue(item.kind, value)}</span>`
        })
      return [
        '<div style="display:flex;align-items:center;gap:0.5rem 1.5rem;flex-wrap:wrap;">',
        ...items,
        '</div>',
      ].join('')
    }
    case 'table': {
      const source = state ? resolvePath(state, block.sourcePath) : undefined
      let rows = Array.isArray(source) ? source : []
      if (block.limit && block.limit > 0) rows = rows.slice(0, block.limit)
      if (!rows.length) return `_${block.emptyText || 'No data yet'}_`
      const body = rows
        .map(
          (row) =>
            `    <tr>${block.columns
              .map(
                (column) =>
                  `<td>${previewValue(column.kind, resolvePath(row, column.path || ''))}</td>`
              )
              .join('')}</tr>`
        )
        .join('\n')
      return [
        '<table>',
        '  <thead>',
        `    <tr>${block.columns.map((column) => `<th>${column.header}</th>`).join('')}</tr>`,
        '  </thead>',
        '  <tbody>',
        body,
        '  </tbody>',
        '</table>',
      ].join('\n')
    }
  }
}

export function previewDocument(
  blocks: TBlock[],
  state?: Record<string, unknown>
): string {
  return blocks
    .map((block) => previewBlock(block, state))
    .filter((section) => section.trim())
    .join('\n\n')
}

const defaultVariablePaths = [
  'app.name',
  'install.name',
  'sandbox.type',
  'sandbox.status',
  'domain.public_domain',
  'domain.internal_domain',
  'runner.status',
]

const excludedRoots = ['secrets', 'cloud_account']

export function getStateVariables(
  state?: Record<string, unknown>
): TStateVariable[] {
  const variables = new Map<string, TStateVariable>()
  defaultVariablePaths.forEach((path) => {
    variables.set(path, { template: `{{.nuon.${path}}}` })
  })
  if (!state) return [...variables.values()]

  const visit = (value: unknown, path: string[]) => {
    if (scalar(value)) {
      variables.set(path.join('.'), {
        template: `{{.nuon.${path.join('.')}}}`,
        value: String(value),
      })
      return
    }
    if (
      !value ||
      typeof value !== 'object' ||
      Array.isArray(value) ||
      excludedRoots.includes(path[0])
    )
      return
    Object.entries(value).forEach(([key, child]) => {
      if (identifier.test(key)) visit(child, [...path, key])
    })
  }
  Object.entries(state).forEach(([key, value]) => visit(value, [key]))
  return [...variables.values()].sort((a, b) =>
    a.template.localeCompare(b.template)
  )
}

export function getArraySources(
  state?: Record<string, unknown>
): TArraySource[] {
  if (!state) return []
  const sources: TArraySource[] = []

  const visit = (value: unknown, path: string[]) => {
    if (!value || typeof value !== 'object' || excludedRoots.includes(path[0]))
      return
    if (Array.isArray(value)) {
      const first = value[0]
      if (first && typeof first === 'object' && !Array.isArray(first)) {
        sources.push({
          path: path.join('.'),
          keys: Object.keys(first).filter(
            (key) =>
              identifier.test(key) &&
              scalar((first as Record<string, unknown>)[key])
          ),
          length: value.length,
        })
      }
      return
    }
    Object.entries(value).forEach(([key, child]) => {
      if (identifier.test(key)) visit(child, [...path, key])
    })
  }
  Object.entries(state).forEach(([key, value]) => visit(value, [key]))
  return sources.sort((a, b) => a.path.localeCompare(b.path))
}
