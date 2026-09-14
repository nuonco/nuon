import { Badge } from '@/components/common/Badge'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { humanize } from '@/utils/string-utils'
import type {
  TInstallUpdate,
  TInstallUpdateComponentDiff,
  TInstallUpdateImpactReason,
} from '@/types'

export interface IInstallUpdateDetails extends IPanel {
  update: TInstallUpdate
  orgId?: string
  installId?: string
  appId?: string
}

const ImpactReasonsRow = ({
  reasons,
}: {
  reasons?: TInstallUpdateImpactReason[]
}) => {
  if (!reasons?.length) return null
  return (
    <div className="flex flex-wrap gap-1.5 mt-1">
      {reasons.map((reason) => (
        <span
          key={`${reason.from}-${reason.edge}`}
          className="flex items-center gap-1"
        >
          <Badge size="sm" variant="code">
            {reason.from}
          </Badge>
          <Text variant="subtext" theme="neutral">
            via {humanize(reason.edge)}
          </Text>
        </span>
      ))}
    </div>
  )
}

// ComputeInstallConfigDiff promotes graph-impacted components into `changed` even when
// their own config is byte-identical. Calling that "Changed" misreads as an edit, so
// matching checksums with no build change are labeled "Impacted" instead.
const operationLabel = (
  component: TInstallUpdateComponentDiff,
  operation: 'added' | 'changed' | 'removed'
): string => {
  const checksumsMatch =
    !!component.old_checksum &&
    component.old_checksum === component.new_checksum
  if (operation === 'changed' && checksumsMatch && !component.build_changed) {
    return 'Impacted'
  }
  return humanize(operation)
}

const ComponentImpactRow = ({
  component,
  operation,
}: {
  component: TInstallUpdateComponentDiff
  operation: 'added' | 'changed' | 'removed'
}) => (
  <div className="flex flex-col gap-0.5 py-2">
    <div className="flex items-center gap-2">
      <Text variant="subtext" weight="strong" family="mono">
        {component.component_name || component.component_id}
      </Text>
      {component.component_type ? (
        <Badge size="sm">{humanize(component.component_type)}</Badge>
      ) : null}
      <Badge size="sm" theme={operation === 'removed' ? 'warn' : 'info'}>
        {operationLabel(component, operation)}
      </Badge>
      {component.build_changed ? (
        <Badge size="sm" theme="neutral">
          Build changed
        </Badge>
      ) : null}
    </div>
    <ImpactReasonsRow reasons={component.impact_reasons} />
  </div>
)

export const installUpdateStatus = (update: TInstallUpdate): string =>
  update.app_config?.version?.status?.status ||
  update.stack?.status?.status ||
  update.install_config?.version?.status?.status ||
  (update.type === 'inputs' ? 'success' : 'unknown')

export const InstallUpdateDetails = ({
  update,
  orgId,
  installId,
  appId,
  ...props
}: IInstallUpdateDetails) => {
  const version = update.app_config?.version
  const branchRun = version?.app_branch_run
  const commit = branchRun?.vcs_connection_commit
  const diff = update.app_config?.diff
  const stackImpacts = diff?.stack_impacts ?? []
  const componentImpacts = [
    ...(diff?.added ?? []).map((component) => ({
      component,
      operation: 'added' as const,
    })),
    ...(diff?.changed ?? []).map((component) => ({
      component,
      operation: 'changed' as const,
    })),
    ...(diff?.removed ?? []).map((component) => ({
      component,
      operation: 'removed' as const,
    })),
  ]

  return (
    <Panel heading="Update details" size="half" {...props}>
      <div className="flex items-center gap-2">
        <Status variant="badge" status={installUpdateStatus(update)} />
        {version?.new_app_config_id ? (
          <Badge size="sm" variant="code">
            {version.new_app_config_id}
          </Badge>
        ) : null}
        <Badge size="sm" theme="neutral">
          {humanize(update.type)}
        </Badge>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Applied">
          {update.created_at ? (
            <Time
              variant="subtext"
              time={update.created_at}
              format="relative"
            />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="Update ID">
          <ID>{update.id}</ID>
        </LabeledValue>
      </div>

      {commit ? (
        <LabeledValue label="Commit">
          <span className="flex flex-wrap items-center gap-2">
            {branchRun?.app_branch?.name ? (
              <Badge size="sm" theme="info">
                {branchRun.app_branch.name}
              </Badge>
            ) : null}
            {commit.sha ? (
              <Badge size="sm" variant="code">
                {commit.sha.slice(0, 7)}
              </Badge>
            ) : null}
            <Text variant="subtext" theme="neutral">
              {commit.message}
            </Text>
            {branchRun?.pr_number ? (
              <Badge size="sm" theme="neutral">
                PR #{branchRun.pr_number}
              </Badge>
            ) : null}
          </span>
        </LabeledValue>
      ) : null}

      <div className="flex flex-wrap items-center gap-4">
        {orgId &&
        appId &&
        version?.app_branch_run_id &&
        branchRun?.app_branch?.id ? (
          <Link
            href={`/${orgId}/apps/${appId}/branches/${branchRun.app_branch.id}/runs/${version.app_branch_run_id}`}
          >
            View branch run
          </Link>
        ) : null}
        {orgId && installId && update?.workflow_id ? (
          <Link
            href={`/${orgId}/installs/${installId}/history/${update.workflow_id}`}
          >
            View workflow
          </Link>
        ) : null}
      </div>

      {stackImpacts.length > 0 ? (
        <>
          <Divider dividerWord="Stack impacts" />
          <div className="flex flex-wrap gap-2">
            {stackImpacts.map((impact) => (
              <Badge key={impact} size="sm" theme="info">
                {humanize(impact)}
              </Badge>
            ))}
          </div>
          <ImpactReasonsRow reasons={diff?.stack_impact_reasons} />
        </>
      ) : null}

      {componentImpacts.length > 0 ? (
        <>
          <Divider dividerWord="Component impacts" />
          <div className="flex flex-col divide-y">
            {componentImpacts.map(({ component, operation }) => (
              <ComponentImpactRow
                key={`${operation}-${component.component_id}`}
                component={component}
                operation={operation}
              />
            ))}
          </div>
        </>
      ) : null}

      {update.inputs?.keys?.length ? (
        <>
          <Divider dividerWord="Inputs" />
          <div className="flex flex-wrap gap-2">
            {update.inputs.keys.map((key) => (
              <Badge key={key} size="sm" variant="code">
                {key}
              </Badge>
            ))}
          </div>
        </>
      ) : null}
    </Panel>
  )
}
