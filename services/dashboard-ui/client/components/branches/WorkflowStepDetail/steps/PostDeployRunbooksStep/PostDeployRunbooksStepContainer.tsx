import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppInstalls } from '@/lib'
import type { TInstall, TInstallWorkflowStep } from '@/types'
import { PostDeployRunbooksStep } from './PostDeployRunbooksStep'
import type { IInstallRunbooksRow } from './InstallRunbooksRow'

interface IPostDeployRunbooksStepContainer {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
}

export const PostDeployRunbooksStepContainer = ({
  step,
  metadata,
}: IPostDeployRunbooksStepContainer) => {
  const { org } = useOrg()
  const { app } = useApp()

  const groupName =
    step.name?.replace(/^run post-deploy runbooks:\s*/i, '') || 'unknown'
  const installEntries = (metadata.installs as any[]) || []

  const { data: appInstalls } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs', org?.id, app?.id],
    queryFn: () =>
      getAppInstalls({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id && installEntries.length > 0,
  })

  const installsById = useMemo(() => {
    const map: Record<string, TInstall> = {}
    for (const inst of appInstalls?.data || []) {
      if (inst.id) map[inst.id] = inst
    }
    return map
  }, [appInstalls])

  const rows: IInstallRunbooksRow[] = installEntries
    .filter((entry: any) => ((entry.runbooks as any[]) || []).length > 0)
    .map((entry: any) => ({
      installId: entry.install_id,
      install: installsById[entry.install_id],
      installHref:
        org?.id && entry.install_id
          ? `/${org.id}/installs/${entry.install_id}`
          : undefined,
      runbooks: ((entry.runbooks as any[]) || []).map((runbook: any) => ({
        runbookId: runbook.runbook_id,
        runbookName: runbook.runbook_name || runbook.runbook_id,
        status: runbook.status,
        workflowHref:
          org?.id && entry.install_id && runbook.workflow_id
            ? `/${org.id}/installs/${entry.install_id}/workflows/${runbook.workflow_id}`
            : undefined,
      })),
    }))

  const runbookNames = useMemo(() => {
    const seen: string[] = []
    for (const row of rows) {
      for (const runbook of row.runbooks) {
        if (!seen.includes(runbook.runbookName)) seen.push(runbook.runbookName)
      }
    }
    return seen
  }, [rows])

  const emptyMessage =
    step.status?.status === 'in-progress'
      ? 'Starting post-deploy runbooks'
      : undefined

  return (
    <PostDeployRunbooksStep
      groupName={groupName}
      runbookNames={runbookNames}
      rows={rows}
      emptyMessage={emptyMessage}
      statusDescription={step.status?.status_human_description}
    />
  )
}
