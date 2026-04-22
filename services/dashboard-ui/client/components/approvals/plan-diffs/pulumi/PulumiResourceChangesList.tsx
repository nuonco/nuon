import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import type { TPulumiChangeAction } from '@/types'
import { cn } from '@/utils/classnames'
import { isComplex, isStringJson, semanticEqual } from '@/utils/terraform-utils'
import { TreeDiffValue } from '../terraform/TreeDiffValue'
import {
  PULUMI_ACTION_BADGE_THEME,
  getPulumiActionBgColor,
  getPulumiActionBorderColor,
} from '../diff-style-utils'

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

const DIFF_STYLES = {
  added:
    'bg-green-500/15 dark:bg-green-500/5 text-green-800 dark:text-green-400',
  removed: 'bg-red-500/15 dark:bg-red-500/5 text-red-800 dark:text-red-400',
  changed:
    'bg-orange-500/15 dark:bg-orange-500/5 text-orange-800 dark:text-orange-400',
  unchanged: '',
}

function getDiffPrefix(kind: string) {
  switch (kind) {
    case 'add':
      return { char: '+', style: DIFF_STYLES.added }
    case 'delete':
      return { char: '-', style: DIFF_STYLES.removed }
    case 'update':
    default:
      return { char: '~', style: DIFF_STYLES.changed }
  }
}

function getActionPrefix(action: string) {
  switch (action) {
    case 'create':
    case 'create-replacement':
      return { char: '+', style: DIFF_STYLES.added }
    case 'delete':
    case 'delete-replaced':
      return { char: '-', style: DIFF_STYLES.removed }
    case 'update':
    case 'replace':
    default:
      return { char: '~', style: DIFF_STYLES.changed }
  }
}

function formatValue(val: unknown): string {
  if (val === null || typeof val === 'undefined') return 'null'
  if (val === '') return '""'
  if (typeof val === 'object') return JSON.stringify(val, null, 2)
  return String(val)
}

function DetailedDiffBody({
  detailedDiff,
}: {
  detailedDiff: Record<string, IPropertyDiff>
}) {
  return (
    <div className="p-4 bg-code border-t shadow-xs min-h-[3rem] max-h-[40rem] overflow-auto font-mono text-[13px] leading-6">
      {Object.entries(detailedDiff).map(([prop, diff]) => {
        const prefix = getDiffPrefix(diff.kind)
        return (
          <div className={cn('flex whitespace-pre', prefix.style)} key={prop}>
            <span className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70">
              {prefix.char}
            </span>
            <span>
              <span className="font-semibold">{prop}</span>
              <span className="opacity-60">
                {'  '}
                {diff.kind}
                {diff.inputDiff ? ' (input)' : ''}
              </span>
            </span>
          </div>
        )
      })}
    </div>
  )
}

function InputsDiffBody({
  action,
  oldInputs,
  newInputs,
}: {
  action: string
  oldInputs?: Record<string, unknown>
  newInputs?: Record<string, unknown>
}) {
  const allKeys = [
    ...new Set([
      ...Object.keys(oldInputs || {}),
      ...Object.keys(newInputs || {}),
    ]),
  ]

  const keyValues = allKeys.map((key) => {
    const before = oldInputs?.[key] ?? null
    const after = newInputs?.[key] ?? null
    const changed = !semanticEqual(before, after)
    return { key, before, after, changed }
  })

  return (
    <div className="p-4 bg-code border-t shadow-xs min-h-[3rem] max-h-[40rem] overflow-auto font-mono text-[13px] leading-6">
      {keyValues.length ? (
        keyValues.map((value, idx) => {
          const prefix = value.changed
            ? getActionPrefix(action)
            : { char: ' ', style: DIFF_STYLES.unchanged }

          const hasComplexValue =
            isComplex(value.before) ||
            isComplex(value.after) ||
            (typeof value.before === 'string' && isStringJson(value.before)) ||
            (typeof value.after === 'string' && isStringJson(value.after))

          if (hasComplexValue) {
            return (
              <div key={value.key + idx}>
                <div className={cn('flex whitespace-pre', prefix.style)}>
                  <span className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70">
                    {prefix.char}
                  </span>
                  <span className="font-semibold">{value.key}:</span>
                </div>
                <TreeDiffValue before={value.before} after={value.after} />
              </div>
            )
          }

          const formattedBefore = formatValue(value.before)
          const formattedAfter = formatValue(value.after)

          return (
            <div
              className={cn('flex whitespace-pre', prefix.style)}
              key={value.key + idx}
            >
              <span className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70">
                {prefix.char}
              </span>
              <span>
                <span className="font-semibold">{value.key}:</span>
                {value.changed ? (
                  <>
                    {'  '}
                    <span
                      className="text-red-800 dark:text-red-400 line-through opacity-70 inline-block max-w-[300px] truncate align-bottom"
                      title={formattedBefore}
                    >
                      {formattedBefore}
                    </span>
                    <span className="opacity-50">{' -> '}</span>
                    <span
                      className="inline-block max-w-[300px] truncate align-bottom"
                      title={formattedAfter}
                    >
                      {formattedAfter}
                    </span>
                  </>
                ) : (
                  <>
                    {'  '}
                    <span
                      className="inline-block max-w-[500px] truncate align-bottom"
                      title={formattedAfter}
                    >
                      {formattedAfter}
                    </span>
                  </>
                )}
              </span>
            </div>
          )
        })
      ) : (
        <Text family="mono">No values to display.</Text>
      )}
    </div>
  )
}

export function PulumiResourceChangesList({
  changes,
}: {
  changes: IResourceChange[]
}) {
  return (
    <div className="flex flex-col divide-y">
      {changes.length ? (
        changes.map((change, idx) => {
          const action = change.action as TPulumiChangeAction
          const bgColor = getPulumiActionBgColor(action)
          const borderColor = getPulumiActionBorderColor(action)

          const hasDiffs =
            change.detailed_diff && Object.keys(change.detailed_diff).length > 0
          const hasInputChanges = change.old_inputs || change.new_inputs

          return (
            <Expand
              key={`${change.urn}-${idx}`}
              id={change.urn}
              className={`border-l-4 ${borderColor}`}
              headerClassName={`w-full px-4 py-3 gap-3 text-left focus:outline-none ${bgColor}`}
              heading={
                <div className="text-left w-full">
                  <div className="flex items-start justify-between w-full">
                    <div className="flex flex-col max-w-[500px]">
                      <Text
                        nowrap
                        className="block truncate font-mono"
                        weight="strong"
                      >
                        {change.type}
                      </Text>
                      <Text variant="subtext" theme="neutral">
                        {change.name}
                      </Text>
                      <Text
                        variant="subtext"
                        theme="neutral"
                        className="font-mono truncate"
                      >
                        {change.urn}
                      </Text>
                    </div>

                    <div className="flex items-center pr-4 self-center">
                      <Badge
                        theme={PULUMI_ACTION_BADGE_THEME[action] || 'neutral'}
                        size="sm"
                      >
                        {change.action}
                      </Badge>
                    </div>
                  </div>
                </div>
              }
            >
              {hasInputChanges ? (
                <InputsDiffBody
                  action={change.action}
                  oldInputs={change.old_inputs}
                  newInputs={change.new_inputs}
                />
              ) : hasDiffs ? (
                <DetailedDiffBody detailedDiff={change.detailed_diff!} />
              ) : (
                <div className="p-4 bg-code border-t shadow-xs font-mono text-[13px] leading-6">
                  <Text family="mono">No values to display.</Text>
                </div>
              )}
            </Expand>
          )
        })
      ) : (
        <div className="px-4 py-3 text-center">
          <EmptyState
            emptyMessage="Try clearing the search term or resetting the filter"
            emptyTitle="No changes to show"
            variant="search"
            size="sm"
          />
        </div>
      )}
    </div>
  )
}
