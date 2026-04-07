import { AdminPanel } from '@/components/admin/AdminPanel'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { InviteUserButton } from '@/components/team/InviteUser'
import { cn } from '@/utils/classnames'
import { UserProfile } from '../UserProfile/UserProfile'
import { Button } from '@/components/common/Button'

export interface IUserDropdown
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id' | 'variant'> {
  hideOrgSettings?: boolean
  isAdmin: boolean
  authServiceUrl: string
  notificationsSupported: boolean
  notificationPermission: string
  muted: boolean
  onToggleMute: () => void
  onRequestPermission: () => Promise<string>
  onAddPanel: (panel: React.ReactElement) => void
  onAddToast: (toast: React.ReactElement) => void
  user?: { name?: string; email?: string; picture?: string } | null
  isUserLoading: boolean
}

export const UserDropdown = ({
  buttonClassName,
  hideOrgSettings,
  isAdmin,
  authServiceUrl,
  notificationsSupported,
  notificationPermission,
  muted,
  onToggleMute,
  onRequestPermission,
  onAddPanel,
  onAddToast,
  user,
  isUserLoading,
  ...props
}: IUserDropdown) => {
  return (
    <Dropdown
      buttonClassName={cn('text-left !px-px !py-px', buttonClassName)}
      buttonText={<UserProfile isLoading={isUserLoading} user={user} />}
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
        {!hideOrgSettings && <hr />}
        {!hideOrgSettings && isAdmin && (
          <Text variant="label" theme="neutral">
            Admin
          </Text>
        )}
        {!hideOrgSettings && isAdmin && (
          <Button
            onClick={() => onAddPanel(<AdminPanel />)}
            isMenuButton
          >
            Admin panel <Icon variant="Sliders" />
          </Button>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href="/admin/temporal" isExternal>
            Temporal UI <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href="/admin/swagger" isExternal>
            Admin API swagger <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href="/public/swagger" isExternal>
            API swagger <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && <hr />}
        {!hideOrgSettings && (
          <Text variant="label" theme="neutral">
            User settings
          </Text>
        )}
        {notificationsSupported && notificationPermission === 'granted' ? (
          <Button onClick={() => {
            onToggleMute()
            onAddToast(
              <Toast heading={muted ? 'Notifications enabled' : 'Notifications disabled'}>
                <Text>{muted ? 'You will receive desktop notifications.' : 'Desktop notifications are now muted.'}</Text>
              </Toast>
            )
          }}>
            {muted ? 'Enable' : 'Disable'} notifications <Icon variant={muted ? 'Bell' : 'BellSlash'} />
          </Button>
        ) : notificationsSupported && notificationPermission !== 'denied' ? (
          <Button onClick={async () => {
            const result = await onRequestPermission()
            if (result === 'granted') {
              onAddToast(
                <Toast heading="Notifications enabled">
                  <Text>You will receive desktop notifications.</Text>
                </Toast>
              )
            }
          }}>
            Enable notifications <Icon variant="Bell" />
          </Button>
        ) : null}
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
