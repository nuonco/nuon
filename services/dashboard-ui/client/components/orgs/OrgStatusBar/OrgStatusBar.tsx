import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { Time } from '@/components/common/Time'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { VCSConnectionsStatusIndicator } from '@/components/vcs-connections/VCSConnectionsStatusIndicator'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getStatusTheme } from '@/utils/status-utils'
import type { TApp, TAppBranch, TAppConfig, TInstall, TInstallStack, TOrg, TRunnerHeartbeat, TWorkflow, TWorkflowStepApproval } from '@/types'

interface IOrgStatusBar {
  org: TOrg
  app?: TApp
  branch?: TAppBranch
  latestConfig?: TAppConfig
  install?: TInstall
  stack?: TInstallStack
  runnerConnected: boolean
  runnerStatus: string
  runnerHeartbeat?: TRunnerHeartbeat
  runnerId?: string
  approvals: TWorkflowStepApproval[]
  activeWorkflows: TWorkflow[]
  approvalItems: TContextTooltipItem[]
  workflowItems: TContextTooltipItem[]
}

export const OrgStatusBar = ({
  org,
  app,
  branch,
  latestConfig,
  install,
  stack,
  runnerConnected,
  runnerStatus,
  runnerHeartbeat,
  runnerId,
  approvals,
  activeWorkflows,
  approvalItems,
  workflowItems,
}: IOrgStatusBar) => {
  const runner = runnerId ? { id: runnerId } : undefined

  return (
    <div className="hidden md:flex border-t w-full px-4 py-1.5 items-center flex-none bg-code z-[1] gap-3">
      <Text family="mono" variant="subtext" className="!flex items-center gap-1.5">
        {org.sandbox_mode && (
          <Tooltip tipContent={<Text variant="subtext" as="span">Sandbox mode</Text>} tipContentClassName="!py-0.5" position="top">
            <Icon
              variant="TestTubeIcon"
              className="!w-[14px] !h-[14px] shrink-0"
              size="14"
            />
          </Tooltip>
        )}
        {org.name}
      </Text>

      <VCSConnectionsStatusIndicator />

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
              <Icon variant="GitBranchIcon" size={12} className="text-cool-grey-500 dark:text-cool-grey-400" />
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
