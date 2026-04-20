import { ReactNode } from 'react'
import { Text } from '@/components/common/Text'

interface DevSectionProps {
  title: string
  subtitle?: string | ReactNode
  metadata?: ReactNode
  children: ReactNode
}

export const DevSection = ({
  title,
  subtitle,
  metadata,
  children,
}: DevSectionProps) => {
  return (
    <div className="space-y-6">
      <div className="flex flex-col space-y-2">
        <Text variant="h3" weight="strong">
          {title}
        </Text>
        {subtitle && (
          <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
            {subtitle}
          </Text>
        )}
        {metadata && <div className="flex gap-2 pt-2">{metadata}</div>}
      </div>

      <div className="space-y-6">{children}</div>
    </div>
  )
}
