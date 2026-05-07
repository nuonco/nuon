import { Text } from '@/components/common/Text'
import type { TTerraformOutputChange } from '@/types'
import { cn } from '@/utils/classnames'
import { deepEqual, isComplex, isStringJson } from '@/utils/terraform-utils'
import { DiffLineExpandButton } from '../DiffLineExpandModal'
import { TreeDiffValue } from './TreeDiffValue'

type TTerraformValues = Pick<
  TTerraformOutputChange,
  'before' | 'after' | 'action'
>

const DIFF_STYLES = {
  added: 'bg-green-500/15 dark:bg-green-500/5 text-green-800 dark:text-green-400',
  removed: 'bg-red-500/15 dark:bg-red-500/5 text-red-800 dark:text-red-400',
  changed: 'bg-orange-500/15 dark:bg-orange-500/5 text-orange-800 dark:text-orange-400',
  unchanged: '',
}

function getDiffPrefix(action: string, changed: boolean) {
  if (!changed) return { char: ' ', style: DIFF_STYLES.unchanged }

  switch (action) {
    case 'create':
      return { char: '+', style: DIFF_STYLES.added }
    case 'delete':
    case 'destroy':
      return { char: '-', style: DIFF_STYLES.removed }
    case 'replace':
      return { char: '~', style: DIFF_STYLES.changed }
    case 'update':
      return { char: '~', style: DIFF_STYLES.changed }
    default:
      return { char: ' ', style: DIFF_STYLES.unchanged }
  }
}

function isKnownAfterApply(val: string) {
  return val === 'Known after apply' || val === 'Value known after apply'
}

export const TerraformValuesDiff = ({
  values,
}: {
  values: TTerraformValues
}) => {
  const valuesDiff = mapBeforeAfterToKeyValues(values)

  const formatValue = (val: any) => {
    if (val === null || typeof val === 'undefined') return 'null'
    if (val === '') return '""'
    if (typeof val === 'object') return JSON.stringify(val, null, 2)
    return String(val)
  }

  return (
    <div className="p-4 bg-code border-t shadow-xs min-h-[3rem] max-h-[40rem] overflow-auto font-mono text-[13px] leading-6">
      {valuesDiff.length ? (
        valuesDiff.map((value, idx) => {
          const prefix = getDiffPrefix(values.action, value.changed)

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
                <TreeDiffValue
                  before={value.before}
                  after={value.after}
                />
              </div>
            )
          }

          const formattedBefore = formatValue(value.before)
          const formattedAfter = formatValue(value.after)
          const isLongValue =
            formattedBefore.length > 40 || formattedAfter.length > 40

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
                      className={cn(
                        'inline-block max-w-[300px] truncate align-bottom',
                        {
                          'italic opacity-60':
                            isKnownAfterApply(formattedAfter),
                        }
                      )}
                      title={formattedAfter}
                    >
                      {formattedAfter}
                    </span>
                    {isLongValue && (
                      <DiffLineExpandButton
                        label={value.key}
                        prefix={prefix.char as '~' | '+' | '-'}
                        before={value.before}
                        after={value.after}
                      />
                    )}
                  </>
                ) : (
                  <>
                    {'  '}
                    <span
                      className={cn(
                        'inline-block max-w-[500px] truncate align-bottom',
                        {
                          'italic opacity-60':
                            isKnownAfterApply(formattedAfter),
                        }
                      )}
                      title={formattedAfter}
                    >
                      {formattedAfter}
                    </span>
                    {isLongValue && (
                      <DiffLineExpandButton
                        label={value.key}
                        prefix={prefix.char as '~' | '+' | '-'}
                        before={value.before}
                        after={value.after}
                      />
                    )}
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

type BeforeAfterObject = {
  before?: any
  after?: any
  [key: string]: any
}

type KeyValuePair = {
  key: string
  before: any
  after: any
  changed: boolean // Flag to indicate if this property actually changed
}

function mapBeforeAfterToKeyValues(obj: BeforeAfterObject): KeyValuePair[] {
  const result: KeyValuePair[] = []

  // Get all unique keys from both before and after objects
  const beforeKeys =
    obj.before && typeof obj.before === 'object' ? Object.keys(obj.before) : []
  const afterKeys =
    obj.after && typeof obj.after === 'object' ? Object.keys(obj.after) : []
  const allKeys = [...new Set([...beforeKeys, ...afterKeys])]

  // Include ALL keys, but mark which ones actually changed
  allKeys.forEach((key) => {
    const beforeValue =
      obj.before && typeof obj.before === 'object' ? obj.before[key] : undefined
    const afterValue =
      obj.after && typeof obj.after === 'object' ? obj.after[key] : undefined

    const normalizedBefore = beforeValue ?? null
    const normalizedAfter = afterValue ?? null
    const hasChanged = !deepEqual(normalizedBefore, normalizedAfter)

    result.push({
      key,
      before: normalizedBefore,
      after: normalizedAfter,
      changed: hasChanged,
    })
  })

  return result
}

