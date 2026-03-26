import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'

import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { useActiveWorkflows } from '@/hooks/use-active-workflows'
import { useOrg } from '@/hooks/use-org'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import {
  getApp,
  getAppBranch,
  getAppConfigs,
  getInstall,
  getInstallStack,
  getRunnerLatestHeartbeat,
} from '@/lib'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getStatusTheme } from '@/utils/status-utils'
import { isLessThan15SecondsOld } from '@/utils/time-utils'

export const OrgStatusBar = () => {
  const { org } = useOrg()
  const { approvals } = useWorkflowApprovals()
  const { activeWorkflows } = useActiveWorkflows()
  const { appId, branchId, installId } = useParams()

  const { data: app } = useQuery({
    queryKey: ['app', org.id, appId],
    queryFn: () => getApp({ orgId: org.id, appId: appId! }),
    enabled: !!appId,
  })

  const { data: appConfigs } = useQuery({
    queryKey: ['app-configs', org.id, appId],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: appId!, limit: 1 }),
    enabled: !!appId,
    refetchInterval: 30_000,
  })
  const latestConfig = appConfigs?.[0]

  const { data: branch } = useQuery({
    queryKey: ['app-branch', org.id, appId, branchId],
    queryFn: () => getAppBranch({ orgId: org.id, appId: appId!, branchId: branchId! }),
    enabled: !!appId && !!branchId,
  })

  const { data: install } = useQuery({
    queryKey: ['install', org.id, installId],
    queryFn: () => getInstall({ orgId: org.id, installId: installId! }),
    enabled: !!installId,
  })

  const { data: stack } = useQuery({
    queryKey: ['install-stack', org.id, installId],
    queryFn: () => getInstallStack({ installId: installId!, orgId: org.id }),
    enabled: !!installId,
    refetchInterval: 30_000,
  })

  const runner = org.runner_group?.runners?.[0]
  const { data: heartbeats } = useQuery({
    queryKey: ['runner-heartbeat', org.id, runner?.id],
    queryFn: () =>
      getRunnerLatestHeartbeat({ runnerId: runner!.id!, orgId: org.id }),
    refetchInterval: 10_000,
    enabled: !!runner?.id,
  })
  const runnerHeartbeat =
    heartbeats?.install ?? heartbeats?.org ?? heartbeats?.build ?? undefined
  const runnerConnected = isLessThan15SecondsOld(runnerHeartbeat?.created_at)
  const runnerStatus = runnerConnected ? 'connected' : 'not-connected'

  const workflowItems: TContextTooltipItem[] = activeWorkflows.map((workflow) => ({
    id: workflow.id ?? '',
    title: workflow.name || toSentenceCase(snakeToWords(workflow.type)),
    subtitle: workflow.metadata?.owner_name || workflow.status?.status || undefined,
    href: workflow.owner_id
      ? `/${org.id}/installs/${workflow.owner_id}/workflows/${workflow.id}`
      : undefined,
    leftContent: (
      <Status
        status={workflow.status?.status ?? ''}
        isWithoutText
        variant="timeline"
        iconSize={16}
      />
    ),
  }))

  const ownerNames = new Map(
    activeWorkflows
      .filter((w) => w.owner_id && w.metadata?.owner_name)
      .map((w) => [w.owner_id!, w.metadata!.owner_name!])
  )

  const approvalItems: TContextTooltipItem[] = approvals.map((approval) => {
    const step = approval.workflow_step
    const href =
      step?.owner_id && step?.install_workflow_id
        ? `/${org.id}/installs/${step.owner_id}/workflows/${step.install_workflow_id}`
        : undefined
    const installName = step?.owner_id ? ownerNames.get(step.owner_id) : undefined
    return {
      id: approval.id ?? '',
      title: step?.name ? toSentenceCase(step.name) : 'Approval required',
      subtitle: installName || approval.type || undefined,
      href,
    }
  })

  return (
    <div className="hidden md:flex border-t w-full px-4 py-1.5 items-center flex-full sticky bottom-0 bg-code z-[1] gap-3">
      <Text family="mono" variant="subtext">
        {org.name}
      </Text>

      {runner && (
        <ContextTooltip
          position="top"
          title="Build runner"
          items={[
            {
              href: `/${org.id}/runner`,
              id: runner.id ?? 'runner',
              title: runnerConnected ? 'Connected' : 'Not connected',
              subtitle: runnerHeartbeat?.created_at ? (
                <Time
                  time={runnerHeartbeat.created_at}
                  variant="label"
                  theme="neutral"
                />
              ) : undefined,
              leftContent: (
                <Status
                  status={runnerStatus}
                  isWithoutText
                  variant="timeline"
                  iconSize={16}
                />
              ),
            },
          ]}
        >
          <Text theme={getStatusTheme(runnerStatus)}>
            <Icon variant="HammerIcon" size={14} className="cursor-default" />
          </Text>
        </ContextTooltip>
      )}

      <ContextTooltip
        position="top"
        title="Pending approvals"
        showCount
        width="w-64"
        items={approvalItems}
      >
        <Text
          theme={approvals.length ? 'warn' : 'neutral'}
          family="mono"
          variant="subtext"
          className="!flex gap-1.5 items-center cursor-default"
        >
          <Icon variant="BellIcon" size={14} />
          {approvals.length}
        </Text>
      </ContextTooltip>

      {activeWorkflows.length > 0 && (
        <ContextTooltip
          position="top"
          title="Active workflows"
          showCount
          width="w-72"
          items={workflowItems}
        >
          <Text
            theme="info"
            family="mono"
            variant="subtext"
            className="!flex gap-1.5 items-center cursor-default"
          >
            <Icon variant="TreeStructureIcon" size={14} />
            {activeWorkflows.length}
          </Text>
        </ContextTooltip>
      )}

      {app && (
        <>
          <span className="text-cool-grey-300 dark:text-white/20 text-xs">
            ›
          </span>
          <Text family="mono" variant="subtext">
            {app.name}
          </Text>

          {branch && (
            <>
              <span className="text-cool-grey-300 dark:text-white/20 text-xs">
                ›
              </span>
              <Icon variant="GitBranch" size={12} className="text-cool-grey-500 dark:text-cool-grey-400" />
              <Text family="mono" variant="subtext">
                {branch.name}
              </Text>
            </>
          )}

          {latestConfig && (
            <ContextTooltip
              position="top"
              title="Config sync"
              items={[
                {
                  id: latestConfig.id ?? 'config',
                  title: toSentenceCase(latestConfig.status ?? ''),
                  subtitle: latestConfig.created_at ? (
                    <Time
                      time={latestConfig.created_at}
                      variant="label"
                      theme="neutral"
                    />
                  ) : undefined,
                  leftContent: (
                    <Status
                      status={latestConfig.status ?? ''}
                      isWithoutText
                      variant="timeline"
                      iconSize={16}
                    />
                  ),
                },
              ]}
            >
              <Text theme={getStatusTheme(latestConfig.status ?? '')}>
                <Icon
                  variant="ArrowsCounterClockwiseIcon"
                  size={14}
                  className="cursor-default"
                />
              </Text>
            </ContextTooltip>
          )}
        </>
      )}

      {install && (
        <>
          <span className="text-cool-grey-300 dark:text-white/20 text-xs">
            ›
          </span>
          <Text family="mono" variant="subtext">
            {install.name}
          </Text>

          <InstallStatuses
            install={install}
            stack={stack}
            variant="icon"
            tooltipPosition="top"
          />
        </>
      )}
    </div>
  )
}
