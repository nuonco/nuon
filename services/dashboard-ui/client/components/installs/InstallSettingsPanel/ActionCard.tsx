import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'

export const ActionCard = ({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) => (
  <Card className="!p-4 !gap-4">
    <HeadingGroup className="gap-1">
      <Text weight="strong">{title}</Text>
      <Text variant="subtext" theme="neutral">
        {description}
      </Text>
    </HeadingGroup>
    <div className="mt-auto">{children}</div>
  </Card>
)
