import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type {
  TCustomerManagedBundleCandidate,
  TCustomerManagedBundleChange,
} from '@/lib/ctl-api/installs/customer-managed-support-snapshots'
import { diffLines } from '@/utils/code-utils'
import { formatBytes, snakeToWords, toSentenceCase } from '@/utils/string-utils'

const definition = (
  change: TCustomerManagedBundleChange,
  side: 'previous' | 'candidate'
) => {
  if (change.kind === 'component') {
    return change[`${side}_component_definition`]
  }
  if (change.kind === 'action') return change[`${side}_action_definition`]
  if (change.kind === 'runbook') return change[`${side}_runbook_definition`]
  return undefined
}

const tomlValue = (value: unknown) => {
  if (typeof value === 'string') return JSON.stringify(value)
  if (value === null) return 'null'
  return JSON.stringify(value)
}

const flattenDefinition = (value: unknown, path: string[] = []): string[] => {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .flatMap(([key, child]) => flattenDefinition(child, [...path, key]))
  }
  return [`${path.join('.')} = ${tomlValue(value)}`]
}

const displayName = (change: TCustomerManagedBundleChange) =>
  change.kind === 'stack-asset' && change.name === 'root'
    ? 'Install stack'
    : change.name

const changeTheme = {
  added: 'success',
  changed: 'warn',
  removed: 'error',
  unchanged: 'neutral',
} as const

const DigestChange = ({
  label,
  previous,
  candidate,
}: {
  label: string
  previous?: string
  candidate?: string
}) => {
  if (!previous && !candidate) return null
  return (
    <span className="flex flex-wrap items-center gap-2">
      <Text variant="subtext" theme="neutral">
        {label}
      </Text>
      {previous ? <Hash hash={previous} length={10} /> : <Text>None</Text>}
      <Icon variant="ArrowRightIcon" size={14} />
      {candidate ? <Hash hash={candidate} length={10} /> : <Text>Removed</Text>}
    </span>
  )
}

const StagedChange = ({ change }: { change: TCustomerManagedBundleChange }) => {
  const previousDefinition = definition(change, 'previous')
  const candidateDefinition = definition(change, 'candidate')
  const hasDefinition = !!previousDefinition || !!candidateDefinition

  return (
    <Expand
      id={`staged-bundle-${change.kind}-${change.name}`.replace(
        /[^a-zA-Z0-9-_]/g,
        '-'
      )}
      isOpen
      className="border rounded-md"
      headerClassName="px-4 py-3"
      heading={
        <div className="flex flex-1 items-center justify-between gap-4 text-left">
          <span className="flex flex-col min-w-0">
            <Text weight="strong">{displayName(change)}</Text>
            <Text variant="subtext" theme="neutral">
              {toSentenceCase(snakeToWords(change.kind))}
              {change.detail ? ` · ${change.detail}` : ''}
            </Text>
          </span>
          <Badge theme={changeTheme[change.change]}>{change.change}</Badge>
        </div>
      }
    >
      <div className="flex flex-col gap-4 border-t p-4">
        <DigestChange
          label="Content"
          previous={change.previous_digest}
          candidate={change.candidate_digest}
        />
        <DigestChange
          label="Configuration"
          previous={change.previous_config_digest}
          candidate={change.candidate_config_digest}
        />
        {hasDefinition ? (
          <div className="flex flex-col gap-2">
            <Text variant="label" theme="neutral">
              Bundled definition
            </Text>
            <CodeBlock language="toml" isDiff>
              {diffLines(
                previousDefinition
                  ? flattenDefinition(previousDefinition).join('\n')
                  : undefined,
                candidateDefinition
                  ? flattenDefinition(candidateDefinition).join('\n')
                  : undefined
              )}
            </CodeBlock>
          </div>
        ) : null}
      </div>
    </Expand>
  )
}

export const SnapshotStagedBundleDiff = ({
  candidate,
}: {
  candidate: TCustomerManagedBundleCandidate
}) => {
  const changes = candidate.changes.filter(
    ({ change }) => change !== 'unchanged'
  )
  const counts = changes.reduce(
    (result, change) => ({
      ...result,
      [change.change]: result[change.change] + 1,
    }),
    { added: 0, changed: 0, removed: 0 }
  )

  return (
    <Card className="border-warn-500/40">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <HeadingGroup>
          <span className="flex items-center gap-2">
            <Text variant="base" weight="strong">
              Staged bundle
            </Text>
            <Badge theme="warn">Not active</Badge>
          </span>
          <Text variant="subtext" theme="neutral">
            Uploaded and processed in the customer environment, but not
            deployed.
          </Text>
          <Time time={candidate.staged_at} format="long-datetime" />
        </HeadingGroup>
        {candidate.archive_size ? (
          <Text variant="subtext" theme="neutral">
            {formatBytes(candidate.archive_size)}
          </Text>
        ) : null}
      </div>
      <div className="flex flex-wrap items-center gap-3 border-y py-4">
        <Hash hash={candidate.previous_digest} length={16} />
        <Icon variant="ArrowRightIcon" />
        <Hash hash={candidate.bundle.bundle_digest} length={16} />
        <span className="flex flex-wrap gap-2 ml-auto">
          {counts.added ? (
            <Badge theme="success">{counts.added} added</Badge>
          ) : null}
          {counts.changed ? (
            <Badge theme="warn">{counts.changed} changed</Badge>
          ) : null}
          {counts.removed ? (
            <Badge theme="error">{counts.removed} removed</Badge>
          ) : null}
        </span>
      </div>
      {changes.length ? (
        <div className="flex flex-col gap-3">
          {changes.map((change) => (
            <StagedChange
              change={change}
              key={`${change.kind}-${change.name}`}
            />
          ))}
        </div>
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No bundle changes"
          emptyMessage="The staged bundle matches the active bundle."
        />
      )}
    </Card>
  )
}
