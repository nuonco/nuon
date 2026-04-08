import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel } from '@/components/surfaces/Panel'
import type { TInstallStack } from '@/types'
import { cn } from '@/utils/classnames'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { indexToOrdinal } from '@/utils/string-utils'

type TStackVersion = TInstallStack['versions'][number]

export const StackVersionDetails = ({
  version,
}: {
  version: TStackVersion
}) => {
  return (
    <Panel
      size="3/4"
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
      triggerButton={{
        children: (
          <>
            <Icon variant="Info" />
          </>
        ),
        className: '!p-1',
        size: 'sm',
        variant: 'ghost',
      }}
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

const StackVersionRuns = ({ version }: { version: TStackVersion }) => {
  const reversedRuns = version?.runs ? [...version.runs].reverse() : []
  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Runs
      </Text>

      {reversedRuns.length ? (
        reversedRuns.map((run, displayIdx) => {
          const originalIdx = (version?.runs?.length ?? 0) - 1 - displayIdx
          return (
            <Expand
              key={run?.id}
              id={`run-${run?.id}`}
              className="border rounded-md"
              isOpen={displayIdx === 0}
              heading={
                <Text variant="base">
                  {indexToOrdinal(originalIdx)} run &middot;{' '}
                  <Time variant="subtext" time={run?.created_at} />
                </Text>
              }
            >
              <div className="border-t p-4 flex flex-col gap-2">
                <ClickToCopyButton
                  className="w-fit self-end"
                  textToCopy={JSON.stringify(
                    run?.data_contents || run?.data || {}
                  )}
                />
                <div className="overflow-auto max-h-[600px]">
                  <KeyValueList
                    values={objectToKeyValueArray(run?.data_contents || {})}
                  />
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
