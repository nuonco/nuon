import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { AdminPanel } from '@/components/admin/AdminPanel'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { InviteUserButton } from '@/components/team/InviteUser'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import { UserProfile } from './UserProfile'
import { Button } from '@/components/common/Button'

export interface IUserDropdown
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id' | 'variant'> {
  hideOrgSettings?: boolean
}

export const UserDropdown = ({
  buttonClassName,
  hideOrgSettings,
  ...props
}: IUserDropdown) => {
  const { isAdmin } = useAuth()
  const { authServiceUrl } = useConfig()
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
        {!hideOrgSettings && (
          <Text variant="label" theme="neutral">
            Org settings
          </Text>
        )}
        {!hideOrgSettings && <InviteUserButton isMenuButton />}
        {!hideOrgSettings && (
          <Link href="/onboarding">
            Re-open onboarding <Icon variant="Signpost" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin ? (
          <Button
            onClick={() => {
              addPanel(<AdminPanel />)
            }}
          >
            Admin panel <Icon variant="Sliders" />
          </Button>
        ) : null}
        {!hideOrgSettings && <hr />}
        <Link
          href={`${authServiceUrl}/logout`}
          className="!text-red-800 dark:!text-red-500"
          title="Sign out"
          isExternal
          target="_self"
        >
          Sign out <Icon variant="SignOut" />
        </Link>
      </Menu>
    </Dropdown>
  )
}
