import { parseDiffFromFile } from '@pierre/diffs'

export interface IChangeCounts {
  added: number
  removed: number
}

export const endWithNewline = (value: string) =>
  !value || value.endsWith('\n') ? value : `${value}\n`

export const changeCounts = (before: string, after: string): IChangeCounts => {
  const file = (contents: string) => ({
    name: 'change.txt',
    contents: endWithNewline(contents),
    lang: 'text' as const,
  })

  return parseDiffFromFile(file(before), file(after)).hunks.reduce(
    (totals, hunk) => ({
      added: totals.added + hunk.additionLines,
      removed: totals.removed + hunk.deletionLines,
    }),
    { added: 0, removed: 0 }
  )
}

export interface ITextDiff {
  before: string
  after: string
  language: string
  filename?: string
}

export interface ITerraformDiff {
  before: unknown
  after: unknown
  beforeSensitive?: unknown
  afterSensitive?: unknown
  afterUnknown?: unknown
  filename?: string
}

const KNOWN_AFTER_APPLY = Symbol('known-after-apply')
const SENSITIVE_VALUE = Symbol('sensitive-value')

const sortedKeys = (value: Record<string, unknown>) =>
  Object.keys(value).sort((left, right) =>
    left < right ? -1 : left > right ? 1 : 0
  )

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const applyMask = (
  value: unknown,
  mask: unknown,
  replacement: symbol
): unknown => {
  if (mask === true) return replacement
  if (!isRecord(mask) && !Array.isArray(mask)) return value

  if (Array.isArray(value) || Array.isArray(mask)) {
    const source = Array.isArray(value) ? value : []
    const maskEntries = Array.isArray(mask) ? mask : []
    const length = Math.max(source.length, maskEntries.length)
    return Array.from({ length }, (_, index) =>
      applyMask(source[index], maskEntries[index], replacement)
    )
  }

  const source = isRecord(value) ? value : {}
  const maskRecord = isRecord(mask) ? mask : {}
  const keys = new Set([...Object.keys(source), ...Object.keys(maskRecord)])

  return Object.fromEntries(
    [...keys].map((key) => [
      key,
      applyMask(source[key], maskRecord[key], replacement),
    ])
  )
}

const hclKey = (key: string) =>
  /^[A-Za-z_][A-Za-z0-9_-]*$/.test(key) ? key : JSON.stringify(key)

const scalar = (value: unknown): string => {
  if (value === KNOWN_AFTER_APPLY) return '(known after apply)'
  if (value === SENSITIVE_VALUE) return '(sensitive value)'
  if (value === null || value === undefined) return 'null'
  if (typeof value === 'string') return JSON.stringify(value)
  if (typeof value === 'number')
    return Number.isFinite(value) ? String(value) : 'null'
  if (typeof value === 'boolean') return String(value)
  return JSON.stringify(value)
}

const serializeValue = (value: unknown, indent: number): string[] => {
  const padding = '  '.repeat(indent)

  if (Array.isArray(value)) {
    if (!value.length) return ['[]']

    const lines = ['[']
    value.forEach((item) => {
      const rendered = serializeValue(item, indent + 1)
      lines.push(`${'  '.repeat(indent + 1)}${rendered[0]}`)
      lines.push(...rendered.slice(1))
      lines[lines.length - 1] = `${lines[lines.length - 1]},`
    })
    lines.push(`${padding}]`)
    return lines
  }

  if (isRecord(value)) {
    if (!Object.keys(value).length) return ['{}']

    const lines = ['{']
    sortedKeys(value).forEach((key) => {
      const rendered = serializeValue(value[key], indent + 1)
      lines.push(`${'  '.repeat(indent + 1)}${hclKey(key)} = ${rendered[0]}`)
      lines.push(...rendered.slice(1))
    })
    lines.push(`${padding}}`)
    return lines
  }

  return [scalar(value)]
}

export const serializeTerraform = (
  value: unknown,
  {
    sensitive,
    unknown,
  }: {
    sensitive?: unknown
    unknown?: unknown
  } = {}
): string => {
  const withUnknowns = applyMask(value, unknown, KNOWN_AFTER_APPLY)
  const masked = applyMask(withUnknowns, sensitive, SENSITIVE_VALUE)

  if (!isRecord(masked)) return serializeValue(masked, 0).join('\n')

  return sortedKeys(masked)
    .flatMap((key) => {
      const rendered = serializeValue(masked[key], 0)
      return [`${hclKey(key)} = ${rendered[0]}`, ...rendered.slice(1)]
    })
    .join('\n')
}

export const textDiff = (diff: ITextDiff): ITextDiff => diff

export const terraformDiff = ({
  before,
  after,
  beforeSensitive,
  afterSensitive,
  afterUnknown,
  filename,
}: ITerraformDiff): ITextDiff => ({
  before: serializeTerraform(before, { sensitive: beforeSensitive }),
  after: serializeTerraform(after, {
    sensitive: afterSensitive,
    unknown: afterUnknown,
  }),
  language: 'terraform',
  filename,
})
