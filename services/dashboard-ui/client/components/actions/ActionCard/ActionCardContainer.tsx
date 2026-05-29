import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallActionsLatestRuns } from '@/lib'
import type { TActionConfigTriggerType } from '@/types'
import { InstallActionManualRunModalContainer } from '../InstallActionManualRun/InstallActionManualRunContainer'
import { ActionCard } from './ActionCard'

interface IActionCardContainer {
  id?: string
  name?: string
}

export const ActionCardContainer = ({ id, name }: IActionCardContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const { data: result, isLoading, error } = useQuery({
    queryKey: ['install-actions-card', org?.id, install?.id, name, id],
    queryFn: () =>
      getInstallActionsLatestRuns({
        orgId: org.id,
        installId: install.id,
        limit: 50,
        offset: 0,
        q: name || undefined,
      }),
    enabled: !!org?.id && !!install?.id && !!(name || id),
  })

  if (!id && !name) {
    return <ActionCard error="Missing id or name attribute" />
  }

  const actions = result?.data ?? []
  const action = name
    ? actions.find((a) => a.action_workflow?.name === name)
    : actions.find((a) => a.action_workflow_id === id)

  if (!isLoading && !error && actions.length > 0 && !action) {
    return <ActionCard error={`Action "${name || id}" not found`} />
  }

  const recentRun = action?.runs?.at(0)
  const actionWorkflow = action?.action_workflow
  const actionConfigId = actionWorkflow?.configs?.[0]?.id
  const triggers = actionWorkflow?.configs?.[0]?.triggers ?? []
  const canRun = triggers.some((t) => t.type === 'manual')
  const href = action?.action_workflow_id
    ? `/${org.id}/installs/${install.id}/actions/${action.action_workflow_id}`
    : undefined

  return (
    <ActionCard
      name={actionWorkflow?.name || name}
      triggerType={recentRun?.triggered_by_type as TActionConfigTriggerType}
      status={recentRun?.status_v2?.status}
      href={href}
      isLoading={isLoading}
      error={error ? 'Failed to load action' : undefined}
      hasRun={!!recentRun}
      canRun={canRun}
      onRun={canRun && actionWorkflow && actionConfigId ? () => {
        addModal(
          <InstallActionManualRunModalContainer
            action={actionWorkflow}
            actionConfigId={actionConfigId}
          />
        )
      } : undefined}
    />
  )
}
