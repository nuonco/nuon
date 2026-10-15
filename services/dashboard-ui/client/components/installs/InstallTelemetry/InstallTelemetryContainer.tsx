import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallTelemetrySettings } from '@/lib/ctl-api/installs/get-install-telemetry-settings'
import { InstallTelemetry } from './InstallTelemetry'
import { InstallTelemetryModalContainer } from './InstallTelemetryModalContainer'

export const InstallTelemetryContainer = () => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const orgId = org?.id
  const installId = install?.id
  const settings = useQuery({
    queryKey: ['install-telemetry', orgId, installId],
    queryFn: () =>
      getInstallTelemetrySettings({ orgId: orgId!, installId: installId! }),
    enabled: !!orgId && !!installId,
    retry: false,
  })
  if (!orgId || !installId) return null
  return (
    <InstallTelemetry
      settings={settings.data}
      isLoading={settings.isPending}
      error={settings.error}
      onManage={() =>
        addModal(
          <InstallTelemetryModalContainer install={install} orgId={orgId} />
        )
      }
      onRetry={() => {
        void settings.refetch()
      }}
    />
  )
}
