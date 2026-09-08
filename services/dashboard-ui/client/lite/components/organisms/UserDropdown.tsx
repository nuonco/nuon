import type { ReactNode } from 'react'
import type { TOrg } from '@/types'
import { cn } from '@/utils/classnames'
import { Button } from '../atoms/Button'
import { Dropdown, type IDropdown } from '../atoms/Dropdown'
import { Icon } from '../atoms/Icon'
import {
  Menu,
  MenuItem,
  MenuSeparator,
  MenuSubmenu,
} from '../molecules/Menu'
import { OrgProfile } from '../molecules/OrgProfile'
import { UserProfile, type IUserProfileData } from '../molecules/UserProfile'

export interface IUserDropdown extends Omit<IDropdown, 'children' | 'trigger'> {
  user?: IUserProfileData | null
  loading?: boolean
  compact?: boolean
  signOutHref: string
  triggerClassName?: string
  orgSwitcher?: ReactNode
  org?: TOrg | null
  orgLoading?: boolean
}

export const UserDropdown = ({
  user,
  loading = false,
  compact = false,
  signOutHref,
  triggerClassName,
  orgSwitcher,
  org,
  orgLoading = false,
  align = 'end',
  matchTriggerWidth,
  stretch = false,
  ...props
}: IUserDropdown) => (
  <Dropdown
    align={align}
    matchTriggerWidth={matchTriggerWidth ?? !compact}
    stretch={stretch}
    trigger={
      <Button
        variant="ghost"
        iconOnly={compact}
        aria-label={compact ? 'Open user menu' : undefined}
        className={cn(stretch && 'w-full', triggerClassName)}
      >
        <UserProfile
          user={user}
          loading={loading}
          compact={compact}
          avatarSize={compact ? 'sm' : 'md'}
        />
      </Button>
    }
    {...props}
  >
    <Menu className="w-72">
      {orgSwitcher ? (
        <>
          <MenuSubmenu
            label={
              <OrgProfile org={org} loading={orgLoading} avatarSize="sm" />
            }
          >
            {orgSwitcher}
          </MenuSubmenu>
          <MenuSeparator />
        </>
      ) : null}
      <MenuItem
        href={signOutHref}
        external
        target="_self"
        tone="danger"
        icon={<Icon variant="SignOutIcon" />}
      >
        Sign out
      </MenuItem>
    </Menu>
  </Dropdown>
)
