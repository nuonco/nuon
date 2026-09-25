import { useQueries } from '@tanstack/react-query'
import { InstallActionManualRunModal } from '@/components/actions/InstallActionManualRun'
import { RunRunbookModal } from '@/components/runbooks/RunRunbook'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallAction, getInstallRunbook } from '@/lib'
import type { TInstallAction } from '@/types'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'
import { isActionInAppConfig } from '@/utils/app-config-membership'
import { ActivityPins, type TActivityPinCard } from './ActivityPins'
import { ActivityPinsPanel } from './ActivityPinsPanel'
import { useActivityPins } from './use-activity-pins'

export const ActivityPinsContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig } = useInstallAppConfig()
  const { addModal } = useSurfaces()
  const { pins, toggle } = useActivityPins()

  const queries = useQueries({
    queries: pins.map((pin) =>
      pin.kind === 'runbook'
        ? {
            queryKey: ['activity-pin', 'runbook', org?.id, install?.id, pin.id],
            queryFn: () =>
              getInstallRunbook({
                orgId: org!.id,
                installId: install!.id,
                runbookId: pin.id,
              }),
            enabled: !!org?.id && !!install?.id,
          }
        : {
            queryKey: ['activity-pin', 'action', org?.id, install?.id, pin.id],
            queryFn: () =>
              getInstallAction({
                orgId: org!.id,
                installId: install!.id,
                actionId: pin.id,
                limit: 1,
                offset: 0,
              }),
            enabled: !!org?.id && !!install?.id,
          }
    ),
  })

  const items: TActivityPinCard[] = pins.map((pin, index) => {
    const query = queries[index]
    const loading = !!query?.isLoading
    const missing = !!query?.isError

    if (pin.kind === 'runbook') {
      const installRunbook = query?.data as TInstallRunbook | undefined
      return {
        id: `runbook:${pin.id}`,
        kind: pin.kind,
        name: installRunbook?.runbook?.name ?? '',
        loading,
        missing,
        onRemove: missing ? () => toggle(pin) : undefined,
        onRun:
          installRunbook && !missing
            ? () => addModal(<RunRunbookModal installRunbook={installRunbook} />)
            : undefined,
      }
    }

    const installAction = query?.data as TInstallAction | undefined
    const workflow = installAction?.action_workflow
    const actionConfigId = workflow?.configs?.[0]?.id
    const removed = !isActionInAppConfig(appConfig, pin.id)
    const canRun = !!workflow && !!actionConfigId && !removed

    return {
      id: `action:${pin.id}`,
      kind: pin.kind,
      name: workflow?.name ?? '',
      loading,
      missing,
      onRemove:
        missing || (!loading && !canRun) ? () => toggle(pin) : undefined,
      onRun:
        canRun && workflow && actionConfigId
          ? () =>
              addModal(
                <InstallActionManualRunModal
                  action={workflow}
                  actionConfigId={actionConfigId}
                />
              )
          : undefined,
    }
  })

  return (
    <ActivityPins
      items={items}
      panel={
        <ActivityPinsPanel
          panelKey="activity-pins"
          triggerButton={{
            variant: 'secondary',
            children: pins.length ? 'Edit pins' : 'Pin shortcuts',
          }}
        />
      }
    />
  )
}
