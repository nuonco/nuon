import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Badge } from '@/components/common/Badge'
import { CompositeError } from '@/components/common/CompositeError'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { Panel } from '@/components/surfaces/Panel'
import type { TWorkflowStep } from '@/types'
import { WORKFLOW_BADGE_MAP } from '@/utils/workflow-utils'
import type { TLatestRollout, TRolloutInstall } from './fixtures'

const awaitingApprovalBadge = WORKFLOW_BADGE_MAP['approval-awaiting']

const versionLabel = (install: TRolloutInstall) =>
  install.status === 'in-progress' || install.status === 'error'
    ? `v${install.configVersion} → v${install.targetConfigVersion}`
    : `v${install.configVersion}`

const approvalStep = (install: TRolloutInstall): TWorkflowStep =>
  ({
    id: `${install.id}-plan`,
    install_workflow_id: `wf-${install.id}`,
    execution_type: 'approval',
    approval: {
      id: `${install.id}-approval`,
      type: 'helm_approval',
    },
    status: { status: 'approval-awaiting' },
  }) as TWorkflowStep

const AwaitingApprovalBadge = () => (
  <Badge size="sm" {...awaitingApprovalBadge} />
)

const InstallCardFace = ({
  install,
  commit,
}: {
  install: TRolloutInstall
  commit: TLatestRollout['commit']
}) => (
  <>
    <span className="flex min-w-0 items-baseline gap-2">
      <Text variant="body" weight="strong" className="truncate">
        {install.name}
      </Text>
      <Text
        variant="subtext"
        theme="neutral"
        family="mono"
        className="shrink-0"
      >
        {versionLabel(install)}
      </Text>
      <Text
        variant="subtext"
        theme="neutral"
        family="mono"
        flex
        className="shrink-0"
      >
        <Icon variant="GitCommitIcon" />
        {commit.previousSha.slice(0, 7)} → {commit.sha.slice(0, 7)}
      </Text>
    </span>
    <span className="flex shrink-0 items-center gap-2">
      {install.awaitingApproval ? <AwaitingApprovalBadge /> : null}
      <Status status={install.status} />
    </span>
  </>
)

export const InstallRolloutPanel = ({
  install,
  commit,
}: {
  install: TRolloutInstall
  commit: TLatestRollout['commit']
}) => {
  if (install.status === 'pending') {
    return (
      <Tooltip
        className="!w-full"
        position="top"
        tipContent="Rollout hasn't started yet."
      >
        <div className="flex w-full cursor-default items-center justify-between gap-3 rounded-md border bg-white px-4 py-3 text-left shadow-sm dark:bg-dark-grey-900">
          <InstallCardFace install={install} commit={commit} />
        </div>
      </Tooltip>
    )
  }

  return (
    <Panel
      panelKey={`install-rollout-${install.id}`}
      size="3/4"
      heading={
        <span className="flex min-w-0 items-center gap-3">
          <Text variant="base" weight="strong" className="truncate">
            {install.name}
          </Text>
          {install.awaitingApproval ? <AwaitingApprovalBadge /> : null}
          <Status status={install.status} />
        </span>
      }
      triggerButton={{
        variant: 'ghost',
        className:
          '!h-auto !w-full !whitespace-normal !justify-between !gap-3 !px-4 !py-3 text-left !rounded-md !shadow-sm !bg-white dark:!bg-dark-grey-900 !text-inherit !border !border-[color:var(--border-color)] !transition-[background-color] !duration-fast !ease-cubic hover:!bg-cool-grey-50 dark:hover:!bg-white/5',
        children: <InstallCardFace install={install} commit={commit} />,
      }}
    >
      {install.compositeError ? (
        <CompositeError error={install.compositeError} />
      ) : null}
      {install.awaitingApproval ? (
        <ApprovalBanner step={approvalStep(install)} />
      ) : null}
      <Tabs
        naturalHeight
        initActiveTab="workflow"
        tabs={{
          workflow: (
            <Text variant="subtext" theme="neutral">
              Provision workflow placeholder
            </Text>
          ),
          changes: (
            <Text variant="subtext" theme="neutral">
              Changes placeholder
            </Text>
          ),
          template: (
            <Text variant="subtext" theme="neutral">
              Template placeholder
            </Text>
          ),
        }}
      />
    </Panel>
  )
}
