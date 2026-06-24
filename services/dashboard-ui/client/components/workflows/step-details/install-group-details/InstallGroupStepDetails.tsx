import { Badge } from '@/components/common/Badge'
import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import type { TWorkflowStep } from '@/types'
import { StepButtons } from '../StepButtons'

interface IInstallDiffEntry {
  component_id?: string
  component_name?: string
  component_type?: string
  old_checksum?: string
  new_checksum?: string
}

interface IInstallPlanEntry {
  install_id?: string
  install_name?: string
  status?: string
  workflow_id?: string
  install_config_update_id?: string
  diff?: {
    added?: IInstallDiffEntry[]
    changed?: IInstallDiffEntry[]
    removed?: IInstallDiffEntry[]
    unchanged?: IInstallDiffEntry[]
    sandbox_changed?: boolean
    stack_changed?: boolean
  }
}

interface IInstallGroupPlan {
  install_group?: string
  installs?: IInstallPlanEntry[]
}

interface IInstallMetadataEntry {
  install_id?: string
  status?: string
  workflow_id?: string
  added?: number
  changed?: number
  removed?: number
  unchanged?: number
  sandbox_changed?: boolean
  stack_changed?: boolean
}

const ComponentList = ({
  label,
  entries,
  theme,
}: {
  label: string
  entries?: IInstallDiffEntry[]
  theme: 'success' | 'error' | 'info'
}) => {
  if (!entries?.length) return null
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <Text variant="subtext" theme="neutral">
        {label}:
      </Text>
      {entries.map((e) => (
        <Badge key={e?.component_id} theme={theme} size="sm" variant="code">
          {e?.component_name || e?.component_id}
        </Badge>
      ))}
    </div>
  )
}

const InstallDiffDetail = ({ diff }: { diff: IInstallPlanEntry['diff'] }) => {
  if (!diff) return null

  const hasChanges =
    (diff.added?.length ?? 0) > 0 ||
    (diff.changed?.length ?? 0) > 0 ||
    (diff.removed?.length ?? 0) > 0

  if (!hasChanges && !diff.sandbox_changed && !diff.stack_changed) {
    return (
      <Text variant="subtext" theme="neutral">
        No changes detected
      </Text>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <ComponentList label="Added" entries={diff.added} theme="success" />
      <ComponentList label="Changed" entries={diff.changed} theme="info" />
      <ComponentList label="Removed" entries={diff.removed} theme="error" />
      {diff.sandbox_changed ? (
        <div className="flex items-center gap-1.5">
          <Badge theme="warn" size="sm">
            sandbox changed
          </Badge>
        </div>
      ) : null}
      {diff.stack_changed ? (
        <div className="flex items-center gap-1.5">
          <Badge theme="warn" size="sm">
            stack changed
          </Badge>
        </div>
      ) : null}
    </div>
  )
}

const DiffSummaryBadges = ({ entry }: { entry: IInstallMetadataEntry }) => {
  const hasChanges =
    (entry?.added ?? 0) > 0 ||
    (entry?.changed ?? 0) > 0 ||
    (entry?.removed ?? 0) > 0

  if (!hasChanges && !entry?.sandbox_changed && !entry?.stack_changed) {
    return null
  }

  return (
    <span className="flex items-center gap-1.5">
      {(entry?.added ?? 0) > 0 ? (
        <Badge theme="success" size="sm">
          +{entry.added}
        </Badge>
      ) : null}
      {(entry?.changed ?? 0) > 0 ? (
        <Badge theme="info" size="sm">
          ~{entry.changed}
        </Badge>
      ) : null}
      {(entry?.removed ?? 0) > 0 ? (
        <Badge theme="error" size="sm">
          -{entry.removed}
        </Badge>
      ) : null}
    </span>
  )
}

const PlanInstallCard = ({
  entry,
  planEntry,
  orgId: _orgId,
  isFirst,
}: {
  entry: IInstallMetadataEntry
  planEntry?: IInstallPlanEntry
  orgId: string
  isFirst: boolean
}) => {
  return (
    <Expand
      id={`install-${entry?.install_id}`}
      className="border rounded-lg"
      isOpen={isFirst}
      heading={
        <div className="flex items-center justify-between w-full pr-2">
          <span className="flex items-center gap-3">
            <ID>{entry?.install_id}</ID>
            <Status variant="badge" status={entry?.status} />
            <DiffSummaryBadges entry={entry} />
          </span>
        </div>
      }
    >
      <div className="border-t p-4">
        {planEntry?.diff ? (
          <InstallDiffDetail diff={planEntry.diff} />
        ) : (
          <Text variant="subtext" theme="neutral">
            No diff available
          </Text>
        )}
      </div>
    </Expand>
  )
}

const DeployInstallCard = ({
  entry,
  orgId,
  isFirst,
}: {
  entry: IInstallMetadataEntry
  orgId: string
  isFirst: boolean
}) => {
  return (
    <Expand
      id={`install-${entry?.install_id}`}
      className="border rounded-lg"
      isOpen={isFirst}
      heading={
        <div className="flex items-center justify-between w-full pr-2">
          <span className="flex items-center gap-3">
            <ID>{entry?.install_id}</ID>
            <Status variant="badge" status={entry?.status} />
          </span>
          <span className="flex items-center gap-3">
            {entry?.workflow_id ? (
              <Link
                href={`/${orgId}/installs/${entry.install_id}/workflows/${entry.workflow_id}`}
              >
                <Text variant="subtext" theme="neutral">
                  View workflow
                </Text>
              </Link>
            ) : null}
          </span>
        </div>
      }
    >
      <div className="border-t p-4 flex flex-col gap-2">
        <div className="flex items-center gap-2">
          <Text variant="subtext" theme="neutral">
            Status:
          </Text>
          <Status variant="badge" status={entry?.status} />
        </div>
        {entry?.workflow_id ? (
          <Link
            href={`/${orgId}/installs/${entry.install_id}/workflows/${entry.workflow_id}`}
          >
            <Text variant="subtext">Open workflow</Text>
          </Link>
        ) : null}
      </div>
    </Expand>
  )
}

export const InstallGroupStepDetails = ({
  step,
}: {
  step?: TWorkflowStep
}) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''

  const metadata = step?.status?.metadata as
    | Record<string, unknown>
    | undefined
  const installs = (metadata?.installs ?? []) as IInstallMetadataEntry[]
  const groupName = metadata?.install_group_name as string | undefined
  const totalInstalls = metadata?.total_installs as number | undefined

  const isPlan = step?.name?.startsWith('plan install group')
  const hasApproval = step?.execution_type === 'approval' && !!step?.approval

  const { plan, isLoading: planLoading } = useQueryApprovalPlan({
    step: step!,
  })
  const groupPlan = plan as IInstallGroupPlan | undefined
  const planInstalls = groupPlan?.installs ?? []

  const planByInstallId = new Map<string, IInstallPlanEntry>()
  for (const p of planInstalls) {
    if (p?.install_id) {
      planByInstallId.set(p.install_id, p)
    }
  }

  if (!installs?.length && !planInstalls?.length) {
    return (
      <div className="p-4">
        <Text variant="subtext" theme="neutral">
          {totalInstalls
            ? `Preparing ${totalInstalls} installs...`
            : 'No installs in group'}
        </Text>
      </div>
    )
  }

  const displayInstalls = installs?.length ? installs : planInstalls.map((p) => ({
    install_id: p?.install_id,
    status: 'success',
  }))

  return (
    <div className="p-4 flex flex-col gap-3">
      {groupName ? (
        <div className="flex items-center gap-2 mb-1">
          <Text variant="subtext" weight="strong">
            install group: {groupName}
          </Text>
          <Text variant="subtext" theme="neutral">
            {displayInstalls.length}{' '}
            {displayInstalls.length === 1 ? 'install' : 'installs'}
          </Text>
        </div>
      ) : null}

      {hasApproval && step ? (
        <StepButtons step={step} />
      ) : null}

      {isPlan && planLoading && !plan ? (
        <Skeleton height="200px" width="100%" />
      ) : null}

      {displayInstalls.map((entry, idx) =>
        isPlan ? (
          <PlanInstallCard
            key={entry?.install_id ?? idx}
            entry={entry}
            planEntry={planByInstallId.get(entry?.install_id ?? '')}
            orgId={orgId}
            isFirst={idx === 0}
          />
        ) : (
          <DeployInstallCard
            key={entry?.install_id ?? idx}
            entry={entry}
            orgId={orgId}
            isFirst={idx === 0}
          />
        )
      )}
    </div>
  )
}
