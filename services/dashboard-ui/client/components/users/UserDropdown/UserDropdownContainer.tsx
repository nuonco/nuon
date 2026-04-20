import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useNotifications } from '@/hooks/use-notifications'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { UserDropdown, type IUserDropdown } from './UserDropdown'

type IUserDropdownContainerProps = Omit<
  IUserDropdown,
  | 'isAdmin'
  | 'isNuonEmployee'
  | 'isDev'
  | 'apiUrl'
  | 'adminDashboardUrl'
  | 'authServiceUrl'
  | 'notificationsSupported'
  | 'notificationPermission'
  | 'muted'
  | 'onToggleMute'
  | 'onRequestPermission'
  | 'onAddPanel'
  | 'onAddToast'
  | 'user'
  | 'isUserLoading'
>

export const UserDropdownContainer = (props: IUserDropdownContainerProps) => {
  const { isAdmin, isNuonEmployee, user, isLoading } = useAuth()
  const { apiUrl, authServiceUrl, adminDashboardUrl, isDev } = useConfig()
  const { addPanel } = useSurfaces()
  const { addToast } = useToast()
  const { permission, requestPermission, isSupported, muted, toggleMute } = useNotifications()

  return (
    <UserDropdown
      isAdmin={!!isAdmin}
      isNuonEmployee={!!isNuonEmployee}
      isDev={!!isDev}
      apiUrl={apiUrl}
      adminDashboardUrl={adminDashboardUrl}
      authServiceUrl={authServiceUrl}
      notificationsSupported={isSupported}
      notificationPermission={permission ?? ''}
      muted={muted}
      onToggleMute={toggleMute}
      onRequestPermission={requestPermission}
      onAddPanel={addPanel}
      onAddToast={addToast}
      user={user}
      isUserLoading={isLoading}
      {...props}
    />
  )
}
