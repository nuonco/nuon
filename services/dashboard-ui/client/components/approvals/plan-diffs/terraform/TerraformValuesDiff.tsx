import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import type { TTerraformOutputChange } from '@/types'
import { cn } from '@/utils/classnames'
import { TerraformValueModal } from './TerraformValueModal'

type TTerraformValues = Pick<
  TTerraformOutputChange,
  'before' | 'after' | 'action'
>

export const TerraformValuesDiff = ({
  values,
}: {
  values: TTerraformValues
}) => {
  const valuesDiff = mapBeforeAfterToKeyValues(values)

  // Helper for displaying arrays/objects
  const formatValue = (val: any) => {
    if (val === null || typeof val === 'undefined') return 'null'
    if (val === '') return '""' // Add this line for empty strings
    if (typeof val === 'object') return JSON.stringify(val, null, 2)
    return String(val)
  }

  // Get diff symbol
  const getDiffSymbol = (action: string, hasChanged: boolean) => {
    const blank = <Text theme="neutral">&nbsp;&nbsp;</Text>
    if (!hasChanged) return blank

    switch (action) {
      case 'replace':
        return <Text theme="brand">-/+</Text>
      case 'create':
        return <Text theme="success">+</Text>
      case 'destroy':
        return <Text theme="error">-</Text>
      case 'update':
        return <Text theme="warn">~</Text>
      default:
        return null
    }
  }

  return (
    <div className="p-4 bg-code border-t shadow-xs min-h-[3rem] max-h-[40rem] overflow-auto">
      {valuesDiff.length ? (
        valuesDiff.map((value, idx) => {
          const formattedBefore = formatValue(value.before)
          const formattedAfter = formatValue(value.after)

          return (
            <div
              className="flex items-center flex-nowrap w-max"
              key={value.key + idx}
            >
              <Text family="mono" weight="strong">
                {getDiffSymbol(values.action, value?.changed)} {value.key}
              </Text>
              :{'  '}
              {value?.changed ? (
                <>
                  <Badge
                    className="ml-2 line-through !text-nowrap !border-none !text-sm"
                    variant="code"
                    size="sm"
                    theme="error"
                  >
                    {formattedBefore.length >= 50 ? (
                      <span className="flex items-center gap-2">
                        <span className="max-w-[200px] !inline-block truncate">
                          {formattedBefore}{' '}
                        </span>
                        <TerraformValueModal
                          isBefore
                          valueKey={value.key}
                          value={formattedBefore}
                        />
                      </span>
                    ) : (
                      formattedBefore
                    )}
                  </Badge>
                  <Text
                    className="!text-nowrap mx-2"
                    family="mono"
                    theme="neutral"
                  >{`->`}</Text>
                  <Badge
                    className={cn('!text-nowrap !border-none !text-sm', {
                      'italic opacity-60 bg-black/5 dark:bg-white/5':
                        formattedAfter === 'Known after apply' ||
                        formattedAfter === 'Value known after apply',
                    })}
                    variant="code"
                    size="sm"
                    theme={
                      formatValue(value.after) === 'Known after apply' ||
                      formatValue(value.after) === 'Value known after apply'
                        ? 'neutral'
                        : values.action === 'create'
                          ? 'success'
                          : values.action === 'replace'
                            ? 'brand'
                            : values.action === 'delete'
                              ? 'error'
                              : 'warn'
                    }
                  >
                    {formattedAfter.length >= 50 ? (
                      <span className="flex items-center gap-2">
                        <span className="max-w-[200px] !inline-block truncate">
                          {formattedAfter}{' '}
                        </span>
                        <TerraformValueModal
                          valueKey={value.key}
                          value={formattedAfter}
                        />
                      </span>
                    ) : (
                      formattedAfter
                    )}
                  </Badge>
                </>
              ) : (
                <Badge
                  className="ml-2 !text-nowrap !border-none !text-sm"
                  theme="neutral"
                  variant="code"
                  size="sm"
                >
                  {formattedAfter.length >= 50 ? (
                    <span className="flex items-center gap-2">
                      <span className="max-w-[400px] !inline-block truncate">
                        {formattedAfter}{' '}
                      </span>
                      <TerraformValueModal
                        valueKey={value.key}
                        value={formattedAfter}
                      />
                    </span>
                  ) : (
                    formattedAfter
                  )}
                </Badge>
              )}
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

    // Check if values actually changed
    const hasChanged = !deepEqual(beforeValue, afterValue)

    result.push({
      key,
      before: beforeValue ?? null,
      after: afterValue ?? null,
      changed: hasChanged,
    })
  })

  return result
}

// Deep equality check function
function deepEqual(a: any, b: any): boolean {
  if (a === b) return true

  if (a == null || b == null) return a === b

  if (typeof a !== typeof b) return false

  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false
    return a.every((item, index) => deepEqual(item, b[index]))
  }

  if (typeof a === 'object') {
    const keysA = Object.keys(a)
    const keysB = Object.keys(b)

    if (keysA.length !== keysB.length) return false

    return keysA.every(
      (key) => keysB.includes(key) && deepEqual(a[key], b[key])
    )
  }

  return false
}
