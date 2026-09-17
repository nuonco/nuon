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
import { humanize } from '@/utils/string-utils'
import { getStatusTheme } from '@/utils/status-utils'
import { cn } from '@/utils/classnames'
import type { TApp, TAppBranch, TAppConfig, TInstall, TInstallStack, TOrg, TWorkflow, TWorkflowStepApproval } from '@/types'

interface IOrgStatusBar {
  org: TOrg
  app?: TApp
  branch?: TAppBranch
  latestConfig?: TAppConfig
  install?: TInstall
  stack?: TInstallStack
  approvals: TWorkflowStepApproval[]
  activeWorkflows: TWorkflow[]
  approvalItems: TContextTooltipItem[]
  workflowItems: TContextTooltipItem[]
  byocName?: string
  byocColor?: string
  byocTextColor?: string
}

const NAME_WIDTHS = {
  org: 'shrink-0 max-w-[16ch] @6xl:max-w-[24ch] @7xl:max-w-[32ch]',
  app: 'shrink-0 max-w-[20ch] @6xl:max-w-[28ch] @7xl:max-w-[36ch]',
  branch: 'shrink-0 max-w-[16ch] @6xl:max-w-[24ch] @7xl:max-w-[32ch]',
  install: 'min-w-[14ch] @max-[29rem]:min-w-[10ch]',
}

const HIDE_TIERS = {
  org: '@max-[54rem]:!hidden',
  app: '@max-[47rem]:hidden',
  branch: '@max-[38rem]:hidden',
}

const Separator = () => (
  <span className="shrink-0 text-cool-grey-300 dark:text-white/20 text-xs">
    ›
  </span>
)

export const OrgStatusBar = ({
  org,
  app,
  branch,
  latestConfig,
  install,
  stack,
  approvals,
  activeWorkflows,
  approvalItems,
  workflowItems,
  byocName,
  byocColor,
  byocTextColor,
}: IOrgStatusBar) => {
  const hasResourceContext = !!app || !!install

  return (
    <div className="@container hidden md:flex border-t w-full px-4 py-1.5 items-center flex-none flex-nowrap overflow-hidden bg-code z-[1] gap-3">
      <Text
        family="mono"
        variant="subtext"
        nowrap
        className={cn('!flex items-center gap-1.5', NAME_WIDTHS.org, {
          [HIDE_TIERS.org]: hasResourceContext,
        })}
      >
        {org.sandbox_mode && (
          <Tooltip tipContent={<Text variant="subtext" as="span">Sandbox mode</Text>} tipContentClassName="!py-0.5" position="top">
            <Icon
              variant="TestTubeIcon"
              className="!w-[14px] !h-[14px] shrink-0"
              size="14"
            />
          </Tooltip>
        )}
        <span className="min-w-0 truncate">{org.name}</span>
      </Text>

      <VCSConnectionsStatusIndicator />

      <ContextTooltip
        className="shrink-0"
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
          nowrap
          className="!flex gap-1.5 items-center cursor-default"
        >
          <Icon variant="BellIcon" size={14} />
          {approvals.length}
        </Text>
      </ContextTooltip>

      {activeWorkflows.length > 0 && (
        <ContextTooltip
          className="shrink-0"
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
            nowrap
            className="!flex gap-1.5 items-center cursor-default"
          >
            <Icon variant="TreeStructureIcon" size={14} />
            {activeWorkflows.length}
          </Text>
        </ContextTooltip>
      )}

      {app && (
        <>
          <div
            className={cn('flex items-center gap-3 shrink-0', {
              [HIDE_TIERS.app]: !!install,
            })}
          >
            <Separator />
            <Text
              family="mono"
              variant="subtext"
              nowrap
              className={cn('truncate', NAME_WIDTHS.app)}
            >
              {app.name}
            </Text>
          </div>

          {branch && (
            <div
              className={cn('flex items-center gap-3 shrink-0', {
                [HIDE_TIERS.branch]: !!install,
              })}
            >
              <Separator />
              <Icon variant="GitBranchIcon" size={12} className="shrink-0 text-cool-grey-500 dark:text-cool-grey-400" />
              <Text
                family="mono"
                variant="subtext"
                nowrap
                className={cn('truncate', NAME_WIDTHS.branch)}
              >
                {branch.name}
              </Text>
            </div>
          )}

          {latestConfig && (
            <ContextTooltip
              className="shrink-0"
              position="top"
              title="Config sync"
              items={[
                {
                  id: latestConfig.id ?? 'config',
                  title: humanize(latestConfig.status ?? ''),
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
              <Text theme={getStatusTheme(latestConfig.status ?? '')} className="!flex">
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
        <div className="flex items-center gap-3 min-w-0">
          <Separator />
          <Text
            family="mono"
            variant="subtext"
            nowrap
            className={cn('truncate', NAME_WIDTHS.install)}
          >
            {install.name}
          </Text>

          <InstallStatuses
            className="shrink-0"
            install={install}
            stack={stack}
            variant="icon"
            tooltipPosition="top"
          />
        </div>
      )}

      {byocName && (
        <span
          className="ml-auto shrink-0 max-w-32 @3xl:max-w-56 truncate rounded px-2 py-px font-mono text-xs font-strong tracking-widest uppercase"
          style={{ backgroundColor: byocColor, color: byocTextColor }}
        >
          {byocName}
        </span>
      )}
    </div>
  )
}
