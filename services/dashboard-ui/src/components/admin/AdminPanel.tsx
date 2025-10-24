'use client'

import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'

// NOTE: Old admin controls
import { AdminControls } from '@/components/old/AdminModal'

export const AdminPanel = ({ size = 'full', ...props }: IPanel) => {
  return (
    <Panel
      heading={
        <Text weight="strong" variant="h2">
          Admin panel
        </Text>
      }
      size={size}
      {...props}
    >
      <AdminControls />
    </Panel>
  )
}
