import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { InstallGroupDiff } from '@/components/approvals/plan-diffs/install-group/InstallGroupDiff'
import type { InstallDiffEntry } from '@/components/approvals/plan-diffs/install-group/InstallGroupDiff'

interface IPlanGroupStep {
  installs: any[]
  groupName?: string
  orgId: string
  labelColors?: Record<string, string>
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
        fields: [],
      })),
      ...removed.map((c: any) => ({
        name: c.component_name || c.component_id,
        op: 'remove' as const,
        componentType: c.component_type,
        fields: [{ key: 'type', op: 'remove', diff: `'${c.component_type || ''}' -> ''` }],
      })),
    ]

    const sandboxChanged = diff?.sandbox_changed || inst.sandbox_changed
    const stackChanged = diff?.stack_changed || inst.stack_changed

    const infraEntities = [
      ...(sandboxChanged ? [{ name: 'Sandbox', op: 'change' as const, fields: [] }] : []),
      ...(stackChanged ? [{ name: 'Stack', op: 'change' as const, fields: [] }] : []),
    ]

    const sections = [
      ...(componentEntities.length > 0
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
        : []),
      ...(infraEntities.length > 0
        ? [{
            name: 'Infrastructure',
            sectionKey: 'infrastructure',
            grouped: true,
            additions: 0,
            removals: 0,
            changed: infraEntities.length,
            entities: infraEntities,
            fields: [],
          }]
        : []),
    ]

    return {
      installId: inst.install_id || inst.install_name,
      installName: inst.install_name || inst.install_id,
      installLabels: inst.install_labels,
      status: inst.status,
      sandboxChanged,
      stackChanged,
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
  labelColors,
  hasResponse,
  responseType,
  showApproveBar,
  isInProgress,
  actions,
}: IPlanGroupStep) => {
  return (
    <div className="flex flex-col gap-3">
      {hasResponse && (
        <Banner theme="success">
          <Text weight="strong">
            Plan {responseType === 'approve' ? 'approved' : responseType || 'responded'}
          </Text>
        </Banner>
      )}

      <InstallGroupDiff
        groupName={groupName || 'install group'}
        installs={transformInstalls(installs)}
        isLoading={isInProgress}
        labelColors={labelColors}
      />

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
