import type { TCompositeError } from '@/types'

export const CONFIG_VALIDATION_FAILED = 'app_branch_run.config_validation_failed'

export const isConfigValidationError = (error?: TCompositeError) =>
  error?.type === CONFIG_VALIDATION_FAILED

export function configDiagnosticLines(error?: TCompositeError): string[] {
  if (!isConfigValidationError(error)) return []
  const body =
    error?.sections?.find((section) => section?.kind === 'code')?.body ?? ''
  return body
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}
