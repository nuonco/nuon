import { useEffect, useState } from 'react'
import type { TRunner } from '@/types'
import { ID } from '@/components/common/ID'
import { AdminSection } from '../shared/AdminSection'
import { AdminActionGroup } from '../shared/AdminActionGroup'
import { AdminActionCard } from '../shared/AdminActionCard'
import { AdminMetadataPanel, AdminInfoCard } from '../shared/AdminMetadata'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { AdminFeatureToggleCard } from '../shared/AdminFeatureToggleCard'
import { AdminRunnersCard } from '../shared/AdminRunnersCard'
import { useOrg } from '@/hooks/use-org'
import { useAuth } from '@/hooks/use-auth'
import {
  adminGetOrgRunner,
  adminAddSupportUsersToOrg,
  adminRemoveSupportUsersFromOrg,
  adminReprovisionOrg,
  adminRestartOrg,
  adminRestartOrgRunners,
  adminRestartRunner,
  adminGracefulRunnerShutdown,
  adminForceRunnerShutdown,
  adminInvalidateRunnerToken,
  adminEnableOrgDebugMode,
  adminDeprovisionOrg,
  adminForgetOrgInstalls,
} from '@/lib'

interface AdminOrgSectionProps {
  orgId: string
}

export const AdminOrgSection = ({ orgId }: AdminOrgSectionProps) => {
  const { org } = useOrg()
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''
  const [runner, setRunner] = useState<TRunner>()
  const [runnerLoading, setRunnerLoading] = useState(true)

  useEffect(() => {
    if (orgId) {
      adminGetOrgRunner({ orgId }).then((r) => {
        setRunner(r)
        setRunnerLoading(false)
      })
    }
  }, [orgId])

  const runnerId = runner?.id ?? ''

  const metadata = (
    <AdminMetadataPanel>
      <div className="flex justify-between items-start gap-4">
        <div className="flex flex-col gap-4">
          <AdminInfoCard
            title="Org Runner ID"
            value={runner?.id}
            copyable
            loading={runnerLoading}
          />
          <AdminRunnersCard orgId={orgId} />
        </div>
        <div className="flex-shrink-0">
          <TemporalLink namespace="orgs" eventLoopId={orgId} />
        </div>
      </div>
    </AdminMetadataPanel>
  )

  return (
    <AdminSection
      title="Organization controls"
      subtitle={
        <div className="flex gap-2">
          Managing org: <ID>{orgId}</ID>
        </div>
      }
      metadata={metadata}
    >
      <AdminActionGroup title="Org settings" icon="Users">
        <AdminActionCard
          title="Add support users"
          description="Add all Nuon support users to current org"
          action={() => adminAddSupportUsersToOrg({ orgId, adminEmail })}
        />
        <AdminActionCard
          title="Remove support users"
          description="Remove all Nuon support users from current org"
          action={() => adminRemoveSupportUsersFromOrg({ orgId, adminEmail })}
        />
        <AdminFeatureToggleCard org={org} orgId={orgId} />
      </AdminActionGroup>

      <AdminActionGroup
        title="Infrastructure"
        icon="HardDrives"
        variant="warning"
      >
        <AdminActionCard
          title="Reprovision org"
          description="Reprovision current org infrastructure"
          action={() => adminReprovisionOrg({ orgId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will reprovision the entire organization infrastructure. This may cause downtime."
        />
        <AdminActionCard
          title="Restart org"
          description="Restart current org event loop"
          action={() => adminRestartOrg({ orgId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart the organization event loop. Continue?"
        />
      </AdminActionGroup>

      <AdminActionGroup title="Runner control" icon="Play">
        <AdminActionCard
          title="Restart all runners"
          description="Restart all of current org runners"
          action={() => adminRestartOrgRunners({ orgId, adminEmail })}
          requiresConfirmation
          confirmationText="This will restart all runners for this organization. Continue?"
        />
        <AdminActionCard
          title="Restart runner"
          description="Restart the current org runner"
          action={() => adminRestartRunner({ runnerId, adminEmail })}
          requiresConfirmation
          confirmationText="This will restart the org runner. Continue?"
        />
        <AdminActionCard
          title="Graceful shutdown"
          description="Graceful shutdown of current org runner"
          action={() => adminGracefulRunnerShutdown({ runnerId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will gracefully shutdown the org runner. Continue?"
        />
        <AdminActionCard
          title="Force shutdown"
          description="Forceful shutdown of current org runner"
          action={() => adminForceRunnerShutdown({ runnerId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will forcefully shutdown the org runner and may cause data loss."
        />
      </AdminActionGroup>

      <AdminActionGroup title="Teardown & cleanup" icon="Trash" variant="danger">
        <AdminActionCard
          title="Deprovision org"
          description="Deprovision all org infrastructure. Keeps database records but tears down cloud resources."
          action={() => adminDeprovisionOrg({ orgId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          inputText="yesimsure"
          confirmationText="WARNING: This will deprovision ALL infrastructure for this organization including runners and installs. Database records will be preserved. This may take a while to complete."
        />
        <AdminActionCard
          title="Forget all org installs"
          description="Permanently forget all installs for this org. This cannot be undone."
          action={() => adminForgetOrgInstalls({ orgId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          inputText="yesimsure"
          confirmationText="WARNING: This will permanently delete ALL install records for this organization. Any running infrastructure will be orphaned and must be cleaned up manually (e.g. via aws-nuke). This action CANNOT be undone. Only use this when installs are broken beyond repair."
        />
      </AdminActionGroup>

      <AdminActionGroup title="Security & debug" icon="Shield" variant="danger">
        <AdminActionCard
          title="Invalidate runner token"
          description="Invalidate runner service account token"
          action={() => adminInvalidateRunnerToken({ runnerId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will invalidate the runner service account token. Live runners will no longer be able to connect to the API."
        />
        <AdminActionCard
          title="Enable debug mode"
          description="Enable debug mode for this org (logs all requests)"
          action={() => adminEnableOrgDebugMode({ orgId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will enable debug mode which logs all requests for this org."
        />
      </AdminActionGroup>
    </AdminSection>
  )
}
