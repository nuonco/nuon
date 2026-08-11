import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import type { TAPIError } from '@/types'

export interface IFormErrorBanner {
  error: TAPIError | null | undefined
  fallback: string
}

export const FormErrorBanner = ({ error, fallback }: IFormErrorBanner) =>
  error ? (
    <Banner theme="error">
      <span className="flex flex-col gap-1">
        <Text>{error?.error || fallback}</Text>
        {error?.description ? (
          <Text variant="subtext">{error.description}</Text>
        ) : null}
      </span>
    </Banner>
  ) : null
