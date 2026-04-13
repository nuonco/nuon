import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'

interface IPropertyDiff {
  kind: string
  inputDiff: boolean
}

interface IResourceChange {
  urn: string
  type: string
  name: string
  action: string
  diffs?: string[]
  detailed_diff?: Record<string, IPropertyDiff>
  old_inputs?: Record<string, unknown>
  new_inputs?: Record<string, unknown>
  provider?: string
}

interface IPulumiPreviewResult {
  stdout: string
  stderr: string
  change_summary: Record<string, number>
  resource_changes?: IResourceChange[]
  diagnostics?: string[]
}

const ACTION_CONFIG: Record<
  string,
  { label: string; badge: string; className: string }
> = {
  create: {
    label: 'Create',
    badge: '+',
    className: 'text-green-600 dark:text-green-400',
  },
  update: {
    label: 'Update',
    badge: '~',
    className: 'text-amber-600 dark:text-amber-400',
  },
  delete: {
    label: 'Delete',
    badge: '-',
    className: 'text-red-600 dark:text-red-400',
  },
  replace: {
    label: 'Replace',
    badge: '±',
    className: 'text-orange-600 dark:text-orange-400',
  },
  'create-replacement': {
    label: 'Create replacement',
    badge: '+',
    className: 'text-orange-600 dark:text-orange-400',
  },
  'delete-replaced': {
    label: 'Delete replaced',
    badge: '-',
    className: 'text-orange-600 dark:text-orange-400',
  },
  same: {
    label: 'Unchanged',
    badge: '=',
    className: 'text-grey-500 dark:text-grey-400',
  },
  read: {
    label: 'Read',
    badge: '→',
    className: 'text-blue-600 dark:text-blue-400',
  },
  refresh: {
    label: 'Refresh',
    badge: '↻',
    className: 'text-blue-600 dark:text-blue-400',
  },
}

function PulumiSummary({
  changeSummary,
}: {
  changeSummary: Record<string, number>
}) {
  const entries = Object.entries(changeSummary).filter(([, count]) => count > 0)
  if (entries.length === 0) {
    return (
      <div className="px-4 sm:px-6 py-3 border-b">
        <Text variant="subtext">No changes detected</Text>
      </div>
    )
  }

  return (
    <div className="px-4 sm:px-6 py-3 border-b flex gap-4 flex-wrap">
      {entries.map(([action, count]) => {
        const config = ACTION_CONFIG[action]
        return (
          <Text key={action} variant="subtext">
            <span className={config?.className || ''}>
              {count} to {(config?.label || action).toLowerCase()}
            </span>
          </Text>
        )
      })}
    </div>
  )
}

function ResourceChangeRow({ change }: { change: IResourceChange }) {
  const [expanded, setExpanded] = useState(false)
  const config = ACTION_CONFIG[change.action] || {
    label: change.action,
    badge: '?',
    className: '',
  }

  const hasDiffs =
    change.detailed_diff && Object.keys(change.detailed_diff).length > 0
  const hasInputChanges = change.old_inputs || change.new_inputs

  return (
    <div className="border-b last:border-b-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-4 sm:px-6 py-3 flex items-center gap-3 text-left hover:bg-cool-grey-100 dark:hover:bg-dark-grey-800 transition-colors"
        type="button"
      >
        <span
          className={`font-mono text-sm font-bold w-5 text-center ${config.className}`}
        >
          {config.badge}
        </span>
        <div className="flex-1 min-w-0">
          <Text variant="subtext" className="truncate">
            <span className={config.className}>{config.label}</span>{' '}
            <span className="font-mono">{change.type}</span>{' '}
            <span className="font-semibold">{change.name}</span>
          </Text>
        </div>
        {(hasDiffs || hasInputChanges) && (
          <span className="text-grey-400 text-xs">{expanded ? '▼' : '▶'}</span>
        )}
      </button>

      {expanded && (hasDiffs || hasInputChanges) && (
        <div className="px-4 sm:px-6 pb-3 pl-12">
          {hasDiffs && (
            <div className="space-y-1">
              {Object.entries(change.detailed_diff!).map(([prop, diff]) => (
                <div key={prop} className="font-mono text-xs">
                  <span className="text-grey-500">{prop}: </span>
                  <span
                    className={
                      diff.kind === 'add'
                        ? 'text-green-600 dark:text-green-400'
                        : diff.kind === 'delete'
                          ? 'text-red-600 dark:text-red-400'
                          : 'text-amber-600 dark:text-amber-400'
                    }
                  >
                    {diff.kind}
                    {diff.inputDiff ? ' (input)' : ''}
                  </span>
                </div>
              ))}
            </div>
          )}

          {!hasDiffs && hasInputChanges && (
            <pre className="text-xs font-mono whitespace-pre-wrap text-grey-600 dark:text-grey-400 max-h-64 overflow-y-auto">
              {JSON.stringify(change.new_inputs || change.old_inputs, null, 2)}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}

export function PulumiDiff({
  plan,
}: {
  plan: IPulumiPreviewResult | undefined
}) {
  if (!plan) {
    return <Banner theme="neutral">No Pulumi preview data available</Banner>
  }

  const hasResourceChanges =
    plan.resource_changes && plan.resource_changes.length > 0

  return (
    <div className="flex flex-col gap-6">
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4 border-b">
          <Text variant="base" weight="strong">
            Pulumi preview
          </Text>
        </div>

        {plan.change_summary && (
          <PulumiSummary changeSummary={plan.change_summary} />
        )}

        {hasResourceChanges ? (
          <div>
            {plan.resource_changes!.map((change) => (
              <ResourceChangeRow key={change.urn} change={change} />
            ))}
          </div>
        ) : (
          plan.stdout && (
            <div className="px-4 sm:px-6 py-4">
              <pre className="text-xs font-mono whitespace-pre-wrap overflow-x-auto text-grey-700 dark:text-grey-300">
                {plan.stdout}
              </pre>
            </div>
          )
        )}

        {plan.diagnostics && plan.diagnostics.length > 0 && (
          <div className="px-4 sm:px-6 py-4 border-t">
            <Text variant="subtext" weight="strong" className="mb-2">
              Diagnostics
            </Text>
            {plan.diagnostics.map((d, i) => (
              <pre
                key={i}
                className="text-xs font-mono whitespace-pre-wrap text-amber-700 dark:text-amber-300"
              >
                {d}
              </pre>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
