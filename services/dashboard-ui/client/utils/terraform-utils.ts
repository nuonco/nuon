import YAML from 'yaml'
import type {
  TTerraformResourceChange,
  TTerraformChangeAction,
  TTerraformOutputChange,
  TTerraformPlan,
} from '@/types'

export type DiffLine = {
  indent: number
  prefix: '+' | '-' | '~' | ' '
  text: string
  type: 'added' | 'removed' | 'changed' | 'unchanged'
}

export function isComplex(val: any): boolean {
  return val !== null && typeof val === 'object'
}

export function deepEqual(a: any, b: any): boolean {
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
    return keysA.every(
      (key) => keysB.includes(key) && deepEqual(a[key], b[key])
    )
  }

  return false
}

function isEmptyCollection(val: any): boolean {
  if (val == null) return true
  if (Array.isArray(val)) return val.length === 0
  if (typeof val === 'object') return Object.keys(val).length === 0
  return false
}

export function semanticEqual(a: any, b: any): boolean {
  if (a === b) return true
  if (isEmptyCollection(a) && isEmptyCollection(b)) return true
  if (a == null || b == null) return false
  if (typeof a !== typeof b) return false

  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false
    return a.every((item, index) => semanticEqual(item, b[index]))
  }

  if (typeof a === 'object') {
    const allKeys = new Set([...Object.keys(a), ...Object.keys(b)])
    return [...allKeys].every((key) => semanticEqual(a[key], b[key]))
  }

  return false
}

const PREFIX_MAP = {
  '+': 'added',
  '-': 'removed',
  '~': 'changed',
  ' ': 'unchanged',
} as const

function renderScalar(val: any): string {
  if (val === null || val === undefined) return 'null'
  if (typeof val === 'string') return JSON.stringify(val)
  return String(val)
}

function renderFullValue(
  val: any,
  prefix: '+' | '-',
  indent: number,
  key?: string
): DiffLine[] {
  const type = PREFIX_MAP[prefix]
  const keyPrefix = key !== undefined ? `${JSON.stringify(key)}: ` : ''

  if (!isComplex(val)) {
    return [{ indent, prefix, type, text: `${keyPrefix}${renderScalar(val)}` }]
  }

  if (Array.isArray(val)) {
    if (val.length === 0) {
      return [{ indent, prefix, type, text: `${keyPrefix}[]` }]
    }
    const lines: DiffLine[] = [{ indent, prefix, type, text: `${keyPrefix}[` }]
    val.forEach((item) => {
      lines.push(...renderFullValue(item, prefix, indent + 1))
    })
    lines.push({ indent, prefix, type, text: ']' })
    return lines
  }

  const keys = Object.keys(val)
  if (keys.length === 0) {
    return [{ indent, prefix, type, text: `${keyPrefix}{}` }]
  }
  const lines: DiffLine[] = [{ indent, prefix, type, text: `${keyPrefix}{` }]
  keys.forEach((k) => {
    lines.push(...renderFullValue(val[k], prefix, indent + 1, k))
  })
  lines.push({ indent, prefix, type, text: '}' })
  return lines
}

function maybeParseJsonString(val: any): any {
  if (typeof val === 'string' && isStringJson(val)) {
    return JSON.parse(val)
  }
  return val
}

export function generateDiffLines(
  before: any,
  after: any,
  indent = 0,
  maxDepth = 10
): DiffLine[] {
  if (indent > maxDepth) {
    const text =
      before !== undefined && after !== undefined
        ? `${JSON.stringify(before).slice(0, 60)} -> ${JSON.stringify(after).slice(0, 60)}`
        : JSON.stringify(before ?? after).slice(0, 120)
    return [{ indent, prefix: '~', type: 'changed', text }]
  }

  before = maybeParseJsonString(before)
  after = maybeParseJsonString(after)

  if (
    (before === null || before === undefined) &&
    (after === null || after === undefined)
  ) {
    return []
  }

  if (before === null || before === undefined) {
    return renderFullValue(after, '+', indent)
  }

  if (after === null || after === undefined) {
    return renderFullValue(before, '-', indent)
  }

  if (!isComplex(before) && !isComplex(after)) {
    if (deepEqual(before, after)) {
      return [
        {
          indent,
          prefix: ' ',
          type: 'unchanged',
          text: renderScalar(after),
        },
      ]
    }
    return [
      { indent, prefix: '-', type: 'removed', text: renderScalar(before) },
      { indent, prefix: '+', type: 'added', text: renderScalar(after) },
    ]
  }

  if (Array.isArray(before) && Array.isArray(after)) {
    if (deepEqual(before, after)) {
      return renderFullValue(after, ' ' as any, indent).map((l) => ({
        ...l,
        prefix: ' ' as const,
        type: 'unchanged' as const,
      }))
    }

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
        const sub = renderFullValue(after[i], ' ' as any, indent + 1).map(
          (l) => ({
            ...l,
            prefix: ' ' as const,
            type: 'unchanged' as const,
          })
        )
        innerLines.push(...sub)
      } else {
        anyChanged = true
        innerLines.push(
          ...generateDiffLines(before[i], after[i], indent + 1, maxDepth)
        )
      }
    }

    const bracketPrefix = anyChanged ? '~' : ' '
    const bracketType = anyChanged ? 'changed' : 'unchanged'
    return [
      {
        indent,
        prefix: bracketPrefix as DiffLine['prefix'],
        type: bracketType as DiffLine['type'],
        text: '[',
      },
      ...innerLines,
      {
        indent,
        prefix: bracketPrefix as DiffLine['prefix'],
        type: bracketType as DiffLine['type'],
        text: ']',
      },
    ]
  }

  if (
    isComplex(before) &&
    !Array.isArray(before) &&
    isComplex(after) &&
    !Array.isArray(after)
  ) {
    const allKeys = [
      ...new Set([...Object.keys(before), ...Object.keys(after)]),
    ]
    let anyChanged = false
    const innerLines: DiffLine[] = []

    allKeys.forEach((key) => {
      const bVal = before[key]
      const aVal = after[key]

      if (bVal === undefined) {
        anyChanged = true
        renderFullValue(aVal, '+', indent + 1, key).forEach((l) =>
          innerLines.push(l)
        )
      } else if (aVal === undefined) {
        anyChanged = true
        renderFullValue(bVal, '-', indent + 1, key).forEach((l) =>
          innerLines.push(l)
        )
      } else if (deepEqual(bVal, aVal)) {
        const scalar = !isComplex(aVal)
        if (scalar) {
          innerLines.push({
            indent: indent + 1,
            prefix: ' ',
            type: 'unchanged',
            text: `${JSON.stringify(key)}: ${renderScalar(aVal)}`,
          })
        } else {
          const sub = renderFullValue(aVal, ' ' as any, indent + 1, key).map(
            (l) => ({
              ...l,
              prefix: ' ' as const,
              type: 'unchanged' as const,
            })
          )
          innerLines.push(...sub)
        }
      } else {
        anyChanged = true
        if (
          (bVal === null && isComplex(aVal)) ||
          (aVal === null && isComplex(bVal))
        ) {
          renderFullValue(bVal, '-', indent + 1, key).forEach((l) =>
            innerLines.push(l)
          )
          renderFullValue(aVal, '+', indent + 1, key).forEach((l) =>
            innerLines.push(l)
          )
        } else {
          const childLines = generateDiffLines(bVal, aVal, indent + 1, maxDepth)
          if (childLines.length > 0) {
            const firstLine = childLines[0]
            childLines[0] = {
              ...firstLine,
              text: `${JSON.stringify(key)}: ${firstLine.text}`,
            }
          }
          innerLines.push(...childLines)
        }
      }
    })

    const bracePrefix = anyChanged ? '~' : ' '
    const braceType = anyChanged ? 'changed' : 'unchanged'
    return [
      {
        indent,
        prefix: bracePrefix as DiffLine['prefix'],
        type: braceType as DiffLine['type'],
        text: '{',
      },
      ...innerLines,
      {
        indent,
        prefix: bracePrefix as DiffLine['prefix'],
        type: braceType as DiffLine['type'],
        text: '}',
      },
    ]
  }

  // Type mismatch
  return [
    ...renderFullValue(before, '-', indent),
    ...renderFullValue(after, '+', indent),
  ]
}

export function cleanString(str: string): string {
  let s = str ?? ''
  if (s.startsWith('"') && s.endsWith('"')) {
    s = s.slice(1, -1)
  }
  // Replace double-escaped newlines with real newlines
  s = s.replace(/\\n/g, '\n')
  return s
}

export function isStringYaml(str: string): boolean {
  try {
    const parsed = YAML.parse(str)
    return typeof parsed === 'object' && parsed !== null
  } catch {
    return false
  }
}

export function isStringJson(str: string): boolean {
  try {
    const parsed = JSON.parse(str)
    return typeof parsed === 'object' && parsed !== null
  } catch {
    return false
  }
}

export function isTerraformEscapedYaml(str: string): {
  isTerraformYaml: boolean
  yamlContent?: string
} {
  // Check if this looks like escaped YAML content (has \" patterns and YAML structure)
  if (str.includes('\\"') && str.includes(':')) {
    // Try to clean it up
    const cleaned = str
      .replace(/\\n/g, '\n') // Convert \n to actual newlines
      .replace(/\\"/g, '"') // Convert \" to actual quotes

    // Check if the cleaned version looks like YAML structure
    if (
      cleaned.includes(':') &&
      (cleaned.includes('-') || cleaned.includes('  '))
    ) {
      return { isTerraformYaml: true, yamlContent: cleaned }
    }
  }

  return { isTerraformYaml: false }
}

type TIsTerraformWithYaml = {
  isArrayYaml: boolean
  yamlContent?: string
}

export function isTerraformArrayWithYaml(str: string): TIsTerraformWithYaml {
  try {
    const parsed = JSON.parse(str)

    if (
      Array.isArray(parsed) &&
      parsed.length === 1 &&
      typeof parsed[0] === 'string'
    ) {
      // Clean the string element more thoroughly
      let cleanedElement = parsed[0]

      // Remove outer quotes if present
      if (cleanedElement.startsWith('"') && cleanedElement.endsWith('"')) {
        cleanedElement = cleanedElement.slice(1, -1)
      }

      // Replace escaped newlines and quotes
      cleanedElement = cleanedElement.replace(/\\n/g, '\n').replace(/\\"/g, '"')

      // Try to parse as YAML - but be more lenient about what constitutes valid YAML
      try {
        const yamlParsed = YAML.parse(cleanedElement)
        if (yamlParsed && typeof yamlParsed === 'object') {
          return { isArrayYaml: true, yamlContent: cleanedElement }
        }
      } catch (yamlError) {
        // If YAML parsing fails, but it looks like YAML structure, still treat it as YAML
        if (
          cleanedElement.includes(':') &&
          (cleanedElement.includes('-') || cleanedElement.includes('  '))
        ) {
          return { isArrayYaml: true, yamlContent: cleanedElement }
        }
      }
    }
    return { isArrayYaml: false }
  } catch (error) {
    return { isArrayYaml: false }
  }
}

export function detectValueFormat(value: string): {
  displayValue: string
  language: string
  showLineNumbers: boolean
} {
  const cleanValue = cleanString(value)

  // Check for Terraform escaped YAML FIRST - before any other checks
  const terraformEscapedCheck = isTerraformEscapedYaml(cleanValue)

  // Only proceed with other checks if it's NOT Terraform escaped YAML
  let isJSON = false
  let isYAML = false
  let terraformArrayCheck: TIsTerraformWithYaml = { isArrayYaml: false }

  if (!terraformEscapedCheck.isTerraformYaml) {
    isJSON = isStringJson(cleanValue)
    if (!isJSON) {
      isYAML = isStringYaml(cleanValue)
      if (!isYAML) {
        terraformArrayCheck = isTerraformArrayWithYaml(cleanValue)
      }
    }
  }

  if (
    terraformEscapedCheck.isTerraformYaml &&
    terraformEscapedCheck.yamlContent
  ) {
    return {
      displayValue: terraformEscapedCheck.yamlContent,
      language: 'yaml',
      showLineNumbers: true,
    }
  } else if (isJSON) {
    return {
      displayValue: JSON.stringify(JSON.parse(cleanValue), null, 2),
      language: 'json',
      showLineNumbers: true,
    }
  } else if (isYAML) {
    return {
      displayValue: cleanValue,
      language: 'yaml',
      showLineNumbers: true,
    }
  } else if (
    terraformArrayCheck.isArrayYaml &&
    terraformArrayCheck.yamlContent
  ) {
    return {
      displayValue: terraformArrayCheck.yamlContent,
      language: 'yaml',
      showLineNumbers: true,
    }
  } else {
    // Default behavior for all other cases
    return {
      displayValue: cleanValue,
      language: 'sh',
      showLineNumbers: false,
    }
  }
}

const OPERATION_TYPES = [
  'create',
  'update',
  'delete',
  'replace',
  'read',
  'no-op',
  'drift', // Add drift to operation types
]

export function isOutputAfterUnknown(afterUnknown: any): boolean {
  return (
    !!afterUnknown &&
    typeof afterUnknown === 'object' &&
    Object.keys(afterUnknown).length > 0
  )
}

function incrementSummary(summaryObj: Record<string, number>, action: string) {
  if (summaryObj[action] !== undefined) {
    summaryObj[action] += 1
  } else {
    summaryObj[action] = 1
  }
}

function mergeAfterUnknown(after: any, afterUnknown: any): any {
  if (!afterUnknown || typeof afterUnknown !== 'object') {
    return after
  }

  const merged = after ? { ...after } : {}

  function processUnknown(obj: any, unknown: any, target: any) {
    if (!unknown || typeof unknown !== 'object') return

    for (const [key, value] of Object.entries(unknown)) {
      if (value === true) {
        target[key] = 'Known after apply'
      } else if (typeof value === 'object' && value !== null) {
        if (!target[key]) target[key] = {}
        processUnknown(obj?.[key], value, target[key])
      }
    }
  }

  processUnknown(after, afterUnknown, merged)
  return merged
}

export function parseTerraformPlan(plan: TTerraformPlan): {
  resources: {
    summary: Record<TTerraformChangeAction, number>
    changes: TTerraformResourceChange[]
  }
  outputs: {
    summary: Record<TTerraformChangeAction, number>
    changes: TTerraformOutputChange[]
  }
  drift: {
    summary: Record<TTerraformChangeAction, number>
    changes: TTerraformResourceChange[]
  }
} {
  const resourceChanges: TTerraformResourceChange[] = []
  const outputChanges: TTerraformOutputChange[] = []
  const driftChanges: TTerraformResourceChange[] = []

  const resourceSummary: Record<string, number> = Object.fromEntries(
    OPERATION_TYPES.map((op) => [op, 0])
  )
  const outputSummary: Record<string, number> = Object.fromEntries(
    OPERATION_TYPES.map((op) => [op, 0])
  )
  const driftSummary: Record<string, number> = Object.fromEntries(
    OPERATION_TYPES.map((op) => [op, 0])
  )

  function isReplaceActions(actions: string[]): boolean {
    if (actions.length !== 2) return false
    return (
      (actions[0] === 'delete' && actions[1] === 'create') ||
      (actions[0] === 'create' && actions[1] === 'delete')
    )
  }

  // Resource Drift (new section)
  if (Array.isArray(plan.resource_drift)) {
    for (const rd of plan.resource_drift) {
      const mergedAfter = mergeAfterUnknown(
        rd.change.after,
        rd.change.after_unknown
      )

      if (semanticEqual(rd.change.before, mergedAfter)) {
        continue
      }

      if (isReplaceActions(rd.change.actions)) {
        incrementSummary(driftSummary, 'replace')
        driftChanges.push({
          address: rd.address,
          module: rd.module_address ?? null,
          resource: rd.type,
          name: rd.name,
          action: 'replace',
          before: rd.change.before,
          after: mergedAfter,
        })
        continue
      }

      for (const action of rd.change.actions) {
        incrementSummary(driftSummary, action)
        driftChanges.push({
          address: rd.address,
          module: rd.module_address ?? null,
          resource: rd.type,
          name: rd.name,
          action,
          before: rd.change.before,
          after: mergedAfter,
        })
      }
    }
  }

  // Resource Changes
  for (const rc of plan.resource_changes ?? []) {
    const mergedAfter = mergeAfterUnknown(
      rc.change.after,
      rc.change.after_unknown
    )

    if (rc.change.actions.length === 1 && rc.change.actions[0] === 'read') {
      incrementSummary(resourceSummary, 'read')
      resourceChanges.push({
        address: rc.address,
        module: rc.module_address ?? null,
        resource: rc.type,
        name: rc.name,
        action: 'read',
        before: rc.change.before,
        after: mergedAfter,
      })
      continue
    }

    if (isReplaceActions(rc.change.actions)) {
      incrementSummary(resourceSummary, 'replace')
      resourceChanges.push({
        address: rc.address,
        module: rc.module_address ?? null,
        resource: rc.type,
        name: rc.name,
        action: 'replace',
        before: rc.change.before,
        after: mergedAfter,
      })
      continue
    }

    for (const action of rc.change.actions) {
      incrementSummary(resourceSummary, action)
      resourceChanges.push({
        address: rc.address,
        module: rc.module_address ?? null,
        resource: rc.type,
        name: rc.name,
        action,
        before: rc.change.before,
        after: mergedAfter,
      })
    }
  }

  // Output Changes (existing logic)
  if (plan.output_changes) {
    for (const [output, oc] of Object.entries(plan.output_changes)) {
      const mergedAfter = mergeAfterUnknown(oc.after, oc.after_unknown)

      if (oc.actions.length === 1 && oc.actions[0] === 'read') {
        incrementSummary(outputSummary, 'read')
        outputChanges.push({
          output,
          action: 'read',
          before: oc.before,
          after: mergedAfter,
          afterUnknown: oc.after_unknown,
          afterSensitive: oc.after_sensitive,
          beforeSensitive: oc.before_sensitive,
        })
        continue
      }

      if (isReplaceActions(oc.actions)) {
        incrementSummary(outputSummary, 'replace')
        outputChanges.push({
          output,
          action: 'replace',
          before: oc.before,
          after: mergedAfter,
          afterUnknown: oc.after_unknown,
          afterSensitive: oc.after_sensitive,
          beforeSensitive: oc.before_sensitive,
        })
        continue
      }

      for (const action of oc.actions) {
        incrementSummary(outputSummary, action)
        outputChanges.push({
          output,
          action,
          before: oc.before,
          after: mergedAfter,
          afterUnknown: oc.after_unknown,
          afterSensitive: oc.after_sensitive,
          beforeSensitive: oc.before_sensitive,
        })
      }
    }
  }

  // Ensure all operation types are represented
  OPERATION_TYPES.forEach((op) => {
    resourceSummary[op] = resourceSummary[op] || 0
    outputSummary[op] = outputSummary[op] || 0
    driftSummary[op] = driftSummary[op] || 0
  })

  return {
    resources: {
      summary: resourceSummary,
      changes: resourceChanges,
    },
    outputs: {
      summary: outputSummary,
      changes: outputChanges,
    },
    drift: {
      summary: driftSummary,
      changes: driftChanges,
    },
  }
}
