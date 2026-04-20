import React from 'react'
import { UserDropdown } from '@/components/users/UserDropdown'
import { SpotlightTrigger } from '@/components/spotlight/SpotlightTrigger'
import { Badge } from '@/components/common/Badge'
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

        <div className="hidden md:flex items-center gap-4 ml-auto">
          {org?.sandbox_mode && (
            <Tooltip
              tipContent={
                <Text variant="subtext">
                  This organization is running in sandbox mode. Installs are simulated instead of deploying to a real cloud account.
                </Text>
              }
              position="bottom"
            >
              <Badge variant="code" theme="neutral">
                Sandbox mode
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
