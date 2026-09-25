import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppInstalls } from '@/lib'
import type { TInstall, TInstallWorkflowStep } from '@/types'
import { DeployGroupStep } from './DeployGroupStep'
import type { IInstallDeployRow } from './InstallDeployRow'

interface IDeployGroupStepContainer {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
}

export const DeployGroupStepContainer = ({
  step,
  metadata,
}: IDeployGroupStepContainer) => {
  const { org } = useOrg()
  const { app } = useApp()

  const isPreviewApply = step.name?.toLowerCase() === 'apply preview install'
  const groupName = isPreviewApply
    ? (metadata.install_group_name as string) || 'preview install'
    : step.name?.replace(/^deploy install group:\s*/i, '') ||
      (metadata.install_group_name as string) ||
      'unknown'
  const installEntries = (metadata.installs as any[]) || []
  const totalInstalls =
    installEntries.length ||
    (metadata.total_installs as number) ||
    (metadata.install_count as number) ||
    0
  const deployedCount = installEntries.filter(
    (e: any) => e.status === 'success' || e.status === 'deployed'
  ).length

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

  const rows: IInstallDeployRow[] = installEntries.map((entry: any) => ({
    installId: entry.install_id,
    install: installsById[entry.install_id],
    deployStatus: entry.status,
    installHref:
      org?.id && entry.install_id
        ? `/${org.id}/installs/${entry.install_id}`
        : undefined,
    workflowHref:
      org?.id && entry.install_id && entry.workflow_id
        ? `/${org.id}/installs/${entry.install_id}/workflows/${entry.workflow_id}`
        : undefined,
  }))

  const emptyMessage =
    step.status?.status === 'in-progress'
      ? isPreviewApply
        ? 'Deploying to preview install'
        : 'Deploying to install group'
      : undefined

  return (
    <DeployGroupStep
      groupName={groupName}
      totalInstalls={totalInstalls}
      deployedCount={deployedCount}
      rows={rows}
      emptyMessage={emptyMessage}
      variant={isPreviewApply ? 'preview' : 'group'}
    />
  )
}
