import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Menu } from '@/components/common/Menu'
import { SplitButton } from '@/components/common/SplitButton'
import { RunbookConfigPanel } from '@/components/installs/InstallResourceConfig'
import { RunbookReadmePanel } from '@/components/runbooks/RunbookReadmePanel'
import { RunbookRunTimeline } from '@/components/runbooks/RunbookRunTimeline/RunbookRunTimeline'
import { Panel } from '@/components/surfaces/Panel'
import { WorkflowPanelLink } from '@/components/workflows/InstallWorkflowPanel'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallRunbook } from '@/lib'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'
import { RunRunbookModal } from '@/components/runbooks/RunRunbook'

const RunbookRunHistory = ({
  runbookId,
  runbookName,
}: {
  runbookId: string
  runbookName: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()

  const { data, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-runbook', org?.id, install?.id, runbookId],
    queryFn: () =>
      getInstallRunbook({
        orgId: org!.id,
        installId: install!.id,
        runbookId,
      }),
    enabled: !!org?.id && !!install?.id && !!runbookId,
  })

  if (isLoading && !data) return <Loading />

  return (
    <RunbookRunTimeline
      basePath={installLink({
        installId: install?.id,
        appId: install?.app_id,
      })}
      renderTitle={(run) => {
        const workflowId = run.install_workflow_id ?? run.install_workflow?.id
        if (!workflowId) return <span>{runbookName} run</span>
        return (
          <WorkflowPanelLink variant="inline" workflowId={workflowId}>
            {runbookName} run
          </WorkflowPanelLink>
        )
      }}
      runbookName={runbookName}
      runs={data?.runs ?? []}
    />
  )
}

export const InstallRunbookRowActions = ({
  installRunbook,
}: {
  installRunbook: TInstallRunbook
}) => {
  const { addModal, addPanel } = useSurfaces()
  const runbook = installRunbook.runbook
  const runbookId = installRunbook.runbook_id ?? installRunbook.id
  const name = runbook?.name ?? 'Runbook'
  const config = runbook?.configs?.[0]
  const readme = config?.readme

  return (
    <SplitButton
      size="sm"
      variant="secondary"
      buttonProps={{
        children: (
          <>
            Run runbook
            <Icon variant="PlayIcon" />
          </>
        ),
        onClick: () =>
          addModal(<RunRunbookModal installRunbook={installRunbook} />),
      }}
      dropdownProps={{
        alignment: 'right',
        id: `runbook-row-actions-${runbookId}`,
        children: (
          <Menu>
            <Button
              onClick={() =>
                addPanel(
                  <RunbookConfigPanel config={config} name={name} />,
                  `runbook-config-${config?.runbook_id ?? name}`
                )
              }
            >
              View config
              <Icon variant="SlidersHorizontalIcon" />
            </Button>
            {readme ? (
              <Button
                onClick={() =>
                  addPanel(
                    <RunbookReadmePanel
                      panelKey={`runbook-readme-${runbookId}`}
                      readme={readme}
                      runbookName={name}
                      size="half"
                    />,
                    `runbook-readme-${runbookId}`
                  )
                }
              >
                View readme
                <Icon variant="BookOpenTextIcon" />
              </Button>
            ) : null}
            <Button
              onClick={() =>
                addPanel(
                  <Panel
                    heading={`${name} history`}
                    panelKey={`runbook-history-${runbookId}`}
                    size="half"
                  >
                    <RunbookRunHistory
                      runbookId={runbookId}
                      runbookName={name}
                    />
                  </Panel>,
                  `runbook-history-${runbookId}`
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
