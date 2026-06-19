import { Badge } from '@/components/common/Badge'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallStack, TInstallStackVersionRun } from '@/types'
import { cn } from '@/utils/classnames'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { indexToOrdinal } from '@/utils/string-utils'

type TStackVersion = TInstallStack['versions'][number]

export const StackVersionDetails = ({
  version,
  ...props
}: {
  version: TStackVersion
} & IPanel) => {
  return (
    <Panel
      size="3/4"
      {...props}
      className={cn('border-t-6', {
        '!border-t-orange-600 dark:!border-t-orange-500':
          version?.composite_status?.status === 'expired',
        '!border-t-blue-600 dark:!border-t-blue-500':
          version?.composite_status?.status === 'awaiting-user-run',
        '!border-t-green-600 dark:!border-t-green-500':
          version?.composite_status?.status === 'active',
      })}
      headerClassName="!h-fit md:px-6 !items-start !py-4"
      heading={
        <div className="flex flex-col">
          <Status status={version?.composite_status?.status} />
          <Text
            flex
            className="gap-2"
            variant="h3"
            weight="stronger"
          >
            Stack version details
          </Text>
          <Text
            flex
            className="gap-8"
            theme="neutral"
            variant="subtext"
          >
            <span>
              <span>Created:</span>{' '}
              <Time variant="subtext" time={version?.created_at} />
            </span>
            <span>
              <span>Updated:</span>{' '}
              <Time variant="subtext" time={version?.updated_at} />
            </span>
          </Text>
        </div>
      }
    >
      <div className="flex flex-col gap-12 my-8">
        <StackVersionLinks version={version} />

        <StackVersionRuns version={version} />
      </div>

      <Divider dividerWord="Metadata" />

      <StackVersionMetadata version={version} />
    </Panel>
  )
}

const StackVersionLinks = ({ version }: { version: TStackVersion }) => {
  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Links
      </Text>

      {version?.quick_link_url ? (
        <div className="border rounded-md shadow p-2 flex flex-col gap-1">
          <span className="flex justify-between items-center">
            <Text variant="body" weight="strong">
              Install quick link
            </Text>
            <ClickToCopyButton textToCopy={version.quick_link_url} />
          </span>
          <Link href={version.quick_link_url} isExternal>
            <Code>{version.quick_link_url}</Code>
          </Link>
        </div>
      ) : null}

      {version?.template_url ? (
        <div className="border rounded-md shadow p-2 flex flex-col gap-1 mt-3">
          <span className="flex justify-between items-center">
            <Text variant="body" weight="strong">
              Install template
            </Text>
            <ClickToCopyButton textToCopy={version.template_url} />
          </span>
          <Link href={version.template_url} isExternal>
            <Code>{version.template_url}</Code>
          </Link>
        </div>
      ) : null}
    </div>
  )
}

const RunTypeBadge = ({ runType }: { runType?: string }) => {
  if (!runType) return null
  const theme = runType === 'workflow-run' ? 'brand' : 'info'
  const label = runType === 'workflow-run' ? 'Workflow' : 'Out of band'
  return <Badge theme={theme} size="sm">{label}</Badge>
}

const DiffList = ({ label, items, theme }: { label: string; items?: string[]; theme: 'success' | 'error' | 'info' }) => {
  if (!items?.length) return null
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <Text variant="subtext" theme="neutral">{label}:</Text>
      {items.map((item) => (
        <Badge key={item} theme={theme} size="sm" variant="code">{item}</Badge>
      ))}
    </div>
  )
}

const RunDiffs = ({ run }: { run: TInstallStackVersionRun }) => {
  const hasRoleDiff = run?.role_diff?.enabled?.length || run?.role_diff?.disabled?.length
  const hasInputDiff = run?.input_diff?.added?.length || run?.input_diff?.removed?.length || run?.input_diff?.changed?.length

  if (!hasRoleDiff && !hasInputDiff) return null

  return (
    <div className="flex flex-col gap-2">
      {hasRoleDiff ? (
        <div className="flex flex-col gap-1.5">
          <Text variant="subtext" weight="strong">Role changes</Text>
          <DiffList label="Enabled" items={run.role_diff?.enabled} theme="success" />
          <DiffList label="Disabled" items={run.role_diff?.disabled} theme="error" />
        </div>
      ) : null}
      {hasInputDiff ? (
        <div className="flex flex-col gap-1.5">
          <Text variant="subtext" weight="strong">Input changes</Text>
          <DiffList label="Added" items={run.input_diff?.added} theme="success" />
          <DiffList label="Changed" items={run.input_diff?.changed} theme="info" />
          <DiffList label="Removed" items={run.input_diff?.removed} theme="error" />
        </div>
      ) : null}
    </div>
  )
}

const StackVersionRuns = ({ version }: { version: TStackVersion }) => {
  const runs = version?.runs ?? []
  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Runs
      </Text>

      {runs.length ? (
        runs.map((run, idx) => {
          const ordinalIdx = (version?.runs?.length ?? 0) - 1 - idx
          return (
            <Expand
              key={run?.id}
              id={`run-${run?.id}`}
              className="border rounded-md"
              isOpen={idx === 0}
              heading={
                <span className="flex items-center gap-2">
                  <Text variant="base">
                    {indexToOrdinal(ordinalIdx)} run &middot;{' '}
                    <Time variant="subtext" time={run?.created_at} />
                  </Text>
                  <RunTypeBadge runType={run?.run_type} />
                </span>
              }
            >
              <div className="border-t p-4 flex flex-col gap-4">
                <RunDiffs run={run} />
                <div className="flex flex-col gap-2">
                  <div className="flex justify-between items-center">
                    <Text variant="subtext" weight="strong">Outputs</Text>
                    <ClickToCopyButton
                      className="w-fit"
                      textToCopy={JSON.stringify(
                        run?.data_contents || run?.data || {}
                      )}
                    />
                  </div>
                  <div className="overflow-auto max-h-[600px]">
                    <KeyValueList
                      values={objectToKeyValueArray(run?.data_contents || {})}
                    />
                  </div>
                </div>
              </div>
            </Expand>
          )
        })
      ) : (
        <Text variant="subtext">No runs for this stack version.</Text>
      )}
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
      <Status status={status.status} variant="badge" />
      <Time seconds={status.created_at_ts} variant="subtext" theme="neutral" />
    </span>
  )
}

export const StackVersionMetadata = ({
  version,
}: {
  version: TStackVersion
}) => {
  return (
    <div className="flex flex-col gap-2">
      <Expand
        className="border rounded-md"
        id="status-history"
        heading={
          <Text family="mono" variant="subtext">
            View version history
          </Text>
        }
      >
        <div className="border-t flex flex-col p-4 divide-y">
          {version?.composite_status?.history?.map((status, idx) => (
            <StackHistoryStatus
              key={`${status.created_at_ts}-${idx}`}
              status={status}
            />
          ))}
          <StackHistoryStatus status={version?.composite_status} />
        </div>
      </Expand>

      <Expand
        className="border rounded-md"
        id="version-json"
        heading={
          <Text family="mono" variant="subtext">
            View version contents JSON
          </Text>
        }
      >
        <div className="border-t">
          {version?.contents ? (
            <JSONViewer
              className="!border-none !rounded-none"
              data={atob(version?.contents)}
            />
          ) : (
            <div className="px-4 py-6">
              <Text theme="neutral">No version contents to show</Text>
            </div>
          )}
        </div>
      </Expand>

      <Expand
        className="border rounded-md"
        id="version-config"
        heading={
          <Text family="mono" variant="subtext">
            View AWS details
          </Text>
        }
      >
        <div className="border-t flex flex-col p-4 divide-y">
          <KeyValueList
            values={objectToKeyValueArray({
              app_config_id: version?.app_config_id,
              aws_bucket_key: version?.aws_bucket_key,
              aws_bucket_name: version?.aws_bucket_name,
              phone_home_id: version?.phone_home_id,
              phone_home_url: version?.phone_home_url,
            })}
          />
        </div>
      </Expand>
    </div>
  )
}
