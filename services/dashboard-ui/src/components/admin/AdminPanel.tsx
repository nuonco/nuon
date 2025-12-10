'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { AdminControls } from './AdminControls'

export const AdminPanel = ({ size = '3/4', ...props }: IPanel) => {
  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="Gear" size="24" />
          <Text weight="strong" variant="h2">
            Admin Controls
          </Text>
        </div>
      }
      size={size}
      {...props}
    >
      <AdminControls />
    </Panel>
  )
}
