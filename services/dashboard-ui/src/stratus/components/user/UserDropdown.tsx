'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { Bug, SignOut, UserPlus } from '@phosphor-icons/react'
import { useOrg } from '@/stratus/context'
import {
  Button,
  Dropdown,
  Link,
  Menu,
  Text,
  type IDropdown,
} from '@/stratus/components/common'
import { Profile } from './Profile'

export interface IUserDropdown
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id' | 'variant'> {}

export const UserDropdown: FC<IUserDropdown> = ({
  buttonClassName,
  ...props
}) => {
  const { org } = useOrg()
  return (
    <Dropdown
      buttonClassName={classNames('text-left !px-px !py-px', {
        [`${buttonClassName}`]: Boolean(buttonClassName),
      })}
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
          Invite team member <UserPlus />
        </Button>
        <Link href="/settings">
          Report bug <Bug />
        </Link>
        <hr />
        <Link
          href="/api/auth/logout"
          className="!text-red-800 dark:!text-red-500"
          title="Sign out"
        >
          Log out <SignOut />
        </Link>
      </Menu>
    </Dropdown>
  )
}
