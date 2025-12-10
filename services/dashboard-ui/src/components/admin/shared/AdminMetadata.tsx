'use client'

import { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Card } from '@/components/common/Card'

interface AdminInfoCardProps {
  title: string
  value: string | undefined | null
  copyable?: boolean
  loading?: boolean
}

interface AdminMetadataPanelProps {
  children: ReactNode
}

export const AdminInfoCard = ({
  title,
  value,
  copyable = false,
  loading = false,
}: AdminInfoCardProps) => {
  return (
    <div className="flex flex-col space-y-1">
      <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
        {title}
      </Text>
      {loading ? (
        <div className="h-5 w-32 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
      ) : value ? (
        copyable ? (
          <ClickToCopy className="font-mono text-sm">{value}</ClickToCopy>
        ) : (
          <Text variant="base" className="font-mono text-sm">
            {value}
          </Text>
        )
      ) : (
        <Text variant="subtext" className="italic">
          Not available
        </Text>
      )}
    </div>
  )
}

export const AdminMetadataPanel = ({ children }: AdminMetadataPanelProps) => {
  return <Card className="w-full">{children}</Card>
}
