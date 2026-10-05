import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Divider } from '@/components/common/Divider'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TInstallDeploymentRecord } from '@/types'

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

const ChangeRows = ({
  changes,
}: {
  changes: TInstallDeploymentRecord['change_groups'][number]['changes']
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
          {change.is_redacted ? (
            <Badge size="sm" theme="neutral">
              Redacted
            </Badge>
          ) : (
            <>
              {change.previous_value !== undefined && (
                <Text
                  variant="subtext"
                  family="mono"
                  theme="neutral"
                  className="truncate"
                >
                  {change.previous_value}
                </Text>
              )}
              {change.previous_value !== undefined &&
                change.next_value !== undefined && (
                  <Icon
                    variant="ArrowRightIcon"
                    size={12}
                    className="shrink-0 text-cool-grey-400"
                  />
                )}
              {change.next_value !== undefined && (
                <Text variant="subtext" family="mono" className="truncate">
                  {change.next_value}
                </Text>
              )}
            </>
          )}
        </div>
      </div>
    ))}
  </div>
)

export const DeploymentConfigChanges = ({
  deployment,
}: {
  deployment?: TInstallDeploymentRecord
}) => {
  const groups = (deployment?.change_groups ?? []).filter(
    (group) => group.changes.length || group.file_diff
  )
  if (!groups.length)
    return (
      <EmptyState
        emptyTitle="No template updates"
        emptyMessage="This deployment has no app config changes."
      />
    )
  return (
    <div className="flex flex-col gap-4">
      {groups.map((group) => (
        <div key={group.id} className="flex flex-col gap-3">
          <Divider dividerWord={group.label} />
          {group.summary && (
            <Text variant="subtext" theme="neutral">
              {group.summary}
            </Text>
          )}
          {group.changes.length > 0 && <ChangeRows changes={group.changes} />}
          {group.file_diff && (
            <CodeBlock language={group.diff_language ?? 'diff'} showCopy>
              {group.file_diff}
            </CodeBlock>
          )}
        </div>
      ))}
    </div>
  )
}
