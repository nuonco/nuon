import YAML from 'yaml'
import type {
  TTerraformResourceChange,
  TTerraformChangeAction,
  TTerraformOutputChange,
  TTerraformPlan,
} from '@/types'

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
} {
  const resourceChanges: TTerraformResourceChange[] = []
  const outputChanges: TTerraformOutputChange[] = []

  const resourceSummary: Record<string, number> = Object.fromEntries(
    OPERATION_TYPES.map((op) => [op, 0])
  )
  const outputSummary: Record<string, number> = Object.fromEntries(
    OPERATION_TYPES.map((op) => [op, 0])
  )

  // Resource Changes
  for (const rc of plan.resource_changes) {
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

    for (const action of rc.change.actions) {
      incrementSummary(resourceSummary, action)
      if (action === 'replace') {
        incrementSummary(resourceSummary, 'delete')
        incrementSummary(resourceSummary, 'create')
      }
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

  // Output Changes
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

      for (const action of oc.actions) {
        incrementSummary(outputSummary, action)
        if (action === 'replace') {
          incrementSummary(outputSummary, 'delete')
          incrementSummary(outputSummary, 'create')
        }
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

  OPERATION_TYPES.forEach((op) => {
    resourceSummary[op] = resourceSummary[op] || 0
    outputSummary[op] = outputSummary[op] || 0
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
  }
}
