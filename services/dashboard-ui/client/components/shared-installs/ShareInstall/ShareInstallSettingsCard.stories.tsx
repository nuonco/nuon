export default {
  title: 'Playground / Customer portal / Settings entry',
}

import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { mockShares } from '../fixtures'
import { ShareInstallButton } from './ShareInstallButton'
import { SharedWithList } from '../SharedWithList/SharedWithList'

const noop = () => {}

const ActionCard = ({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) => (
  <Card className="!p-4 !gap-4 max-w-md">
    <HeadingGroup className="gap-1">
      <Text weight="strong">{title}</Text>
      <Text variant="subtext" theme="neutral">
        {description}
      </Text>
    </HeadingGroup>
    <div className="mt-auto">{children}</div>
  </Card>
)

export const SettingsCard = () => (
  <div className="p-6 flex flex-col gap-6 max-w-3xl">
    <ActionCard
      title="Share"
      description="Invite a customer to view this install in the portal."
    >
      <ShareInstallButton
        installName="Acme production"
        shares={mockShares.filter((share) => share.install_id === 'inst-1')}
        onSubmit={noop}
        onRevoke={noop}
        onResend={noop}
      />
    </ActionCard>

    <div className="flex flex-col gap-3">
      <Text variant="base" weight="strong">
        Shared with
      </Text>
      <SharedWithList
        shares={mockShares.filter((share) => share.install_id === 'inst-1')}
        onRevoke={noop}
        onResend={noop}
      />
    </div>
  </div>
)
