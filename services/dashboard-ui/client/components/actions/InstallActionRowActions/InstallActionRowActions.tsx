import { useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Menu } from '@/components/common/Menu'
import { SplitButton } from '@/components/common/SplitButton'
import { InstallActionRunTimeline } from '@/components/actions/InstallActionRunTimeline/InstallActionRunTimeline'
import { ActionConfigPanel } from '@/components/installs/InstallResourceConfig'
import { Panel } from '@/components/surfaces/Panel'
import { WorkflowPanelLink } from '@/components/workflows/InstallWorkflowPanel'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallAction } from '@/lib'
import type { TAction, TActionConfig } from '@/types'
import { InstallActionManualRunModal } from '@/components/actions/InstallActionManualRun'

const HISTORY_LIMIT = 10

const ActionRunHistory = ({
  actionId,
  name,
}: {
  actionId: string
  name: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const [offset, setOffset] = useState(0)

  const { data, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-action-history',
      org?.id,
      install?.id,
      actionId,
      offset,
    ],
    queryFn: () =>
      getInstallAction({
        orgId: org!.id,
        installId: install!.id,
        actionId,
        limit: HISTORY_LIMIT,
        offset,
      }),
    enabled: !!org?.id && !!install?.id && !!actionId,
  })

  const runs = data?.runs ?? []
  const hasNext = runs.length >= HISTORY_LIMIT

  if (isLoading && !data) return <Loading />

  return (
    <div className="flex flex-col gap-4">
      <InstallActionRunTimeline
        actionId={actionId}
        actionName={name}
        basePath={installLink({
          installId: install?.id,
          appId: install?.app_id,
        })}
        pagination={{ hasNext: false, offset: 0, limit: HISTORY_LIMIT }}
        renderTitle={(run) => {
          const workflowId = run.workflow_id || run.install_workflow_id
          if (!workflowId) return <span>{name} run</span>
          return (
            <WorkflowPanelLink variant="inline" workflowId={workflowId}>
              {name} run
            </WorkflowPanelLink>
          )
        }}
        runs={runs}
      />
      {offset > 0 || hasNext ? (
        <div className="flex items-center justify-center gap-3">
          <Button
            disabled={offset === 0}
            onClick={() =>
              setOffset((current) => Math.max(current - HISTORY_LIMIT, 0))
            }
            size="sm"
            variant="secondary"
          >
            <Icon variant="CaretLeftIcon" />
            Previous
          </Button>
          <Button
            disabled={!hasNext}
            onClick={() => setOffset((current) => current + HISTORY_LIMIT)}
            size="sm"
            variant="secondary"
          >
            Next
            <Icon variant="CaretRightIcon" />
          </Button>
        </div>
      ) : null}
    </div>
  )
}

export const InstallActionRowActions = ({
  action,
  actionId,
  config,
  name,
  removed,
}: {
  action?: TAction
  actionId: string
  config?: TActionConfig
  name: string
  removed?: boolean
}) => {
  const { addModal, addPanel } = useSurfaces()
  const canRun = !!config?.triggers?.some(
    (trigger) => trigger.type === 'manual'
  )
  const canStart = canRun && !removed && !!action && !!config?.id
  const disabledReason = removed
    ? "This action is no longer in the install's app config version."
    : !canRun
      ? 'This action does not have a manual trigger.'
      : undefined

  return (
    <SplitButton
      size="sm"
      variant="secondary"
      buttonProps={{
        children: (
          <>
            Run action
            <Icon variant="PlayIcon" />
          </>
        ),
        disabled: !canStart,
        onClick: () => {
          if (!action || !config?.id) return
          addModal(
            <InstallActionManualRunModal
              action={action}
              actionConfigId={config.id}
            />
          )
        },
        tooltipProps: disabledReason
          ? { position: 'left', tipContent: disabledReason }
          : undefined,
      }}
      dropdownProps={{
        alignment: 'right',
        id: `action-row-actions-${actionId}`,
        children: (
          <Menu>
            <Button
              onClick={() =>
                addPanel(
                  <ActionConfigPanel config={config} name={name} />,
                  `action-config-${config?.action_workflow_id ?? name}`
                )
              }
            >
              View config
              <Icon variant="SlidersHorizontalIcon" />
            </Button>
            <Button
              onClick={() =>
                addPanel(
                  <Panel
                    heading={`${name} history`}
                    panelKey={`action-history-${actionId}`}
                    size="half"
                  >
                    <ActionRunHistory actionId={actionId} name={name} />
                  </Panel>,
                  `action-history-${actionId}`
                )
              }
            >
              View history
              <Icon variant="ClockCounterClockwiseIcon" />
            </Button>
          </Menu>
        ),
      }}
    />
  )
}
