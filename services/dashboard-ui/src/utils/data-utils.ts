import type { TKeyValue } from '@/types'

export function objectToKeyValueArray(obj: Record<string, any>): TKeyValue[] {
  return Object.entries(obj).map(([key, value]) => ({
    key,
    value: formatValue(value),
    type: getValueType(value),
  }))
}

function getValueType(value: any): string {
  if (value === null) {
    return 'null'
  }

  if (value === undefined) {
    return 'undefined'
  }

  if (Array.isArray(value)) {
    return 'array'
  }

  if (typeof value === 'object') {
    return 'object'
  }

  return typeof value // 'string', 'number', 'boolean', 'function', etc.
}

function formatValue(value: any): string {
  if (value === null) {
    return 'null'
  }

  if (value === undefined) {
    return 'undefined'
  }

  if (typeof value === 'string') {
    return value
  }

  if (typeof value === 'boolean' || typeof value === 'number') {
    return String(value)
  }

  if (typeof value === 'object') {
    if (Array.isArray(value)) {
      return `[${value.map((item) => formatValue(item)).join(', ')}]`
    }

    try {
      return JSON.stringify(value, null, 2)
    } catch (error) {
      return '[Object - Unable to serialize]'
    }
  }

  if (typeof value === 'function') {
    return '[Function]'
  }

  return String(value)
}
