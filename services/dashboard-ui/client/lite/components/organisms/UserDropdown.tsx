import { cn } from '@/utils/classnames'
import { Button } from '../atoms/Button'
import { Dropdown, type IDropdown } from '../atoms/Dropdown'
import { Icon } from '../atoms/Icon'
import { Menu, MenuItem } from '../molecules/Menu'
import { UserProfile, type IUserProfileData } from '../molecules/UserProfile'

export interface IUserDropdown extends Omit<IDropdown, 'children' | 'trigger'> {
  user?: IUserProfileData | null
  loading?: boolean
  compact?: boolean
  signOutHref: string
  triggerClassName?: string
}

export const UserDropdown = ({
  user,
  loading = false,
  compact = false,
  signOutHref,
  triggerClassName,
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
    <Menu>
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
