'use client'

import { AdminSection } from '../shared/AdminSection'
import { AdminActionGroup } from '../shared/AdminActionGroup'
import { AdminActionCard } from '../shared/AdminActionCard'
import { AdminMetadataPanel } from '../shared/AdminMetadata'
import { AdminTemporalLink } from '../../old/AdminTemporalLink'

// Import admin actions
import { reprovisionApp } from '@/actions/admin/reprovision-app'
import { restartApp } from '@/actions/admin/restart-app'

interface AdminAppSectionProps {
  orgId: string
  appId: string
}

export const AdminAppSection = ({ orgId, appId }: AdminAppSectionProps) => {
  const metadata = (
    <AdminMetadataPanel>
      <div className="space-y-1">
        <AdminTemporalLink namespace="apps" id={appId} />
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
          action={() => reprovisionApp(appId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will reprovision the app infrastructure. This may affect all installs of this app."
        />
        <AdminActionCard
          title="Restart app"
          description="Restart current app event loop"
          action={() => restartApp(appId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart the app event loop. Continue?"
        />
      </AdminActionGroup>
    </AdminSection>
  )
}