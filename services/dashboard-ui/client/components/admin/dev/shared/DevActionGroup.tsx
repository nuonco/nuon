import { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { Card } from '@/components/common/Card'
import type { TIconVariant } from '@/components/common/Icon'

interface DevActionGroupProps {
  title: string
  icon?: TIconVariant
  variant?: 'default' | 'warning' | 'danger'
  children: ReactNode
}

const getVariantClasses = (variant: DevActionGroupProps['variant']) => {
  switch (variant) {
    case 'warning':
      return 'border-l-amber-500'
    case 'danger':
      return 'border-l-rose-500'
    default:
      return 'border-l-teal-500'
  }
}

const getIconColor = (variant: DevActionGroupProps['variant']) => {
  switch (variant) {
    case 'warning':
      return 'text-amber-600 dark:text-amber-400'
    case 'danger':
      return 'text-rose-600 dark:text-rose-400'
    default:
      return 'text-teal-600 dark:text-teal-400'
  }
}

export const DevActionGroup = ({
  title,
  icon,
  variant = 'default',
  children,
}: DevActionGroupProps) => {
  return (
    <Card className={`border-l-4 ${getVariantClasses(variant)}`}>
      <div className="space-y-4">
        <div className="flex flex-col">
          <div className="flex items-center gap-3">
            {icon && (
              <Icon
                variant={icon}
                size="20"
                className={getIconColor(variant)}
              />
            )}
            <Text variant="base" weight="strong">
              {title}
            </Text>
          </div>
        </div>

        <div className="grid gap-3 md:grid-cols-2">{children}</div>
      </div>
    </Card>
  )
}
