import type { TAPIError } from '@/types'

export const actionErrorMessage = (
  error: unknown,
  fallback: string
): string => {
  if (!error) return ''
  if (error instanceof Error) return error.message || fallback
  const apiError = error as TAPIError
  return apiError?.error || apiError?.description || fallback
}
