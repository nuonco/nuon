'use client'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminRunnersPanel } from './AdminRunnersPanel'

interface AdminRunnersCardProps {
  orgId: string
}

export const AdminRunnersCard = ({ orgId }: AdminRunnersCardProps) => {
  const { addPanel } = useSurfaces()

  const handleOpenPanel = () => {
    const panel = (
      <AdminRunnersPanel
        orgId={orgId}
      />
    )
    addPanel(panel)
  }

  return (
    <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
      <div className="flex flex-col">
        <Text variant="base" weight="strong">
          Manage all runners
        </Text>
        <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
          View and control all runners in this organization
        </Text>
      </div>
      
      <Button
        onClick={handleOpenPanel}
        variant="secondary"
      >
        <Icon variant="SlidersHorizontal" />
        Manage runners
      </Button>
    </div>
  )
}
