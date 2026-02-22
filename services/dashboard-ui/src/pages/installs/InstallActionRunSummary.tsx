import { useParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { ActionStepGraph } from '@/components/actions/ActionStepsGraph'
import { InstallActionRunOutputs } from '@/components/actions/InstallActionRunOutputs'
import { Text } from '@/components/common/Text'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import { hydrateActionRunSteps } from '@/utils/action-utils'
import type { TInstallActionRun, TInstallAction } from '@/types'

export default function InstallActionRunSummary() {
  const { orgId, installId, actionId, runId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installActionRun } = usePolling<TInstallActionRun>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/${actionId}/runs/${runId}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const { data: installAction } = useQuery<TInstallAction>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/${actionId}`,
  })

  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun || null}
      shouldPoll
    >
      <div className="flex flex-col gap-6">
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: org?.name || '',
            },
            {
              path: `/${orgId}/installs`,
              text: 'Installs',
            },
            {
              path: `/${orgId}/installs/${installId}`,
              text: install?.name || '',
            },
            {
              path: `/${orgId}/installs/${installId}/actions`,
              text: 'Actions',
            },
            {
              path: `/${orgId}/installs/${installId}/actions/${actionId}`,
              text: installAction?.action_workflow?.name || 'Action',
            },
            {
              path: `/${orgId}/installs/${installId}/actions/${actionId}/${runId}`,
              text: `${installActionRun?.trigger_type || ''} run`,
            },
          ]}
        />
        <ActionStepGraph
          steps={hydrateActionRunSteps({
            steps: installActionRun?.steps,
            stepConfigs: installActionRun?.config?.steps,
          })}
        />

        <Text variant="body" weight="strong">
          Outputs
        </Text>
        <InstallActionRunOutputs />
      </div>
    </InstallActionRunProvider>
  )
}
