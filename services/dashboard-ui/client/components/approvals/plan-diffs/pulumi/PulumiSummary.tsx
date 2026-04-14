import { Text } from '@/components/common/Text'
import type { TTextTheme } from '@/components/common/Text'
import type { TPulumiChangeAction } from '@/types'

const ACTION_THEME: Record<TPulumiChangeAction, { theme: TTextTheme; label: string }> = {
  create: { theme: 'success', label: 'to create' },
  update: { theme: 'warn', label: 'to update' },
  delete: { theme: 'error', label: 'to delete' },
  replace: { theme: 'brand', label: 'to replace' },
  'create-replacement': { theme: 'brand', label: 'to create (replacement)' },
  'delete-replaced': { theme: 'brand', label: 'to delete (replaced)' },
  same: { theme: 'neutral', label: 'unchanged' },
  read: { theme: 'info', label: 'to read' },
  refresh: { theme: 'info', label: 'to refresh' },
}

export function PulumiSummary({
  changeSummary,
}: {
  changeSummary: Record<string, number>
}) {
  const entries = Object.entries(changeSummary).filter(([, count]) => count > 0)
  if (entries.length === 0) {
    return (
      <div className="px-4 py-3 sm:px-6 border-b bg-cool-grey-100 dark:bg-dark-grey-800">
        <Text variant="subtext">No changes detected</Text>
      </div>
    )
  }

  return (
    <div className="px-4 py-3 sm:px-6 border-b bg-cool-grey-100 dark:bg-dark-grey-800">
      <div className="flex space-x-4">
        {entries.map(([action, count]) => {
          const config = ACTION_THEME[action as TPulumiChangeAction]
          return (
            <div key={action} className="flex items-center gap-1.5">
              <Text variant="base" theme={config?.theme || 'neutral'} weight="strong">
                {count}
              </Text>
              <Text variant="subtext" theme="neutral">
                {config?.label || action}
              </Text>
            </div>
          )
        })}
      </div>
    </div>
  )
}
