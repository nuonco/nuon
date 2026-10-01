import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useNotifications } from '@/hooks/use-notifications'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { resetFirstRunJourney } from '@/hooks/use-first-run-journey'
import { trackEvent } from '@/lib/posthog-analytics'
import type { TAPIError } from '@/types'
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

async function probeGrafanaHealth(): Promise<boolean> {
  try {
    const res = await fetch('/admin/grafana/api/health', {
      credentials: 'include',
      method: 'GET',
    })
    return res.ok
  } catch {
    return false
  }
}

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

  const { data: grafanaReachable = false } = useQuery({
    queryKey: ['admin', 'grafana', 'health'],
    queryFn: probeGrafanaHealth,
    enabled: !!grafanaUiUrl && !isLoading && !!user,
    staleTime: 5 * 60 * 1000,
    retry: false,
  })

  const reopenOnboarding = async () => {
    try {
      await resetFirstRunJourney()
      trackEvent({ event: 'onboarding_reopen', status: 'ok', user, props: {} })
    } catch (err) {
      trackEvent({
        event: 'onboarding_reopen',
        status: 'error',
        user,
        props: { err: (err as TAPIError)?.error },
      })
    } finally {
      window.location.assign('/onboarding?reopen=1')
    }
  }

  return (
    <UserDropdown
      isByoc={!!isByoc}
      isAdmin={!!isAdmin}
      isNuonEmployee={!!isNuonEmployee}
      isDev={!!isDev}
      apiUrl={apiUrl}
      adminDashboardUrl={adminDashboardUrl}
      grafanaUiUrl={grafanaReachable ? grafanaUiUrl : undefined}
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
