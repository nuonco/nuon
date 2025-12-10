'use client'

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

// Import admin actions
import { addSupportUsersToOrg } from '@/actions/admin/add-support-users'
import { removeSupportUsersFromOrg } from '@/actions/admin/remove-support-users'
import { reprovisionOrg } from '@/actions/admin/reprovision-org'
import { restartOrg } from '@/actions/admin/restart-org'
import { restartOrgRunners } from '@/actions/admin/restart-org-runners'
import { restartOrgRunner } from '@/actions/admin/restart-org-runner'
import { gracefulOrgRunnerShutdown } from '@/actions/admin/graceful-org-runner-shutdown'
import { forceOrgRunnerShutdown } from '@/actions/admin/force-org-runner-shutdown'
import { invalidateOrgRunnerToken } from '@/actions/admin/invalidate-org-runner-token'
import { enableOrgDebugMode } from '@/actions/admin/enable-org-debug-mode'
import { getOrgRunner } from '@/actions/admin/get-org-runner'

interface AdminOrgSectionProps {
  orgId: string
}

export const AdminOrgSection = ({ orgId }: AdminOrgSectionProps) => {
  const { org } = useOrg()
  const [runner, setRunner] = useState<TRunner>()
  const [runnerLoading, setRunnerLoading] = useState(true)

  useEffect(() => {
    if (orgId) {
      getOrgRunner(orgId).then((r) => {
        setRunner(r)
        setRunnerLoading(false)
      })
    }
  }, [orgId])

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
      {/* Org Settings */}
      <AdminActionGroup title="Org settings" icon="Users">
        <AdminActionCard
          title="Add support users"
          description="Add all Nuon support users to current org"
          action={() => addSupportUsersToOrg(orgId)}
        />
        <AdminActionCard
          title="Remove support users"
          description="Remove all Nuon support users from current org"
          action={() => removeSupportUsersFromOrg(orgId)}
        />
        <AdminFeatureToggleCard org={org} orgId={orgId} />
      </AdminActionGroup>

      {/* Infrastructure */}
      <AdminActionGroup
        title="Infrastructure"
        icon="HardDrives"
        variant="warning"
      >
        <AdminActionCard
          title="Reprovision org"
          description="Reprovision current org infrastructure"
          action={() => reprovisionOrg(orgId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will reprovision the entire organization infrastructure. This may cause downtime."
        />
        <AdminActionCard
          title="Restart org"
          description="Restart current org event loop"
          action={() => restartOrg(orgId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will restart the organization event loop. Continue?"
        />
      </AdminActionGroup>

      {/* Runner Management */}
      <AdminActionGroup title="Runner control" icon="Play">
        <AdminActionCard
          title="Restart all runners"
          description="Restart all of current org runners"
          action={() => restartOrgRunners(orgId)}
          requiresConfirmation
          confirmationText="This will restart all runners for this organization. Continue?"
        />
        <AdminActionCard
          title="Restart runner"
          description="Restart the current org runner"
          action={() => restartOrgRunner(orgId)}
          requiresConfirmation
          confirmationText="This will restart the org runner. Continue?"
        />
        <AdminActionCard
          title="Graceful shutdown"
          description="Graceful shutdown of current org runner"
          action={() => gracefulOrgRunnerShutdown(orgId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will gracefully shutdown the org runner. Continue?"
        />
        <AdminActionCard
          title="Force shutdown"
          description="Forceful shutdown of current org runner"
          action={() => forceOrgRunnerShutdown(orgId)}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will forcefully shutdown the org runner and may cause data loss."
        />
      </AdminActionGroup>

      {/* Security & Debug */}
      <AdminActionGroup title="Security & debug" icon="Shield" variant="danger">
        <AdminActionCard
          title="Invalidate runner token"
          description="Invalidate runner service account token"
          action={() => invalidateOrgRunnerToken(orgId)}
          variant="danger"
          requiresConfirmation
          requiresInput
          confirmationText="This will invalidate the runner service account token. Live runners will no longer be able to connect to the API."
        />
        <AdminActionCard
          title="Enable debug mode"
          description="Enable debug mode for this org (logs all requests)"
          action={() => enableOrgDebugMode(orgId)}
          variant="warning"
          requiresConfirmation
          confirmationText="This will enable debug mode which logs all requests for this org."
        />
      </AdminActionGroup>
    </AdminSection>
  )
}
