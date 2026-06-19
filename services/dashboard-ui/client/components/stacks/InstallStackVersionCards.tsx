import { Badge } from '@/components/common/Badge'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallStack, TInstallStackVersionRun } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { StackVersionActions } from './StackVersionActions'

type TStackVersion = TInstallStack['versions'][number]

const RunTypeBadge = ({ runType }: { runType?: string }) => {
  if (!runType) return null
  const theme = runType === 'workflow-run' ? 'brand' : 'info'
  const label = runType === 'workflow-run' ? 'Workflow' : 'Out of band'
  return (
    <Badge theme={theme} size="sm">
      {label}
    </Badge>
  )
}

const DiffList = ({
  label,
  items,
  theme,
}: {
  label: string
  items?: string[]
  theme: 'success' | 'error' | 'info'
}) => {
  if (!items?.length) return null
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <Text variant="subtext" theme="neutral">
        {label}:
      </Text>
      {items.map((item) => (
        <Badge key={item} theme={theme} size="sm" variant="code">
          {item}
        </Badge>
      ))}
    </div>
  )
}

const RunDiffs = ({ run }: { run: TInstallStackVersionRun }) => {
  const hasRoleDiff =
    run?.role_diff?.enabled?.length || run?.role_diff?.disabled?.length
  const hasInputDiff =
    run?.input_diff?.added?.length ||
    run?.input_diff?.removed?.length ||
    run?.input_diff?.changed?.length

  if (!hasRoleDiff && !hasInputDiff) return null

  return (
    <div className="flex flex-col gap-2">
      {hasRoleDiff ? (
        <div className="flex flex-col gap-1.5">
          <Text variant="subtext" weight="strong">
            Role changes
          </Text>
          <DiffList
            label="Enabled"
            items={run.role_diff?.enabled}
            theme="success"
          />
          <DiffList
            label="Disabled"
            items={run.role_diff?.disabled}
            theme="error"
          />
        </div>
      ) : null}
      {hasInputDiff ? (
        <div className="flex flex-col gap-1.5">
          <Text variant="subtext" weight="strong">
            Input changes
          </Text>
          <DiffList
            label="Added"
            items={run.input_diff?.added}
            theme="success"
          />
          <DiffList
            label="Changed"
            items={run.input_diff?.changed}
            theme="info"
          />
          <DiffList
            label="Removed"
            items={run.input_diff?.removed}
            theme="error"
          />
        </div>
      ) : null}
    </div>
  )
}

const StackRunCard = ({ run }: { run: TInstallStackVersionRun }) => {
  const hasDiffs =
    run?.role_diff?.enabled?.length ||
    run?.role_diff?.disabled?.length ||
    run?.input_diff?.added?.length ||
    run?.input_diff?.removed?.length ||
    run?.input_diff?.changed?.length

  return (
    <div className="border rounded-md">
      <div className="flex items-center justify-between p-3">
        <span className="flex items-center gap-2">
          <Text variant="subtext">
            <Time variant="subtext" time={run?.created_at} format="relative" />
          </Text>
          <RunTypeBadge runType={run?.run_type} />
        </span>
        <ClickToCopyButton
          className="w-fit"
          textToCopy={JSON.stringify(run?.data_contents || run?.data || {})}
        />
      </div>

      {hasDiffs ? (
        <div className="border-t p-3">
          <RunDiffs run={run} />
        </div>
      ) : null}

      {Object.keys(run?.data_contents || {}).length > 0 ? (
        <Expand
          id={`run-outputs-${run?.id}`}
          className="border-t"
          heading={
            <Text variant="subtext" theme="neutral">
              Outputs
            </Text>
          }
        >
          <div className="border-t overflow-auto max-h-[400px] p-3">
            <KeyValueList
              values={objectToKeyValueArray(run?.data_contents || {})}
            />
          </div>
        </Expand>
      ) : null}
    </div>
  )
}

const StackVersionCard = ({
  version,
  orgId,
  appId,
  isFirst,
}: {
  version: TStackVersion
  orgId: string
  appId: string
  isFirst: boolean
}) => {
  const runs = version?.runs ?? []

  return (
    <Expand
      id={`version-${version?.id}`}
      className="border rounded-lg"
      isOpen={isFirst}
      heading={
        <div className="flex items-center justify-between w-full pr-2">
          <span className="flex items-center gap-3">
            <ID>{version?.id}</ID>
            <Status variant="badge" status={version?.composite_status?.status} />
            <Text variant="subtext" theme="neutral">
              {runs.length} {runs.length === 1 ? 'run' : 'runs'}
            </Text>
          </span>
          <span className="flex items-center gap-3">
            <Text variant="subtext" theme="neutral">
              <Link href={`/${orgId}/apps/${appId}`}>
                {version?.app_config_id}
              </Link>
            </Text>
            <Time
              variant="subtext"
              time={version?.created_at}
              format="relative"
            />
            <span
              onClick={(e) => e.stopPropagation()}
              onKeyDown={(e) => e.stopPropagation()}
            >
              <StackVersionActions version={version} />
            </span>
          </span>
        </div>
      }
    >
      <div className="border-t p-4 flex flex-col gap-3">
        {runs.length > 0 ? (
          runs.map((run) => <StackRunCard key={run?.id} run={run} />)
        ) : (
          <Text variant="subtext" theme="neutral">
            No runs yet
          </Text>
        )}
      </div>
    </Expand>
  )
}

export const InstallStackVersionCards = ({
  stack,
  orgId,
  appId,
}: {
  stack: TInstallStack
  orgId: string
  appId: string
}) => {
  const versions = stack?.versions ?? []

  if (!versions.length) {
    return (
      <Text variant="subtext" theme="neutral">
        No stack versions
      </Text>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {versions.map((version, idx) => (
        <StackVersionCard
          key={version?.id}
          version={version}
          orgId={orgId}
          appId={appId}
          isFirst={idx === 0}
        />
      ))}
    </div>
  )
}
