import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { DiffMarker, InstallStatusIcon } from '../../shared/icons'

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

export const PlanGroupStep = ({
  installs,
  groupName,
  orgId,
  hasResponse,
  responseType,
  showApproveBar,
  isResponding,
  isInProgress,
  onRespond,
}: IPlanGroupStep) => {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon variant="ListChecksIcon" size={16} className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
          <span className="text-[13px] text-cool-grey-600 dark:text-cool-grey-300">
            install group:{' '}
            <span className="font-semibold text-cool-grey-900 dark:text-white">{groupName}</span>
          </span>
          <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">
            {installs.length} {installs.length === 1 ? 'install' : 'installs'}
          </span>
        </div>
      </div>

      {hasResponse && (
        <div className="flex items-center gap-2 px-4 py-3 rounded-[10px] border border-green-300 dark:border-green-700/40 bg-green-50 dark:bg-green-950/30">
          <Icon variant="CheckCircleIcon" size={18} className="text-green-600 dark:text-green-400 shrink-0" />
          <span className="text-[13px] text-green-700 dark:text-green-300">
            Plan {responseType === 'approve' ? 'approved' : responseType || 'responded'}
          </span>
        </div>
      )}

      {installs.length > 0 && (
        <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
          {installs.map((inst: any, i: number) => (
            <PlanInstallRow key={inst.install_id || i} install={inst} orgId={orgId} />
          ))}
        </div>
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

const PlanInstallRow = ({ install, orgId }: { install: any; orgId: string }) => {
  const [expanded, setExpanded] = useState(false)
  const diff = install.diff

  const added = Array.isArray(diff?.added) ? diff.added : []
  const changed = Array.isArray(diff?.changed) ? diff.changed : []
  const removed = Array.isArray(diff?.removed) ? diff.removed : []

  const addedCount = added.length || (install.added ?? 0)
  const changedCount = changed.length || (install.changed ?? 0)
  const removedCount = removed.length || (install.removed ?? 0)
  const sandboxChanged = diff?.sandbox_changed || install.sandbox_changed
  const stackChanged = diff?.stack_changed || install.stack_changed
  const hasChanges = addedCount > 0 || changedCount > 0 || removedCount > 0 || sandboxChanged || stackChanged
  const hasDetailedDiff = added.length > 0 || changed.length > 0 || removed.length > 0

  const installLabels = install.install_labels as Record<string, string> | undefined
  const labelEntries = installLabels ? Object.entries(installLabels) : []

  return (
    <div>
      <button
        className="flex items-center gap-3 px-4 py-3 w-full text-left hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <svg
          width="12" height="12" viewBox="0 0 12 12" fill="none"
          className={`text-cool-grey-400 shrink-0 transition-transform ${expanded ? 'rotate-90' : ''}`}
        >
          <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>

        <InstallStatusIcon status={install.status} />

        <div className="flex items-center gap-2 min-w-0">
          <span className="text-[13.5px] font-semibold text-cool-grey-900 dark:text-white truncate">
            {install.install_name || install.install_id}
          </span>
          {install.install_id && (
            <a
              href={`/${orgId}/installs/${install.install_id}`}
              onClick={(e) => e.stopPropagation()}
              className="text-cool-grey-400 hover:text-primary-500 dark:text-cool-grey-500 dark:hover:text-primary-400 shrink-0"
            >
              <Icon variant="ArrowSquareOutIcon" size={14} />
            </a>
          )}
          {labelEntries.map(([k, v]) => (
            <span key={k} className="inline-flex items-center px-1.5 py-0.5 rounded border border-cool-grey-200 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 font-mono text-[10.5px] text-cool-grey-500 dark:text-cool-grey-400 shrink-0">
              {k}={v}
            </span>
          ))}
        </div>

        <div className="flex items-center gap-1.5 ml-auto shrink-0">
          {addedCount > 0 && (
            <span className="text-[12px] font-semibold text-green-600 dark:text-green-400">+{addedCount}</span>
          )}
          {changedCount > 0 && (
            <span className="text-[12px] font-semibold text-yellow-600 dark:text-yellow-400">~{changedCount}</span>
          )}
          {removedCount > 0 && (
            <span className="text-[12px] font-semibold text-red-600 dark:text-red-400">-{removedCount}</span>
          )}
          {sandboxChanged && (
            <Badge theme="warn" size="sm">sandbox</Badge>
          )}
          {stackChanged && (
            <Badge theme="warn" size="sm">stack</Badge>
          )}
          {!hasChanges && (
            <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">no changes</span>
          )}
        </div>
      </button>

      {expanded && hasChanges && hasDetailedDiff && (
        <div className="px-4 pb-3 pl-[52px] space-y-1">
          {added.map((c: any) => (
            <div key={c.component_id} className="flex items-center gap-2">
              <DiffMarker op="add" />
              <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200">{c.component_name || c.component_id}</span>
              {c.component_type && (
                <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-500">{c.component_type}</span>
              )}
            </div>
          ))}
          {changed.map((c: any) => (
            <div key={c.component_id} className="flex items-center gap-2">
              <DiffMarker op="change" />
              <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200">{c.component_name || c.component_id}</span>
              {c.component_type && (
                <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-500">{c.component_type}</span>
              )}
            </div>
          ))}
          {removed.map((c: any) => (
            <div key={c.component_id} className="flex items-center gap-2">
              <DiffMarker op="remove" />
              <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200">{c.component_name || c.component_id}</span>
              {c.component_type && (
                <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-500">{c.component_type}</span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
