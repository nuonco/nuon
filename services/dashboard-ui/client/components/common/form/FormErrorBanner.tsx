import { Banner } from '@/components/common/Banner'
import type { TAPIError } from '@/types'

export interface IFormErrorBanner {
  error: TAPIError | null | undefined
  fallback: string
}

export const FormErrorBanner = ({ error, fallback }: IFormErrorBanner) =>
  error ? <Banner theme="error">{error?.error || fallback}</Banner> : null
