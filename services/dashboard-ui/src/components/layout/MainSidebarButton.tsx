'use client'

import { useSidebar } from '@/hooks/use-sidebar'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'

export interface IMainSidebarButton
  extends Omit<
    IButtonAsButton,
    'variant' | 'children' | 'title' | 'aria-label'
  > {
  variant?: 'mobile' | 'default'
}

export const MainSidebarButton = ({
  variant = 'default',
}: IMainSidebarButton) => {
  const { isSidebarOpen, toggleSidebar } = useSidebar()

  // Dynamic labels based on sidebar state
  const ariaLabel = isSidebarOpen ? 'Close sidebar' : 'Open sidebar'

  if (variant === 'mobile') {
    return (
      <Button
        variant="ghost"
        className="!px-2"
        onClick={toggleSidebar}
        aria-label={ariaLabel}
      >
        {isSidebarOpen ? (
          <Icon variant="ArrowLineLeft" aria-hidden="true" />
        ) : (
          <Icon variant="ArrowLineRight" aria-hidden="true" />
        )}
      </Button>
    )
  }

  // default (desktop) variant
  return (
    <Tooltip
      position="bottom"
      tipContent={
        <div className="w-fit">
          <Text
            className="inline-flex gap-2 items-center !text-nowrap"
            variant="subtext"
          >
            {isSidebarOpen ? 'Collapse' : 'Expand'} sidebar
            <span className="inline-flex gap-0.5">
              <Badge variant="code" size="sm">
                ALT
              </Badge>
              <Badge variant="code" size="sm">
                S
              </Badge>
            </span>
          </Text>
        </div>
      }
    >
      <Button
        variant="ghost"
        className="!py-1 !px-1.5"
        onClick={toggleSidebar}
        aria-label={ariaLabel}
      >
        <Icon variant="SidebarSimple" size="20" aria-hidden="true" />
      </Button>
    </Tooltip>
  )
}
