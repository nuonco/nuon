'use client'

import { useUser } from '@auth0/nextjs-auth0'
import { AdminPanel } from '@/components/admin/AdminPanel'
import { Button } from '@/components/common/Button'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'

import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import { UserProfile } from './UserProfile'

import { InvitePanel } from '../OrgInviteModal'

export interface IUserDropdown
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id' | 'variant'> {}

export const UserDropdown = ({ buttonClassName, ...props }: IUserDropdown) => {
  const { user } = useUser()
  const { addPanel } = useSurfaces()

  return (
    <Dropdown
      buttonClassName={cn('text-left !px-px !py-px', buttonClassName)}
      buttonText={<UserProfile />}
      id="profile"
      variant="ghost"
      {...props}
    >
      <Menu className="min-w-56">
        <Text variant="label" theme="neutral">
          Org settings
        </Text>
        <Button
          onClick={() => {
            addPanel(<InvitePanel />)
          }}
        >
          Invite team member <Icon variant="UserPlus" />
        </Button>
        {/* <Link href="/settings">
            Report bug <Icon variant="Bug" />
            </Link> */}
        {user?.email?.endsWith('@nuon.co') ? (
          <Button
            onClick={() => {
              addPanel(<AdminPanel />)
            }}
          >
            Admin panel <Icon variant="Sliders" />
          </Button>
        ) : null}
        <hr />
        <Link
          href="/api/auth/logout"
          className="!text-red-800 dark:!text-red-500"
          title="Sign out"
        >
          Log out <Icon variant="SignOut" />
        </Link>
      </Menu>
    </Dropdown>
  )
}
