import { Badge } from '@/components/common/Badge'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TConfigDiffNode, TInstallConfigVersion } from '@/types'
import {
  collectChangedLeaves,
  type IConfigDiffChange,
} from '../config-diff-utils'

const OP_THEME = {
  add: 'success',
  remove: 'error',
  change: 'warn',
  noop: 'neutral',
  '': 'neutral',
} as const

const OP_PREFIX = {
  add: '+',
  remove: '-',
  change: '~',
  noop: ' ',
  '': ' ',
} as const

export interface IInstallConfigVersionDetails extends IPanel {
  version: TInstallConfigVersion
  diff?: TConfigDiffNode
  isDiffLoading?: boolean
}

export const InstallConfigVersionDetails = ({
  version,
  diff,
  isDiffLoading,
  ...props
}: IInstallConfigVersionDetails) => {
  const sync = version?.install_config_sync
  const commit = sync?.vcs_connection_commit

  return (
    <Panel heading="Install config version" size="half" {...props}>
      <div className="flex items-center gap-2">
        <Status variant="badge" status={version?.status?.status ?? 'unknown'} />
        <Badge size="sm" theme={version?.created ? 'success' : 'info'}>
          {version?.created ? 'Created' : 'Updated'}
        </Badge>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Synced">
          {version?.created_at ? (
            <Time
              variant="subtext"
              time={version.created_at}
              format="relative"
            />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        <LabeledValue label="Triggered by">
          {sync?.triggered_by ?? 'unknown'}
        </LabeledValue>
        <LabeledValue label="Version">
          <ID>{version?.id}</ID>
        </LabeledValue>
        <LabeledValue label="File">
          {version?.file_path ? (
            <Text variant="subtext" family="mono">
              {version.file_path}
            </Text>
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
      </div>

      {commit ? (
        <LabeledValue label="Commit">
          <span className="flex flex-col gap-1">
            {commit.message ? (
              <Text variant="body" weight="strong">
                {commit.message.split('\n')[0]?.trim()}
              </Text>
            ) : null}
            <span className="flex flex-wrap items-center gap-2">
              {commit.sha ? (
                <Badge size="sm" variant="code">
                  {commit.sha.slice(0, 8)}
                </Badge>
              ) : null}
              {commit.author_name ? (
                <Text variant="subtext" theme="neutral">
                  {commit.author_name}
                </Text>
              ) : null}
            </span>
          </span>
        </LabeledValue>
      ) : null}

      <Divider dividerWord="Changes" />

      {isDiffLoading && !diff ? (
        <Loading variant="large" />
      ) : diff ? (
        <ConfigDiffList node={diff} />
      ) : (
        <Text variant="subtext" theme="neutral">
          No diff available for this version.
        </Text>
      )}
    </Panel>
  )
}

const ConfigDiffList = ({ node }: { node: TConfigDiffNode }) => {
  const changes = collectChangedLeaves(node)

  if (changes.length === 0) {
    return (
      <Text variant="subtext" theme="neutral">
        No changes detected
      </Text>
    )
  }

  return (
    <div className="flex flex-col border rounded-md divide-y overflow-hidden">
      {changes.map((change) => (
        <ConfigDiffRow key={change.path} change={change} />
      ))}
    </div>
  )
}

const ConfigDiffRow = ({ change }: { change: IConfigDiffChange }) => {
  const op = change.op || ''

  return (
    <div className="flex items-center gap-3 px-4 py-2.5">
      <Text
        variant="subtext"
        theme={OP_THEME[op]}
        weight="strong"
        family="mono"
        className="shrink-0 w-4 text-center"
      >
        {OP_PREFIX[op]}
      </Text>
      <Text variant="subtext" weight="strong" family="mono">
        {change.path}
      </Text>
      <Text
        variant="subtext"
        theme="neutral"
        family="mono"
        className="ml-auto truncate max-w-[50%]"
      >
        {change.diff}
      </Text>
    </div>
  )
}
