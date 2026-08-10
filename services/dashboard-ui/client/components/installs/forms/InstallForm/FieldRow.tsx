import type { ReactNode } from 'react'
import { Text } from '@/components/common/Text'

export interface IFieldRow {
  labelText: ReactNode
  helpText?: ReactNode
  required?: boolean
  optional?: boolean
  children: ReactNode
}

export const FieldRow = ({
  labelText,
  helpText,
  required,
  optional,
  children,
}: IFieldRow) => (
  <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
    <span className="flex flex-col gap-0">
      <Text variant="body" weight="strong">
        {labelText}{' '}
        {required ? (
          <Text className="ml-1" variant="subtext" theme="error">
            *
          </Text>
        ) : null}
        {optional ? (
          <Text className="ml-1" variant="subtext" theme="neutral">
            (optional)
          </Text>
        ) : null}
      </Text>
      {helpText ? (
        <Text variant="subtext" className="max-w-72">
          {helpText}
        </Text>
      ) : null}
    </span>
    {children}
  </label>
)
