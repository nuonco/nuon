import { Badge } from '@/components/common/Badge'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallStack, TInstallStackVersionRun } from '@/types'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { indexToOrdinal } from '@/utils/string-utils'

type TStackVersion = TInstallStack['versions'][number]

export interface IStackVersionDetails extends IPanel {
  version: TStackVersion
}

export const StackVersionDetails = ({
  version,
  ...props
}: IStackVersionDetails) => {
  const runs = version?.runs ?? []

  return (
    <Panel heading="Stack version" size="half" {...props}>
      <div className="flex items-center gap-2">
        <Status variant="badge" status={version?.composite_status?.status} />
        <ID>{version?.id}</ID>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Created">
          {version?.created_at ? (
            <Time variant="subtext" time={version.created_at} format="relative" />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="Updated">
          {version?.updated_at ? (
            <Time variant="subtext" time={version.updated_at} format="relative" />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="App config">
          {version?.app_config_id ? (
            <ID>{version.app_config_id}</ID>
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="Runs">
          <Text variant="subtext" family="mono">
            {runs.length}
          </Text>
        </LabeledValue>
        {version?.quick_link_url ? (
          <LabeledValue className="col-span-2" label="Install quick link">
            <span className="flex items-center justify-between gap-2 min-w-0">
              <Link href={version.quick_link_url} isExternal>
                <Code>{version.quick_link_url}</Code>
              </Link>
              <ClickToCopyButton textToCopy={version.quick_link_url} />
            </span>
          </LabeledValue>
        ) : null}
        {version?.template_url ? (
          <LabeledValue className="col-span-2" label="Install template">
            <span className="flex items-center justify-between gap-2 min-w-0">
              <Link href={version.template_url} isExternal>
                <Code>{version.template_url}</Code>
              </Link>
              <ClickToCopyButton textToCopy={version.template_url} />
            </span>
          </LabeledValue>
        ) : null}
      </div>

      <Divider dividerWord="Runs" />

      <StackVersionRuns version={version} />

      <Divider dividerWord="Status history" />

      <StackVersionStatusHistory version={version} />

      <Divider dividerWord="Metadata" />

      <StackVersionMetadata version={version} />
    </Panel>
  )
}

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

const StackVersionRuns = ({ version }: { version: TStackVersion }) => {
  const runs = version?.runs ?? []

  if (!runs.length) {
    return (
      <Text variant="subtext" theme="neutral">
        No runs for this stack version.
      </Text>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {runs.map((run, idx) => {
        const ordinalIdx = runs.length - 1 - idx
        return (
          <div key={run?.id} className="flex flex-col gap-3">
            <div className="flex items-center justify-between gap-2">
              <span className="flex items-center gap-2">
                <Text variant="subtext" weight="strong">
                  {indexToOrdinal(ordinalIdx)} run
                </Text>
                <Time variant="subtext" theme="neutral" time={run?.created_at} />
                <RunTypeBadge runType={run?.run_type} />
              </span>
              <ClickToCopyButton
                textToCopy={JSON.stringify(run?.data_contents || run?.data || {})}
              />
            </div>

            <RunDiffs run={run} />

            {Object.keys(run?.data_contents || {}).length > 0 ? (
              <div className="overflow-auto max-h-[400px]">
                <KeyValueList
                  values={objectToKeyValueArray(run?.data_contents || {})}
                />
              </div>
            ) : (
              <Text variant="subtext" theme="neutral">
                No outputs reported for this run.
              </Text>
            )}
          </div>
        )
      })}
    </div>
  )
}

const StackVersionStatusHistory = ({
  version,
}: {
  version: TStackVersion
}) => {
  const history = version?.composite_status?.history ?? []

  return (
    <div className="flex flex-col divide-y">
      {history.map((status, idx) => (
        <StackHistoryStatus
          key={`${status.created_at_ts}-${idx}`}
          status={status}
        />
      ))}
      {version?.composite_status ? (
        <StackHistoryStatus status={version.composite_status} />
      ) : null}
    </div>
  )
}

const StackHistoryStatus = ({
  status,
}: {
  status: TStackVersion['composite_status']['history'][number]
}) => {
  return (
    <span className="flex items-center gap-4 py-2">
      <Status status={status?.status} variant="badge" />
      <Time seconds={status?.created_at_ts} variant="subtext" theme="neutral" />
    </span>
  )
}

export const StackVersionMetadata = ({
  version,
}: {
  version: TStackVersion
}) => {
  return (
    <div className="flex flex-col gap-4">
      <KeyValueList
        values={objectToKeyValueArray({
          app_config_id: version?.app_config_id,
          aws_bucket_key: version?.aws_bucket_key,
          aws_bucket_name: version?.aws_bucket_name,
          phone_home_id: version?.phone_home_id,
          phone_home_url: version?.phone_home_url,
        })}
      />

      {version?.contents ? (
        <JSONViewer data={atob(version.contents)} showCopy />
      ) : (
        <Text variant="subtext" theme="neutral">
          No version contents to show.
        </Text>
      )}
    </div>
  )
}
