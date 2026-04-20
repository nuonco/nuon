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
  isNuonEmployee: boolean
  isDev: boolean
  apiUrl: string
  adminDashboardUrl?: string
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
  isNuonEmployee,
  isDev,
  apiUrl,
  adminDashboardUrl,
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
        {!hideOrgSettings && (isNuonEmployee || isDev) && (
          <Button
            onClick={() => onAddPanel(<AdminPanel />)}
            isMenuButton
          >
            Admin controls <Icon variant="Sliders" />
          </Button>
        )}
        {!hideOrgSettings && isAdmin && adminDashboardUrl && (
          <Link href={adminDashboardUrl} isExternal>
            Admin dashboard <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href="/admin/swagger" isExternal>
            Admin API <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href="/admin/temporal" isExternal>
            Temporal dashboard <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href="/public/swagger" isExternal>
            Public API <Icon variant="ArrowSquareOut" />
          </Link>
        )}
        {!hideOrgSettings && isAdmin && (
          <Link href={`${apiUrl}/httpbin`} isExternal>
            HTTPBin <Icon variant="ArrowSquareOut" />
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
