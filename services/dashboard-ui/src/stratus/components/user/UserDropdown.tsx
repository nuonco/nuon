'use client'

import React from 'react'
import { useOrg } from '@/stratus/context'
import { cn } from '@/stratus/components/helpers'
import {
  Button,
  Dropdown,
  Icon,
  Link,
  Menu,
  Text,
  type IDropdown,
} from '@/stratus/components/common'
import { Profile } from './Profile'

export interface IUserDropdown
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id' | 'variant'> {}

export const UserDropdown = ({ buttonClassName, ...props }: IUserDropdown) => {
  const { org } = useOrg()
  return (
    <Dropdown
      buttonClassName={cn('text-left !px-px !py-px', buttonClassName)}
      buttonText={<Profile />}
      id="profile"
      variant="ghost"
      {...props}
    >
      <Menu className="min-w-56">
        <Text variant="label" theme="muted">
          {org?.name} settings
        </Text>
        <Button>
          Invite team member <Icon variant="UserPlus" />
        </Button>
        <Link href="/settings">
          Report bug <Icon variant="Bug" />
        </Link>
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
