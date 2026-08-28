import { useMemo, useState } from 'react'
import type { TPlan, TPlanResourceChange } from './types'

const KNOWN_AFTER_APPLY = 'Known after apply'

const isComplex = (val: unknown): val is Record<string, unknown> | unknown[] =>
  val !== null && typeof val === 'object'

const deepEqual = (a: any, b: any): boolean => {
  if (a === b) return true
  if (a == null || b == null) return a === b
  if (typeof a !== typeof b) return false
  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false
    return a.every((item, index) => deepEqual(item, b[index]))
  }
  if (typeof a === 'object') {
    const keysA = Object.keys(a)
    const keysB = Object.keys(b)
    if (keysA.length !== keysB.length) return false
    return keysA.every((key) => keysB.includes(key) && deepEqual(a[key], b[key]))
  }
  return false
}

const isStringJson = (str: string): boolean => {
  const trimmed = str.trim()
  if (!(trimmed.startsWith('{') || trimmed.startsWith('['))) return false
  try {
    const parsed = JSON.parse(trimmed)
    return typeof parsed === 'object' && parsed !== null
  } catch {
    return false
  }
}

const maybeParseJsonString = (val: any): any =>
  typeof val === 'string' && isStringJson(val) ? JSON.parse(val.trim()) : val

export function mergeAfterUnknown(after: any, afterUnknown: any): any {
  if (!afterUnknown || typeof afterUnknown !== 'object') return after
  const merged = after ? { ...after } : {}
  const process = (unknown: any, target: any) => {
    if (!unknown || typeof unknown !== 'object') return
    for (const [key, value] of Object.entries(unknown)) {
      if (value === true) {
        target[key] = KNOWN_AFTER_APPLY
      } else if (typeof value === 'object' && value !== null) {
        if (!target[key] || typeof target[key] !== 'object') target[key] = {}
        process(value, target[key])
      }
    }
  }
  process(afterUnknown, merged)
  return merged
}

export type DiffLine = {
  indent: number
  prefix: '+' | '-' | '~' | ' '
  text: string
  type: 'added' | 'removed' | 'changed' | 'unchanged'
}

const PREFIX_MAP = { '+': 'added', '-': 'removed', '~': 'changed', ' ': 'unchanged' } as const

const renderScalar = (val: any): string => {
  if (val === null || val === undefined) return 'null'
  if (typeof val === 'string') return JSON.stringify(val)
  return String(val)
}

function renderFullValue(val: any, prefix: '+' | '-' | ' ', indent: number, key?: string): DiffLine[] {
  const type = PREFIX_MAP[prefix]
  const keyPrefix = key !== undefined ? `${JSON.stringify(key)}: ` : ''
  if (typeof val === 'string' && val.includes('\n')) {
    return [
      { indent, prefix, type, text: `${keyPrefix}|` },
      ...val.split('\n').map((line) => ({ indent: indent + 1, prefix, type, text: line })),
    ]
  }
  if (!isComplex(val)) {
    return [{ indent, prefix, type, text: `${keyPrefix}${renderScalar(val)}` }]
  }
  if (Array.isArray(val)) {
    if (val.length === 0) return [{ indent, prefix, type, text: `${keyPrefix}[]` }]
    const lines: DiffLine[] = [{ indent, prefix, type, text: `${keyPrefix}[` }]
    val.forEach((item) => lines.push(...renderFullValue(item, prefix, indent + 1)))
    lines.push({ indent, prefix, type, text: ']' })
    return lines
  }
  const keys = Object.keys(val)
  if (keys.length === 0) return [{ indent, prefix, type, text: `${keyPrefix}{}` }]
  const lines: DiffLine[] = [{ indent, prefix, type, text: `${keyPrefix}{` }]
  keys.forEach((k) => lines.push(...renderFullValue((val as any)[k], prefix, indent + 1, k)))
  lines.push({ indent, prefix, type, text: '}' })
  return lines
}

export function generateDiffLines(before: any, after: any, indent = 0, maxDepth = 10): DiffLine[] {
  if (indent > maxDepth) {
    const text =
      before !== undefined && after !== undefined
        ? `${JSON.stringify(before).slice(0, 60)} -> ${JSON.stringify(after).slice(0, 60)}`
        : JSON.stringify(before ?? after).slice(0, 120)
    return [{ indent, prefix: '~', type: 'changed', text }]
  }

  before = maybeParseJsonString(before)
  after = maybeParseJsonString(after)

  if ((before === null || before === undefined) && (after === null || after === undefined)) return []
  if (before === null || before === undefined) return renderFullValue(after, '+', indent)
  if (after === null || after === undefined) return renderFullValue(before, '-', indent)

  if (!isComplex(before) && !isComplex(after)) {
    if ((typeof before === 'string' && before.includes('\n')) || (typeof after === 'string' && after.includes('\n'))) {
      if (deepEqual(before, after)) return renderFullValue(after, ' ', indent)
      return [...renderFullValue(before, '-', indent), ...renderFullValue(after, '+', indent)]
    }
    if (deepEqual(before, after)) {
      return [{ indent, prefix: ' ', type: 'unchanged', text: renderScalar(after) }]
    }
    return [
      { indent, prefix: '-', type: 'removed', text: renderScalar(before) },
      { indent, prefix: '+', type: 'added', text: renderScalar(after) },
    ]
  }

  if (Array.isArray(before) && Array.isArray(after)) {
    if (deepEqual(before, after)) return renderFullValue(after, ' ', indent)

    let anyChanged = false
    const innerLines: DiffLine[] = []
    const maxLen = Math.max(before.length, after.length)
    for (let i = 0; i < maxLen; i++) {
      if (i >= before.length) {
        anyChanged = true
        innerLines.push(...renderFullValue(after[i], '+', indent + 1))
      } else if (i >= after.length) {
        anyChanged = true
        innerLines.push(...renderFullValue(before[i], '-', indent + 1))
      } else if (deepEqual(before[i], after[i])) {
        innerLines.push(...renderFullValue(after[i], ' ', indent + 1))
      } else {
        anyChanged = true
        innerLines.push(...generateDiffLines(before[i], after[i], indent + 1, maxDepth))
      }
    }
    const prefix = anyChanged ? '~' : ' '
    const type = anyChanged ? 'changed' : 'unchanged'
    return [
      { indent, prefix, type, text: '[' },
      ...innerLines,
      { indent, prefix, type, text: ']' },
    ]
  }

  if (isComplex(before) && !Array.isArray(before) && isComplex(after) && !Array.isArray(after)) {
    const allKeys = Array.from(new Set([...Object.keys(before), ...Object.keys(after)]))
    let anyChanged = false
    const innerLines: DiffLine[] = []

    allKeys.forEach((key) => {
      const bVal = (before as any)[key]
      const aVal = (after as any)[key]
      if (bVal === undefined) {
        anyChanged = true
        innerLines.push(...renderFullValue(aVal, '+', indent + 1, key))
      } else if (aVal === undefined) {
        anyChanged = true
        innerLines.push(...renderFullValue(bVal, '-', indent + 1, key))
      } else if (deepEqual(bVal, aVal)) {
        innerLines.push(...renderFullValue(aVal, ' ', indent + 1, key))
      } else {
        anyChanged = true
        if ((bVal === null && isComplex(aVal)) || (aVal === null && isComplex(bVal))) {
          innerLines.push(...renderFullValue(bVal, '-', indent + 1, key))
          innerLines.push(...renderFullValue(aVal, '+', indent + 1, key))
        } else {
          const childLines = generateDiffLines(bVal, aVal, indent + 1, maxDepth)
          if (childLines.length > 0) {
            childLines[0] = { ...childLines[0], text: `${JSON.stringify(key)}: ${childLines[0].text}` }
          }
          innerLines.push(...childLines)
        }
      }
    })

    const prefix = anyChanged ? '~' : ' '
    const type = anyChanged ? 'changed' : 'unchanged'
    return [
      { indent, prefix, type, text: '{' },
      ...innerLines,
      { indent, prefix, type, text: '}' },
    ]
  }

  return [...renderFullValue(before, '-', indent), ...renderFullValue(after, '+', indent)]
}

export function findPlanResource(plan: TPlan, address: string): TPlanResourceChange | undefined {
  const changed = plan.resource_changes?.find((rc) => rc.address === address)
  if (changed?.change && (changed.change.before !== undefined || changed.change.after !== undefined)) {
    return changed
  }
  const drifted = plan.resource_drift?.find((rc) => rc.address === address)
  if (drifted?.change && (drifted.change.before !== undefined || drifted.change.after !== undefined)) {
    return drifted
  }
  return changed ?? drifted
}

type PropertyRow = {
  key: string
  before: any
  after: any
  changed: boolean
}

const propertyRows = (before: any, after: any): PropertyRow[] => {
  const beforeKeys = isComplex(before) && !Array.isArray(before) ? Object.keys(before) : []
  const afterKeys = isComplex(after) && !Array.isArray(after) ? Object.keys(after) : []
  return Array.from(new Set([...beforeKeys, ...afterKeys])).map((key) => {
    const beforeValue = (before as any)?.[key] ?? null
    const afterValue = (after as any)?.[key] ?? null
    return { key, before: beforeValue, after: afterValue, changed: !deepEqual(beforeValue, afterValue) }
  })
}

const rowPrefix = (action: string, changed: boolean): { char: DiffLine['prefix']; type: DiffLine['type'] } => {
  if (!changed) return { char: ' ', type: 'unchanged' }
  switch (action) {
    case 'create':
      return { char: '+', type: 'added' }
    case 'delete':
    case 'destroy':
      return { char: '-', type: 'removed' }
    default:
      return { char: '~', type: 'changed' }
  }
}

const COLLAPSE_THRESHOLD = 80
const INITIAL_VISIBLE = 50

const DiffLines = ({ lines, changedOnly = false }: { lines: DiffLine[]; changedOnly?: boolean }) => {
  const [expanded, setExpanded] = useState(false)
  const [showContext, setShowContext] = useState(false)
  const hiddenContext = changedOnly ? lines.filter((line) => line.type === 'unchanged').length : 0
  const relevantLines = changedOnly && !showContext ? lines.filter((line) => line.type !== 'unchanged') : lines
  const needsCollapse = relevantLines.length > COLLAPSE_THRESHOLD
  const visible = needsCollapse && !expanded ? relevantLines.slice(0, INITIAL_VISIBLE) : relevantLines
  return (
    <>
      {hiddenContext > 0 && (
        <button type="button" className="plan-diff-context" onClick={() => setShowContext((value) => !value)}>
          {showContext ? 'Hide unchanged lines' : `Show ${hiddenContext} unchanged ${hiddenContext === 1 ? 'line' : 'lines'}`}
        </button>
      )}
      {visible.map((line, idx) => (
        <div key={idx} className={`plan-diff-line ${line.type}`}>
          <span className="plan-diff-prefix">{line.prefix}</span>
          <span style={{ paddingLeft: `${(line.indent + 1) * 14}px`, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
            {line.text}
          </span>
        </div>
      ))}
      {needsCollapse && (
        <button type="button" className="plan-diff-expand" onClick={() => setExpanded((v) => !v)}>
          {expanded ? 'Show less' : `Show ${relevantLines.length - INITIAL_VISIBLE} more lines`}
        </button>
      )}
    </>
  )
}

export const ValueDiff = ({ before, after }: { before: unknown; after: unknown }) => (
  <div className="plan-diff">
    <DiffLines lines={generateDiffLines(before, after)} changedOnly />
  </div>
)

const ScalarValue = ({ value }: { value: any }) => {
  const formatted =
    value === null || value === undefined ? 'null' : value === '' ? '""' : typeof value === 'object' ? JSON.stringify(value) : String(value)
  return (
    <span className={formatted === KNOWN_AFTER_APPLY ? 'plan-diff-unknown' : undefined} title={formatted}>
      {formatted}
    </span>
  )
}

export const PlanResourceDiff = ({ resource, action, changedOnly = false }: { resource: TPlanResourceChange; action: string; changedOnly?: boolean }) => {
  const [showContext, setShowContext] = useState(false)
  const before = resource.change?.before ?? null
  const after = useMemo(
    () => mergeAfterUnknown(resource.change?.after ?? null, resource.change?.after_unknown),
    [resource]
  )
  const allRows = useMemo(() => propertyRows(before, after), [before, after])
  const hiddenProperties = changedOnly ? allRows.filter((row) => !row.changed).length : 0
  const rows = changedOnly && !showContext ? allRows.filter((row) => row.changed) : allRows

  if (allRows.length === 0) {
    return <div className="plan-diff plan-diff-empty">No attribute values recorded in the plan for this resource.</div>
  }

  return (
    <div className="plan-diff">
      {hiddenProperties > 0 && (
        <button type="button" className="plan-diff-context" onClick={() => setShowContext((value) => !value)}>
          {showContext ? 'Hide unchanged properties' : `Show ${hiddenProperties} unchanged ${hiddenProperties === 1 ? 'property' : 'properties'}`}
        </button>
      )}
      {rows.map((row) => {
        const prefix = rowPrefix(action, row.changed)
        const complex =
          isComplex(row.before) ||
          isComplex(row.after) ||
          (typeof row.before === 'string' && isStringJson(row.before)) ||
          (typeof row.after === 'string' && isStringJson(row.after))

        if (complex) {
          return (
            <div key={row.key}>
              <div className={`plan-diff-line ${row.changed ? prefix.type : 'unchanged'}`}>
                <span className="plan-diff-prefix">{prefix.char}</span>
                <span className="plan-diff-key">{row.key}:</span>
              </div>
              <DiffLines lines={generateDiffLines(row.before, row.after)} changedOnly={changedOnly && !showContext} />
            </div>
          )
        }

        return (
          <div key={row.key} className={`plan-diff-line ${prefix.type}`}>
            <span className="plan-diff-prefix">{prefix.char}</span>
            <span className="plan-diff-scalar">
              <span className="plan-diff-key">{row.key}:</span>{' '}
              {row.changed ? (
                <>
                  <span className="plan-diff-before">
                    <ScalarValue value={row.before} />
                  </span>
                  <span className="plan-diff-arrow"> {'->'} </span>
                  <ScalarValue value={row.after} />
                </>
              ) : (
                <ScalarValue value={row.after} />
              )}
            </span>
          </div>
        )
      })}
    </div>
  )
}
