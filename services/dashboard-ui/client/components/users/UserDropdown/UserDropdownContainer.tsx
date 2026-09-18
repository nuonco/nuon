import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useNotifications } from '@/hooks/use-notifications'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useDashboardPreferences } from '@/hooks/use-dashboard-preferences'
import { UserDropdown, type IUserDropdown } from './UserDropdown'

type IUserDropdownContainerProps = Omit<
  IUserDropdown,
  | 'isByoc'
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
  | 'statusBarEnabled'
  | 'onStatusBarEnabledChange'
  | 'installsTabEnabled'
  | 'onInstallsTabEnabledChange'
  | 'onRequestPermission'
  | 'onAddPanel'
  | 'onAddToast'
  | 'user'
  | 'isUserLoading'
>

export const UserDropdownContainer = (props: IUserDropdownContainerProps) => {
  const { isAdmin, isNuonEmployee, user, isLoading } = useAuth()
  const { apiUrl, authServiceUrl, adminDashboardUrl, isDev, isByoc } =
    useConfig()
  const { addPanel } = useSurfaces()
  const { addToast } = useToast()
  const { permission, requestPermission, isSupported, muted, toggleMute } =
    useNotifications()
  const {
    isInstallsTabEnabled,
    isStatusBarEnabled,
    setIsInstallsTabEnabled,
    setIsStatusBarEnabled,
  } = useDashboardPreferences()

  return (
    <UserDropdown
      isByoc={!!isByoc}
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
      statusBarEnabled={isStatusBarEnabled}
      onStatusBarEnabledChange={setIsStatusBarEnabled}
      installsTabEnabled={isInstallsTabEnabled}
      onInstallsTabEnabledChange={setIsInstallsTabEnabled}
      onRequestPermission={requestPermission}
      onAddPanel={addPanel}
      onAddToast={addToast}
      user={user}
      isUserLoading={isLoading}
      {...props}
    />
  )
}
