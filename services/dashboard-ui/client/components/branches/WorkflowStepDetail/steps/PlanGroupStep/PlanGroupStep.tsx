import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { InstallGroupDiff } from '@/components/approvals/plan-diffs/install-group/InstallGroupDiff'
import type { InstallDiffEntry } from '@/components/approvals/plan-diffs/install-group/InstallGroupDiff'

interface IPlanGroupStep {
  installs: any[]
  groupName?: string
  orgId: string
  hasResponse: boolean
  responseType?: string
  showApproveBar: boolean
  isInProgress: boolean
  actions?: ReactNode
}

function transformInstalls(installs: any[]): InstallDiffEntry[] {
  return installs.map((inst) => {
    const diff = inst.diff
    const added = Array.isArray(diff?.added) ? diff.added : []
    const changed = Array.isArray(diff?.changed) ? diff.changed : []
    const removed = Array.isArray(diff?.removed) ? diff.removed : []

    const componentEntities = [
      ...added.map((c: any) => ({
        name: c.component_name || c.component_id,
        op: 'add' as const,
        componentType: c.component_type,
        fields: [{ key: 'type', op: 'add', diff: `'' -> '${c.component_type || ''}'` }],
      })),
      ...changed.map((c: any) => ({
        name: c.component_name || c.component_id,
        op: 'change' as const,
        componentType: c.component_type,
        fields: [{ key: 'config', op: 'change', diff: 'configuration changed' }],
      })),
      ...removed.map((c: any) => ({
        name: c.component_name || c.component_id,
        op: 'remove' as const,
        componentType: c.component_type,
        fields: [{ key: 'type', op: 'remove', diff: `'${c.component_type || ''}' -> ''` }],
      })),
    ]

    const sections = componentEntities.length > 0
      ? [{
          name: 'Components',
          sectionKey: 'components',
          grouped: true,
          additions: added.length,
          removals: removed.length,
          changed: changed.length,
          entities: componentEntities,
          fields: [],
        }]
      : []

    return {
      installId: inst.install_id || inst.install_name,
      installName: inst.install_name || inst.install_id,
      installLabels: inst.install_labels,
      status: inst.status,
      sandboxChanged: diff?.sandbox_changed || inst.sandbox_changed,
      stackChanged: diff?.stack_changed || inst.stack_changed,
      summary: {
        added: added.length,
        removed: removed.length,
        changed: changed.length,
      },
      sections,
    }
  })
}

export const PlanGroupStep = ({
  installs,
  groupName,
  orgId: _orgId,
  hasResponse,
  responseType,
  showApproveBar,
  isInProgress,
  actions,
}: IPlanGroupStep) => {
  return (
    <div className="space-y-3">
      {hasResponse && (
        <Banner theme="success">
          <Text weight="strong">
            Plan {responseType === 'approve' ? 'approved' : responseType || 'responded'}
          </Text>
        </Banner>
      )}

      {installs.length > 0 && (
        <InstallGroupDiff
          groupName={groupName || 'install group'}
          installs={transformInstalls(installs)}
        />
      )}

      {installs.length === 0 && (
        <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
          <Text variant="subtext" theme="neutral">
            {isInProgress ? 'Computing install diffs...' : 'Waiting to compute plan...'}
          </Text>
        </div>
      )}

      {showApproveBar && (
        <Banner className="@container" theme="warn">
          <div className="flex flex-col gap-2">
            <div className="flex flex-col">
              <Text weight="strong">Install group plan requires review</Text>
              <Text variant="subtext" theme="neutral">
                Review the changes above, then approve to deploy or skip this install group.
              </Text>
            </div>
            {actions && <div className="flex self-end gap-2">{actions}</div>}
          </div>
        </Banner>
      )}
    </div>
  )
}
