import React from 'react'
import { UserDropdown } from '@/components/users/UserDropdown'
import { SpotlightTrigger } from '@/components/spotlight/SpotlightTrigger'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Tooltip } from '@/components/common/Tooltip'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { cn } from '@/utils/classnames'
import { MainSidebarButton } from './MainSidebarButton'

export interface IMainTopbar extends React.HTMLAttributes<HTMLDivElement> {
  hideSidebarButtons?: boolean
  hideOrgSettings?: boolean
}

export const MainTopbar = ({
  className,
  children,
  hideSidebarButtons = false,
  hideOrgSettings = false,
  ...props
}: IMainTopbar) => {
  const { org } = useOrg()

  return (
    <header
      className={cn(
        'py-3 px-4 border-b flex shrink-0 items-center h-[60px] w-full overflow-x-auto md:overflow-visible',
        className
      )}
      {...props}
    >
      <div className="flex items-center gap-2 w-full">
        {hideSidebarButtons ? null : (
          <>
            <div className="md:hidden">
              <MainSidebarButton variant="mobile" />
            </div>
            <div className="hidden md:block">
              <MainSidebarButton />
            </div>
          </>
        )}
        {children}

        <div className="hidden lg:flex items-center gap-4 ml-auto shrink-0">
          {org?.sandbox_mode && (
            <Tooltip
              tipContentClassName="max-w-64"
              tipContent={
                <Text variant="subtext">
                  This organization is running in sandbox mode. Installs are simulated instead of deploying to a real cloud account.
                </Text>
              }
              position="bottom"
            >
              <Badge variant="code" theme="neutral" size="sm" className="shrink-0">
                <Icon variant="TestTubeIcon" size={14} className="xl:hidden" />
                <span className="hidden xl:inline whitespace-nowrap">Sandbox mode</span>
              </Badge>
            </Tooltip>
          )}
          <SpotlightTrigger />
          <UserDropdown alignment="right" hideOrgSettings={hideOrgSettings} />
        </div>
      </div>
    </header>
  )
}
