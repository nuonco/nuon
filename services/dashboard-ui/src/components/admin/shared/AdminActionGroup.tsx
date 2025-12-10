'use client'

import { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { Card } from '@/components/common/Card'
import type { TIconVariant } from '@/components/common/Icon'

interface AdminActionGroupProps {
  title: string
  icon?: TIconVariant
  variant?: 'default' | 'warning' | 'danger'
  children: ReactNode
}

const getVariantClasses = (variant: AdminActionGroupProps['variant']) => {
  switch (variant) {
    case 'warning':
      return 'border-l-yellow-500'
    case 'danger':
      return 'border-l-red-500'
    default:
      return 'border-l-blue-500'
  }
}

const getIconColor = (variant: AdminActionGroupProps['variant']) => {
  switch (variant) {
    case 'warning':
      return 'text-yellow-600 dark:text-yellow-400'
    case 'danger':
      return 'text-red-600 dark:text-red-400'
    default:
      return 'text-blue-600 dark:text-blue-400'
  }
}

export const AdminActionGroup = ({
  title,
  icon,
  variant = 'default',
  children
}: AdminActionGroupProps) => {
  return (
    <Card className={`border-l-4 ${getVariantClasses(variant)}`}>
      <div className="space-y-4">
        {/* Group Header */}
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
        
        {/* Actions Grid */}
        <div className="grid gap-3 md:grid-cols-2">
          {children}
        </div>
      </div>
    </Card>
  )
}