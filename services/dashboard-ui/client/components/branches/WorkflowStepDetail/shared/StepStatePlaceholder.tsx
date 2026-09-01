import type { ReactNode } from 'react'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'

interface IStepStatePlaceholder {
  variant?: 'loading' | 'pending'
  children: ReactNode
}

export const StepStatePlaceholder = ({ variant = 'loading', children }: IStepStatePlaceholder) => (
  <div className="flex items-center gap-3">
    {variant === 'loading' ? (
      <Loading size={16} className="text-cool-grey-400 dark:text-cool-grey-500 shrink-0" />
    ) : (
      <Icon
        variant="ClockIcon"
        size={16}
        className="text-cool-grey-400 dark:text-cool-grey-500 shrink-0"
      />
    )}
    <Text variant="subtext" theme="neutral">
      {children}
    </Text>
  </div>
)
