import type { TRunner } from '@/types'
import { AdminSection } from '../../shared/AdminSection'
import { AdminActionGroup } from '../../shared/AdminActionGroup'
import { AdminActionCard } from '../../shared/AdminActionCard'
import { AdminMetadataPanel, AdminInfoCard } from '../../shared/AdminMetadata'
import { TemporalLink } from '@/components/admin/TemporalLink'
import {
  adminReprovisionInstall,
  adminReprovisionInstallRunner,
  adminRestartInstall,
  adminRestartInstallQueues,
  adminTeardownInstallComponents,
  adminUpdateInstallSandbox,
  adminShutdownRunnerJob,
  adminGracefulRunnerShutdown,
  adminForceRunnerShutdown,
  adminInvalidateRunnerToken,
} from '@/lib'

interface IAdminInstallSection {
  orgId: string
  installId: string
  adminEmail: string
  runner: TRunner | undefined
  runnerLoading: boolean
}

export const AdminInstallSection = ({
  orgId,
  installId,
  adminEmail,
  runner,
  runnerLoading,
}: IAdminInstallSection) => {
  const runnerId = runner?.id ?? ''

  const metadata = (
    <AdminMetadataPanel>
      <AdminInfoCard
        title="Install Runner ID"
        value={runner?.id}
        copyable
        loading={runnerLoading}
      />
      <div className="space-y-1">
        <TemporalLink namespace="installs" eventLoopId={installId} />
      </div>
    </AdminMetadataPanel>
  )

  return (
    <AdminSection
      title="Installation controls"
      subtitle={`Managing install: ${installId}`}
      metadata={metadata}
    >
      <AdminActionGroup title="Install infrastructure" icon="HardDrives" variant="warning">
        <AdminActionCard
          title="Reprovision install"
          description="Reprovision current install sandbox and runner"
          action={() => adminReprovisionInstall({ installId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will reprovision the entire install infrastructure including sandbox and runner."
        />
        <AdminActionCard
          title="Restart install"
          description="Restart current install event loop"
          action={() => adminRestartInstall({ installId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart the install event loop. Continue?"
        />
        <AdminActionCard
          title="Restart install queues"
          description="Restart all Temporal queue workflows for this install"
          action={() => adminRestartInstallQueues({ installId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart all queue Temporal workflows for this install. Continue?"
        />
      </AdminActionGroup>

      <AdminActionGroup title="Component management" icon="Cube" variant="danger">
        <AdminActionCard
          title="Teardown components"
          description="Teardown all components on this install"
          action={() => adminTeardownInstallComponents({ installId, orgId })}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will teardown ALL components on this install. This is destructive and cannot be undone."
        />
        <AdminActionCard
          title="Update sandbox"
          description="Update install sandbox to current app sandbox version"
          action={() => adminUpdateInstallSandbox({ installId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will update the install sandbox to match the current app sandbox version."
        />
      </AdminActionGroup>

      <AdminActionGroup title="Install runner control" icon="Play">
        <AdminActionCard
          title="Reprovision runner"
          description="Reprovision current install runner"
          action={() => adminReprovisionInstallRunner({ runnerId, adminEmail })}
          requiresConfirmation
          confirmationText="This will reprovision the install runner. Continue?"
        />
        <AdminActionCard
          title="Shutdown runner job"
          description="Shutdown the current install runner job"
          action={() => adminShutdownRunnerJob({ installId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will shutdown the current runner job. Continue?"
        />
        <AdminActionCard
          title="Graceful shutdown"
          description="Graceful shutdown of current install runner"
          action={() => adminGracefulRunnerShutdown({ runnerId, adminEmail })}
          variant="warning"
          requiresConfirmation
          confirmationText="This will gracefully shutdown the install runner. Continue?"
        />
        <AdminActionCard
          title="Force shutdown"
          description="Forceful shutdown of current install runner"
          action={() => adminForceRunnerShutdown({ runnerId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will forcefully shutdown the install runner and may cause data loss."
        />
      </AdminActionGroup>

      <AdminActionGroup title="Security" icon="Shield" variant="danger">
        <AdminActionCard
          title="Invalidate runner token"
          description="Invalidate install runner service account token"
          action={() => adminInvalidateRunnerToken({ runnerId, adminEmail })}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will invalidate the install runner service account token. The live runner will no longer be able to connect to the API."
        />
      </AdminActionGroup>
    </AdminSection>
  )
}
