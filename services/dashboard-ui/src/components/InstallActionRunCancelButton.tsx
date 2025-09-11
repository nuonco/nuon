'use client'

import { InstallWorkflowCancelModal } from '@/components/InstallWorkflows/InstallWorkflowCancelModal'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import type { TWorkflow } from '@/types'

export const InstallActionCancelButton = ({
  workflow,
}: {
  workflow: TWorkflow
}) => {
  const { installActionRun } = useInstallActionRun()
  return installActionRun?.runner_job?.id &&
    (installActionRun?.status === 'queued' ||
      installActionRun?.status === 'in-progress') &&
    workflow &&
    !workflow?.finished ? (
    <InstallWorkflowCancelModal installWorkflow={workflow} />
  ) : null
}
