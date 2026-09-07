import type { TAPIError } from '@/types'
import { Card } from '../atoms/Card'
import { Icon } from '../atoms/Icon'
import { Text } from '../atoms/Text'

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
    <Card
      role="alert"
      padding="sm"
      className="flex gap-2.5 bg-diff-remove-section"
    >
      <Icon
        variant="WarningIcon"
        size={16}
        className="mt-0.5 shrink-0 text-field-invalid"
      />
      <span className="flex min-w-0 flex-col gap-1">
        <Text variant="body" weight="medium">
          {heading}
        </Text>
        {apiError?.description ? (
          <Text variant="caption" color="secondary">
            {apiError.description}
          </Text>
        ) : null}
      </span>
    </Card>
  )
}
