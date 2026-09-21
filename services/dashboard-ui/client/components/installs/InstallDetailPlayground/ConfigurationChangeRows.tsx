import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TConfigurationChange } from './types'

const CHANGE_THEME = {
  add: 'success',
  remove: 'error',
  change: 'warn',
} as const

const CHANGE_PREFIX = {
  add: '+',
  remove: '-',
  change: '~',
} as const

export const ConfigurationChangeRows = ({
  changes,
}: {
  changes: TConfigurationChange[]
}) => (
  <div className="flex flex-col divide-y border rounded-md overflow-hidden">
    {changes.map((change, index) => (
      <div
        key={`${change.path}-${change.operation}-${index}`}
        className="grid grid-cols-[1rem_minmax(0,1fr)] md:grid-cols-[1rem_minmax(10rem,1fr)_minmax(0,2fr)] items-center gap-3 px-3 py-2"
      >
        <Text
          variant="subtext"
          family="mono"
          weight="strong"
          theme={CHANGE_THEME[change.operation]}
        >
          {CHANGE_PREFIX[change.operation]}
        </Text>
        <Text variant="subtext" family="mono" weight="strong">
          {change.path}
        </Text>
        <div className="col-start-2 md:col-start-auto flex items-center gap-2 min-w-0">
          {change.isRedacted ? (
            <Badge size="sm" theme="neutral">
              Redacted
            </Badge>
          ) : (
            <>
              {change.previousValue !== undefined && (
                <Text
                  variant="subtext"
                  family="mono"
                  theme="neutral"
                  className="truncate"
                >
                  {change.previousValue}
                </Text>
              )}
              {change.previousValue !== undefined &&
                change.nextValue !== undefined && (
                  <Icon
                    variant="ArrowRightIcon"
                    size={12}
                    className="shrink-0 text-cool-grey-400"
                  />
                )}
              {change.nextValue !== undefined && (
                <Text variant="subtext" family="mono" className="truncate">
                  {change.nextValue}
                </Text>
              )}
            </>
          )}
        </div>
      </div>
    ))}
  </div>
)
