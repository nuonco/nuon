import { AdminSection } from '../../shared/AdminSection'
import { AdminActionGroup } from '../../shared/AdminActionGroup'
import { AdminActionCard } from '../../shared/AdminActionCard'
import { AdminMetadataPanel } from '../../shared/AdminMetadata'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { adminReprovisionApp, adminRestartApp } from '@/lib'

interface IAdminAppSection {
  orgId: string
  appId: string
  adminEmail: string
}

export const AdminAppSection = ({ appId, adminEmail }: IAdminAppSection) => {
  const metadata = (
    <AdminMetadataPanel>
      <div className="space-y-1">
        <TemporalLink namespace="apps" eventLoopId={appId} />
      </div>
    </AdminMetadataPanel>
  )

  return (
    <AdminSection
      title="Application controls"
      subtitle={`Managing app: ${appId}`}
      metadata={metadata}
    >
      <AdminActionGroup title="App infrastructure" icon="Package" variant="warning">
        <AdminActionCard
          title="Reprovision app"
          description="Reprovision current app infrastructure"
          action={() => adminReprovisionApp({ appId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will reprovision the app infrastructure. This may affect all installs of this app."
        />
        <AdminActionCard
          title="Restart app"
          description="Restart current app event loop"
          action={() => adminRestartApp({ appId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart the app event loop. Continue?"
        />
      </AdminActionGroup>
    </AdminSection>
  )
}
