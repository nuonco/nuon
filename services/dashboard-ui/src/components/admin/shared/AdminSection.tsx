'use client'

import { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'

interface AdminSectionProps {
  title: string
  subtitle?: string | ReactNode
  metadata?: ReactNode
  children: ReactNode
}

export const AdminSection = ({
  title,
  subtitle,
  metadata,
  children
}: AdminSectionProps) => {
  return (
    <div className="space-y-6">
      {/* Section Header */}
      <div className="flex flex-col space-y-2">
        <Text variant="h3" weight="strong">{title}</Text>
        {subtitle && (
          <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
            {subtitle}
          </Text>
        )}
        {metadata && (
          <div className="flex gap-2 pt-2">
            {metadata}
          </div>
        )}
      </div>
      
      {/* Section Content */}
      <div className="space-y-6">
        {children}
      </div>
    </div>
  )
}
