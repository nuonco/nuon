import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { InstallGroupDiff } from '@/components/approvals/plan-diffs/install-group/InstallGroupDiff'
import type { InstallDiffEntry } from '@/components/approvals/plan-diffs/install-group/InstallGroupDiff'

type ApprovalResponse = 'approve' | 'deny' | 'deny-skip-current'

interface IPlanGroupStep {
  installs: any[]
  groupName?: string
  orgId: string
  hasResponse: boolean
  responseType?: string
  showApproveBar: boolean
  isResponding: boolean
  isInProgress: boolean
  onRespond: (response: ApprovalResponse) => void
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
  isResponding,
  isInProgress,
  onRespond,
}: IPlanGroupStep) => {
  return (
    <div className="space-y-3">
      {hasResponse && (
        <div className="flex items-center gap-2 px-4 py-3 rounded-[10px] border border-green-300 dark:border-green-700/40 bg-green-50 dark:bg-green-950/30">
          <Icon variant="CheckCircleIcon" size={18} className="text-green-600 dark:text-green-400 shrink-0" />
          <span className="text-[13px] text-green-700 dark:text-green-300">
            Plan {responseType === 'approve' ? 'approved' : responseType || 'responded'}
          </span>
        </div>
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
        <div className="flex items-center gap-3 px-4 py-3 rounded-[10px] border border-yellow-300 dark:border-yellow-700/40 bg-yellow-50 dark:bg-yellow-950/30">
          <Icon variant="WarningCircleIcon" size={18} className="text-yellow-600 dark:text-yellow-400 shrink-0" />
          <span className="text-[13px] text-yellow-700 dark:text-yellow-300 flex-1">
            Review the changes above and approve to proceed with deployment.
          </span>
          <div className="flex items-center gap-2 shrink-0">
            <Button
              variant="danger"
              size="sm"
              onClick={() => onRespond('deny-skip-current')}
              disabled={isResponding}
            >
              Skip
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={() => onRespond('approve')}
              disabled={isResponding}
            >
              {isResponding ? 'Approving...' : 'Approve'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
