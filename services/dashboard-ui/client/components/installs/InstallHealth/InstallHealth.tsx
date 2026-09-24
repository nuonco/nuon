import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'

export interface IInstallHealth {
  resources: ReactNode
  timeline: ReactNode
}

export const InstallHealth = ({ resources, timeline }: IInstallHealth) => (
  <div className="flex flex-col gap-6">
    <Card>{timeline}</Card>
    <div className="flex flex-col gap-4">
      <Text variant="body" weight="strong">
        Resource health
      </Text>
      {resources}
    </div>
  </div>
)
