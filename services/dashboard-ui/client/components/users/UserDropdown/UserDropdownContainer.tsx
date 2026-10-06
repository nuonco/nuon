import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useNotifications } from '@/hooks/use-notifications'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/posthog-analytics'
import { UserDropdown, type IUserDropdown } from './UserDropdown'

type IUserDropdownContainerProps = Omit<
  IUserDropdown,
  | 'isByoc'
  | 'isAdmin'
  | 'isNuonEmployee'
  | 'isDev'
  | 'apiUrl'
  | 'adminDashboardUrl'
  | 'grafanaUiUrl'
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
  | 'onboardingFirstRun'
  | 'onReopenOnboarding'
>

export const UserDropdownContainer = (props: IUserDropdownContainerProps) => {
  const { isAdmin, isNuonEmployee, user, isLoading } = useAuth()
  const {
    apiUrl,
    authServiceUrl,
    adminDashboardUrl,
    grafanaUiUrl,
    isDev,
    isByoc,
    onboardingFirstRun,
  } = useConfig()
  const { addPanel } = useSurfaces()
  const { addToast } = useToast()
  const { permission, requestPermission, isSupported, muted, toggleMute } =
    useNotifications()
  const reopenOnboarding = () => {
    trackEvent({ event: 'onboarding_reopen', status: 'ok', user, props: {} })
    window.location.assign('/onboarding')
  }

  return (
    <UserDropdown
      isByoc={!!isByoc}
      isAdmin={!!isAdmin}
      isNuonEmployee={!!isNuonEmployee}
      isDev={!!isDev}
      apiUrl={apiUrl}
      adminDashboardUrl={adminDashboardUrl}
      grafanaUiUrl={grafanaUiUrl}
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
      onboardingFirstRun={onboardingFirstRun}
      onReopenOnboarding={onboardingFirstRun ? reopenOnboarding : undefined}
      {...props}
    />
  )
}
