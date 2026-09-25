import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'

export interface IBranchManagementDropdown {
  dropdownId: string
  detailHref: string
  editButton: ReactNode
  deploymentPlanButton: ReactNode
  hasConfig: boolean
  isTriggerPending: boolean
  onTriggerRun: () => void
}

export const BranchManagementDropdown = ({
  dropdownId,
  detailHref,
  editButton,
  deploymentPlanButton,
  hasConfig,
  isTriggerPending,
  onTriggerRun,
}: IBranchManagementDropdown) => {
  return (
    <Dropdown
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      id={dropdownId}
      variant="ghost"
    >
      <Menu>
        <Button href={detailHref}>View details</Button>
        <hr />
        <Text variant="label" theme="neutral">
          Settings
        </Text>
        {editButton}
        {deploymentPlanButton}
        <hr />
        <Text variant="label" theme="neutral">
          Controls
        </Text>
        <Button
          isMenuButton
          onClick={onTriggerRun}
          disabled={!hasConfig || isTriggerPending}
          tooltipProps={
            !hasConfig
              ? {
                  className: 'block !w-full',
                  position: 'left',
                  tipContent: 'Sync the app config first',
                }
              : undefined
          }
        >
          Trigger run
          <Icon variant="PlayIcon" />
        </Button>
      </Menu>
    </Dropdown>
  )
}
