import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { ComponentType } from '@/components/components/ComponentType'
import {
  InstallComponentsTable,
  type InstallComponentRow,
} from '@/components/install-components/InstallComponentsTable/InstallComponentsTable'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type {
  TCustomerManagedSnapshotRun,
  TCustomerManagedSnapshotRunStep,
} from '@/lib/ctl-api/installs/customer-managed-support-snapshots'
import type { TComponentType } from '@/types'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

const empty = <Icon variant="MinusIcon" />

function getStringArray(value: unknown): string[] {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === 'string')
    : []
}

function getLabels(value: unknown): Record<string, string> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  return Object.fromEntries(
    Object.entries(value).filter(
      (entry): entry is [string, string] => typeof entry[1] === 'string'
    )
  )
}

function latestComponentStep(
  runs: TCustomerManagedSnapshotRun[],
  componentName: string
): TCustomerManagedSnapshotRunStep | undefined {
  const stepId = `deploy-${componentName}-apply`
  return runs
    .flatMap((run) => run.steps)
    .filter((step) => step.id === stepId)
    .sort((a, b) =>
      (b.finished_at ?? b.started_at ?? '').localeCompare(
        a.finished_at ?? a.started_at ?? ''
      )
    )[0]
}

function latestDriftRun(
  runs: TCustomerManagedSnapshotRun[],
  componentName: string
): TCustomerManagedSnapshotRun | undefined {
  return runs
    .filter((run) => run.ref_kind === 'drift' && run.ref_name === componentName)
    .sort((a, b) => b.started_at.localeCompare(a.started_at))[0]
}

function capturedDeployStatus(
  step?: TCustomerManagedSnapshotRunStep
): ReactNode {
  if (!step) return empty
  const skipped = ['skipped', 'user-skipped'].includes(
    step.result_directive ?? ''
  )
  const status = skipped
    ? step.result_directive
    : step.status === 'finished'
      ? 'active'
      : step.status
  const badge = <Status variant="badge" status={status} />
  const observedAt = step.finished_at ?? step.started_at
  if (!observedAt) return badge
  return (
    <Tooltip
      position="top"
      tipContent={<Time time={observedAt} format="long-datetime" />}
    >
      {badge}
    </Tooltip>
  )
}

function capturedDriftStatus(run?: TCustomerManagedSnapshotRun): ReactNode {
  const drift = run?.steps.find((step) => step.drift)?.drift
  if (!drift || typeof drift.drifted !== 'boolean') return empty
  return drift.drifted ? <Status variant="badge" status="drifted" /> : empty
}

export const CustomerManagedSnapshotComponents = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const runs = snapshot?.snapshot.runs ?? []
  const health = new Map(
    (snapshot?.snapshot.health?.components ?? []).map((component) => [
      component.component_name,
      component,
    ])
  )
  const components: InstallComponentRow[] = (
    snapshot?.snapshot.active_bundle?.contents ?? []
  )
    .filter(({ kind }) => kind === 'component')
    .map((component) => {
      const capturedHealth = health.get(component.name)
      const definition = component.component_definition ?? {}
      const dependencies = getStringArray(definition.dependencies)
      const labels = getLabels(definition.labels)
      const isToggleable = definition.toggleable === true
      const isEnabled = definition.default_enabled !== false
      return {
        componentId:
          capturedHealth?.component_id ?? capturedHealth?.install_component_id,
        componentName: component.name,
        componentType: (
          <ComponentType
            type={
              (capturedHealth?.component_type ??
                component.detail ??
                'unknown') as TComponentType
            }
            variant="subtext"
            colorVariant="color"
          />
        ),
        health: capturedHealth?.health ?? 'unknown',
        healthMessage: capturedHealth?.resources?.length
          ? `${capturedHealth.resources.length} resources captured`
          : undefined,
        href: `/${org.id}/installs/${install.id}/components/${capturedHealth?.component_id ?? component.name}`,
        toggleStatus: isToggleable ? (
          <Badge size="sm" theme={isEnabled ? 'success' : 'neutral'}>
            {isEnabled ? 'Enabled' : 'Disabled'}
          </Badge>
        ) : (
          empty
        ),
        overrideStatus: empty,
        deployStatus: capturedDeployStatus(
          latestComponentStep(runs, component.name)
        ),
        driftStatus: capturedDriftStatus(latestDriftRun(runs, component.name)),
        dependencies: dependencies.length ? (
          <Tooltip
            position="top"
            tipContent={
              <Text as="div" className="p-2" variant="subtext">
                {dependencies.join(', ')}
              </Text>
            }
          >
            <Badge variant="code">{dependencies.length}</Badge>
          </Tooltip>
        ) : (
          empty
        ),
        labels: Object.keys(labels).length ? (
          <span className="flex flex-wrap gap-1">
            {Object.entries(labels).map(([key, value]) => (
              <LabelBadge
                key={key}
                labelKey={key}
                labelValue={value}
                size="sm"
              />
            ))}
          </span>
        ) : (
          empty
        ),
        action: null,
      }
    })

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install components
        </Text>
        <Text variant="subtext" theme="neutral">
          View component inventory and health captured in this support snapshot.
        </Text>
      </HeadingGroup>
      <InstallComponentsTable
        data={components}
        filterActions={null}
        pagination={{ hasNext: false, offset: 0, limit: components.length }}
        isLoading={false}
        showHealth
      />
    </CustomerManagedSnapshotContent>
  )
}
