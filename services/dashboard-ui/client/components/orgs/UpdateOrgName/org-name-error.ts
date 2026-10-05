import type { TAPIError } from '@/types'

const DUPLICATE_NAME = /duplicate key|idx_org_name|23505/i

export const orgNameSubmitError = (
  error: TAPIError | null
): TAPIError | null => {
  if (!error) return null
  const text = `${error.error ?? ''} ${error.description ?? ''}`
  if (error.status === 409 || DUPLICATE_NAME.test(text)) {
    return {
      ...error,
      error: 'An organization with this name already exists.',
      description: 'Choose a different name.',
    }
  }
  return error
}
