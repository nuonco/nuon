'use client'

import { useEffect, useState } from 'react'
import type { TRunner } from '@/types'
import { AdminSection } from '../shared/AdminSection'
import { AdminActionGroup } from '../shared/AdminActionGroup'
import { AdminActionCard } from '../shared/AdminActionCard'
import { AdminMetadataPanel, AdminInfoCard } from '../shared/AdminMetadata'
import { AdminTemporalLink } from '../../old/AdminTemporalLink'

// Import admin actions
import { reprovisionInstall } from '@/actions/admin/reprovision-install'
import { reprovisionInstallRunner } from '@/actions/admin/reprovision-install-runner'
import { restartInstall } from '@/actions/admin/restart-install'
import { teardownInstallComponents } from '@/actions/admin/teardown-install-components'
import { updateInstallSandbox } from '@/actions/admin/update-install-sandbox'
import { shutdownInstallRunnerJob } from '@/actions/admin/shutdown-install-runner-job'
import { gracefulInstallRunnerShutdown } from '@/actions/admin/graceful-install-runner-shutdown'
import { forceInstallRunnerShutdown } from '@/actions/admin/force-install-runner-shutdown'
import { invalidateInstallRunnerToken } from '@/actions/admin/invalidate-install-runner-token'
import { getInstallRunner } from '@/actions/admin/get-install-runner'

interface AdminInstallSectionProps {
  orgId: string
  installId: string
}

export const AdminInstallSection = ({ orgId, installId }: AdminInstallSectionProps) => {
  const [runner, setRunner] = useState<TRunner>()
  const [runnerLoading, setRunnerLoading] = useState(true)

  useEffect(() => {
    if (installId) {
      getInstallRunner(installId).then((r) => {
        setRunner(r)
        setRunnerLoading(false)
      })
    }
  }, [installId])

  const metadata = (
    <AdminMetadataPanel>
      <AdminInfoCard 
        title="Install Runner ID" 
        value={runner?.id} 
        copyable 
        loading={runnerLoading}
      />
      <div className="space-y-1">
        <AdminTemporalLink namespace="installs" id={installId} />
      </div>
    </AdminMetadataPanel>
  )

  return (
    <AdminSection 
      title="Installation controls"
      subtitle={`Managing install: ${installId}`}
      metadata={metadata}
    >
      {/* Infrastructure */}
      <AdminActionGroup title="Install infrastructure" icon="HardDrives" variant="warning">
        <AdminActionCard
          title="Reprovision install"
          description="Reprovision current install sandbox and runner"
          action={() => reprovisionInstall(installId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will reprovision the entire install infrastructure including sandbox and runner."
        />
        <AdminActionCard
          title="Restart install"
          description="Restart current install event loop"
          action={() => restartInstall(installId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart the install event loop. Continue?"
        />
      </AdminActionGroup>

      {/* Component Management */}
      <AdminActionGroup title="Component management" icon="Cube" variant="danger">
        <AdminActionCard
          title="Teardown components"
          description="Teardown all components on this install"
          action={() => teardownInstallComponents(installId)}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will teardown ALL components on this install. This is destructive and cannot be undone."
        />
        <AdminActionCard
          title="Update sandbox"
          description="Update install sandbox to current app sandbox version"
          action={() => updateInstallSandbox(installId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will update the install sandbox to match the current app sandbox version."
        />
      </AdminActionGroup>

      {/* Runner Management */}
      <AdminActionGroup title="Install runner control" icon="Play">
        <AdminActionCard
          title="Reprovision runner"
          description="Reprovision current install runner"
          action={() => reprovisionInstallRunner(installId)}
          requiresConfirmation
          confirmationText="This will reprovision the install runner. Continue?"
        />
        <AdminActionCard
          title="Shutdown runner job"
          description="Shutdown the current install runner job"
          action={() => shutdownInstallRunnerJob(installId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will shutdown the current runner job. Continue?"
        />
        <AdminActionCard
          title="Graceful shutdown"
          description="Graceful shutdown of current install runner"
          action={() => gracefulInstallRunnerShutdown(installId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will gracefully shutdown the install runner. Continue?"
        />
        <AdminActionCard
          title="Force shutdown"
          description="Forceful shutdown of current install runner"
          action={() => forceInstallRunnerShutdown(installId)}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will forcefully shutdown the install runner and may cause data loss."
        />
      </AdminActionGroup>

      {/* Security */}
      <AdminActionGroup title="Security" icon="Shield" variant="danger">
        <AdminActionCard
          title="Invalidate runner token"
          description="Invalidate install runner service account token"
          action={() => invalidateInstallRunnerToken(installId)}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will invalidate the install runner service account token. The live runner will no longer be able to connect to the API."
        />
      </AdminActionGroup>
    </AdminSection>
  )
}