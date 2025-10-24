import classNames from 'classnames'
import React, { type FC } from 'react'

export const LogLineSeverity: FC<{ severity_number: number }> = ({
  severity_number,
}) => {
  return (
    <span
      className={classNames('flex w-0.5 h-3', {
        'bg-primary-400 dark:bg-primary-300': severity_number <= 4,
        'bg-cool-grey-600 dark:bg-cool-grey-500':
          severity_number >= 5 && severity_number <= 8,
        'bg-blue-600 dark:bg-blue-500':
          severity_number >= 9 && severity_number <= 12,
        'bg-orange-600 dark:bg-orange-500':
          severity_number >= 13 && severity_number <= 16,
        'bg-red-600 dark:bg-red-500':
          severity_number >= 17 && severity_number <= 20,
        'bg-red-700 dark:bg-red-600':
          severity_number >= 21 && severity_number <= 24,
      })}
    />
  )
}
