import type { AnyFieldApi } from '@tanstack/react-form'

export function fieldErrorMessage(field: AnyFieldApi): string | undefined {
  const meta = field.state.meta
  if (!meta.isTouched) return undefined
  const first = meta.errors?.[0]
  if (first === undefined || first === null) return undefined
  if (typeof first === 'string') return first
  if (typeof first === 'object' && 'message' in first) {
    return String((first as { message: unknown }).message)
  }
  return String(first)
}
