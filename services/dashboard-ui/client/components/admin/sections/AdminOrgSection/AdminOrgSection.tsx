import type { TOrg, TRunner } from '@/types'
import { ID } from '@/components/common/ID'
import { AdminSection } from '../../shared/AdminSection'
import { AdminActionGroup } from '../../shared/AdminActionGroup'
import { AdminActionCard } from '../../shared/AdminActionCard'
import { AdminMetadataPanel, AdminInfoCard } from '../../shared/AdminMetadata'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { AdminFeatureToggleCard } from '../../shared/AdminFeatureToggleCard'
import { AdminRunnersCard } from '../../shared/AdminRunnersCard'
import {
  adminAddSupportUsersToOrg,
  adminRemoveSupportUsersFromOrg,
  adminReprovisionOrg,
  adminRestartOrg,
  adminRestartOrgRunners,
  adminRestartOrgQueues,
  adminForceRestartOrgQueues,
  adminMigrateOrgQueues,
  adminRestartRunner,
  adminGracefulRunnerShutdown,
  adminForceRunnerShutdown,
  adminInvalidateRunnerToken,
  adminEnableOrgDebugMode,
  adminDeprovisionOrg,
  adminForgetOrgInstalls,
  adminForgetOrg,
  adminGracefulShutdownOrgProcesses,
  adminForceShutdownOrgProcesses,
} from '@/lib'

interface IAdminOrgSection {
  orgId: string
  org: TOrg
  adminEmail: string
  adminDashboardUrl: string | undefined
  runner: TRunner | undefined
  runnerLoading: boolean
}

export const AdminOrgSection = ({
  orgId,
  org,
  adminEmail,
  adminDashboardUrl,
  runner,
  runnerLoading,
}: IAdminOrgSection) => {
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
      <AdminActionGroup title="Org settings" icon="UsersIcon">
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

      <AdminActionGroup title="Infrastructure" icon="HardDrivesIcon" variant="warning">
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
        <AdminActionCard
          title="Restart org queues"
          description="Send a restart hint to all queue workflows for this org"
          action={() => adminRestartOrgQueues({ orgId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will send a restart hint to all queue workflows. Each queue will restart on its next poll cycle (~1-3 min)."
        />
        <AdminActionCard
          title="Force restart org queues"
          description="Immediately terminate and restart all queue workflows for this org via signal"
          action={async () => {
            const resp = await adminForceRestartOrgQueues({ orgId, adminEmail })
            if (resp?.queue_signal_id && resp?.queue_id) {
              window.open(`${adminDashboardUrl}/queues/${resp.queue_id}/signals/${resp.queue_signal_id}`, '_blank')
            }
          }}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will enqueue a signal that terminates and restarts ALL queue workflows for this organization. A new tab will open to monitor the signal progress."
        />
        <AdminActionCard
          title="Migrate org queues"
          description="Create all missing queues and enable the queues feature flag"
          action={() => adminMigrateOrgQueues({ orgId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will create all missing queues for apps, installs, runners, and components, then enable the queues feature flag. Continue?"
        />
      </AdminActionGroup>

      <AdminActionGroup title="Runner control" icon="PlayIcon">
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
        <AdminActionCard
          title="Graceful shutdown all processes"
          description="Request graceful shutdown of all active runner processes in this org"
          action={() => adminGracefulShutdownOrgProcesses({ orgId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will request graceful shutdown of ALL active runner processes for this organization. Each process will complete in-flight work before shutting down."
        />
        <AdminActionCard
          title="Force shutdown all processes"
          description="Force shutdown all active runner processes in this org"
          action={() => adminForceShutdownOrgProcesses({ orgId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will force shutdown ALL active runner processes for this organization. In-flight work may be lost."
        />
      </AdminActionGroup>

      <AdminActionGroup title="Teardown & cleanup" icon="TrashIcon" variant="danger">
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
        <AdminActionCard
          title="Forget org"
          description="Permanently forget this org and all its roles. This cannot be undone."
          action={() => adminForgetOrg({ orgId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          inputText="yesimsure"
          confirmationText="WARNING: This will permanently delete the organization record and all its roles. Any running infrastructure will be orphaned and must be cleaned up manually. This action CANNOT be undone."
        />
      </AdminActionGroup>

      <AdminActionGroup title="Security & debug" icon="ShieldIcon" variant="danger">
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
