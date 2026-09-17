import type { TAPIError } from '@/types'
import { Banner } from '../atoms/Banner'

export interface IFormErrorBanner {
  error: TAPIError | Error | null | undefined
  fallback: string
}

export const FormErrorBanner = ({ error, fallback }: IFormErrorBanner) => {
  if (!error) return null

  const apiError = 'error' in error ? error : undefined
  const heading =
    apiError?.error || ('message' in error ? error.message : '') || fallback

  return (
    <Banner theme="error" heading={heading}>
      {apiError?.description || undefined}
    </Banner>
  )
}
