import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { AppConfigDiffComponent } from '@/components/branches/AppConfigDiff'
import type { DiffSectionData } from '@/components/branches/AppConfigDiff'
import type {
  TCustomerManagedBundle,
  TCustomerManagedBundleArtifact,
} from '@/types'
import { formatBytes, snakeToWords, toSentenceCase } from '@/utils/string-utils'

type TBundleArtifactChange = {
  key: string
  kind: string
  name: string
  change: 'added' | 'changed' | 'removed' | 'unchanged'
  previous?: TCustomerManagedBundleArtifact
  current?: TCustomerManagedBundleArtifact
}

const artifactKey = (artifact: TCustomerManagedBundleArtifact) =>
  `${artifact.kind ?? 'unknown'}:${artifact.logical_name ?? artifact.id ?? 'unknown'}`

const artifactsMatch = (
  previous: TCustomerManagedBundleArtifact,
  current: TCustomerManagedBundleArtifact
) =>
  previous.digest === current.digest &&
  previous.config_digest === current.config_digest &&
  previous.repository === current.repository &&
  previous.media_type === current.media_type &&
  previous.size === current.size &&
  previous.source_type === current.source_type &&
  previous.platform_os === current.platform_os &&
  previous.platform_architecture === current.platform_architecture

export const compareBundleArtifacts = (
  previousArtifacts: TCustomerManagedBundleArtifact[],
  currentArtifacts: TCustomerManagedBundleArtifact[]
): TBundleArtifactChange[] => {
  const previousByKey = new Map(
    previousArtifacts.map((artifact) => [artifactKey(artifact), artifact])
  )
  const currentByKey = new Map(
    currentArtifacts.map((artifact) => [artifactKey(artifact), artifact])
  )
  const keys = new Set([...previousByKey.keys(), ...currentByKey.keys()])

  return [...keys]
    .map((key): TBundleArtifactChange => {
      const previous = previousByKey.get(key)
      const current = currentByKey.get(key)
      return {
        key,
        kind: current?.kind ?? previous?.kind ?? 'unknown',
        name:
          current?.logical_name ?? previous?.logical_name ?? 'Unnamed artifact',
        change: !previous
          ? 'added'
          : !current
            ? 'removed'
            : artifactsMatch(previous, current)
              ? 'unchanged'
              : 'changed',
        previous,
        current,
      }
    })
    .sort((left, right) => left.name.localeCompare(right.name))
}

const changeGroups = [
  {
    label: 'Components',
    description: 'Application workloads and images',
    kinds: ['component', 'image'],
  },
  {
    label: 'Infrastructure',
    description: 'Install stack, sandbox, and supporting assets',
    kinds: ['sandbox', 'stack_asset'],
  },
  {
    label: 'Operations',
    description: 'Actions and runbooks',
    kinds: ['action_step', 'action', 'runbook'],
  },
  {
    label: 'Packaging',
    description: 'Runner, portal, and supporting artifacts',
    kinds: ['runner_binary', 'runner_image', 'portal_binary', 'unknown'],
  },
]

const changeTheme = {
  added: 'success',
  changed: 'info',
  removed: 'error',
  unchanged: 'neutral',
} as const

const shortDigest = (digest?: string) => digest?.replace(/^sha256:/, '')

const isInstallStack = (change: TBundleArtifactChange) => {
  const artifact = change.current ?? change.previous
  return (
    change.kind === 'stack_asset' &&
    (change.name === 'root' || artifact?.repository?.startsWith('compiled:'))
  )
}

const artifactDisplay = (change: TBundleArtifactChange) => {
  if (isInstallStack(change)) {
    return {
      name: 'Install stack',
      kind: 'CloudFormation stack',
      icon: 'StackIcon' as const,
    }
  }
  if (change.kind === 'sandbox') {
    return {
      name: 'Sandbox',
      kind: 'Sandbox',
      icon: 'TerminalWindowIcon' as const,
    }
  }
  if (change.kind === 'component' || change.kind === 'image') {
    return {
      name: change.name,
      kind: 'Component',
      icon: 'CubeIcon' as const,
    }
  }
  return {
    name: change.name,
    kind: toSentenceCase(snakeToWords(change.kind)),
    icon: 'ArchiveIcon' as const,
  }
}

const DigestChange = ({
  label,
  previous,
  current,
}: {
  label: string
  previous?: string
  current?: string
}) => {
  if (!previous && !current) return null
  return (
    <span className="flex flex-wrap items-center gap-2">
      <Text variant="subtext" theme="neutral">
        {label}
      </Text>
      {previous && previous !== current ? (
        <Hash hash={shortDigest(previous)} length={10} />
      ) : null}
      {previous && previous !== current ? (
        <Icon variant="ArrowRightIcon" size={14} />
      ) : null}
      <Hash hash={shortDigest(current ?? previous)} length={10} />
    </span>
  )
}

const artifactHref = (
  artifact: TCustomerManagedBundleArtifact | undefined,
  orgId: string,
  appId: string
) => {
  const appPath = `/${orgId}/apps/${appId}`
  switch (artifact?.kind) {
    case 'component':
    case 'image':
      return artifact.component_id
        ? `${appPath}/components/${artifact.component_id}`
        : undefined
    case 'sandbox':
      return `${appPath}/sandbox`
    case 'action_step':
    case 'action':
      return artifact.action_workflow_id
        ? `${appPath}/actions/${artifact.action_workflow_id}`
        : undefined
    default:
      return undefined
  }
}

const configSectionForArtifact = (
  change: TBundleArtifactChange,
  sections: DiffSectionData[]
): DiffSectionData | undefined => {
  const sectionKey =
    change.kind === 'component' || change.kind === 'image'
      ? 'components'
      : change.kind === 'sandbox'
        ? 'sandbox'
        : isInstallStack(change)
          ? 'stack'
          : change.kind === 'action' || change.kind === 'action_step'
            ? 'actions'
            : change.kind === 'runbook'
              ? 'runbooks'
              : undefined
  if (!sectionKey) return undefined

  const section = sections.find(({ sectionKey: key }) => key === sectionKey)
  if (!section || !section.grouped) return section

  const entityName = change.name.split('/')[0]
  const entities = section.entities.filter(
    ({ name }) =>
      name === change.name ||
      name === entityName ||
      name.replace(/^(component|action|runbook)\./, '') === entityName
  )
  if (!entities.length) return undefined

  return {
    ...section,
    additions: entities.filter(({ op }) => op === 'add').length,
    removals: entities.filter(({ op }) => op === 'remove').length,
    changed: entities.filter(({ op }) => op === 'change').length,
    entities,
  }
}

const ArtifactChange = ({
  change,
  configSections,
  orgId,
  appId,
}: {
  change: TBundleArtifactChange
  configSections: DiffSectionData[]
  orgId: string
  appId: string
}) => {
  const artifact = change.current ?? change.previous
  const display = artifactDisplay(change)
  const href = artifactHref(artifact, orgId, appId)
  const configSection = configSectionForArtifact(change, configSections)
  const contentChanged = change.previous?.digest !== change.current?.digest
  const configChanged =
    change.previous?.config_digest !== change.current?.config_digest
  const summary =
    change.change === 'added'
      ? 'Added to this bundle'
      : change.change === 'removed'
        ? 'Removed from this bundle'
        : contentChanged && configChanged
          ? 'Content and configuration changed'
          : configChanged
            ? 'Configuration changed'
            : contentChanged
              ? 'Content changed'
              : change.change === 'unchanged'
                ? 'No changes'
                : 'Artifact metadata changed'

  return (
    <Expand
      id={`bundle-change-${change.key.replace(/[^a-zA-Z0-9-_]/g, '-')}`}
      isOpen={change.change !== 'unchanged'}
      interactiveHeading
      toggleLabel={`${change.change === 'unchanged' ? 'View' : 'Hide'} ${change.name} details`}
      className="border rounded-md"
      headerClassName="!p-4"
      heading={
        <div className="flex flex-1 flex-wrap items-center justify-between gap-4 text-left">
          <div className="flex items-center gap-3 min-w-0">
            <Icon variant={display.icon} />
            <span className="flex flex-col min-w-0">
              {href ? (
                <Link href={href} className="truncate font-strong">
                  {display.name}
                  <Icon variant="ArrowSquareOutIcon" size={14} />
                </Link>
              ) : (
                <Text weight="strong" className="truncate">
                  {display.name}
                </Text>
              )}
              <Text variant="subtext" theme="neutral">
                {display.kind} · {summary}
              </Text>
            </span>
          </div>
          <Badge theme={changeTheme[change.change]}>{change.change}</Badge>
        </div>
      }
    >
      <div className="flex flex-col gap-4 border-t p-4">
        <DigestChange
          label="Content"
          previous={change.previous?.digest}
          current={change.current?.digest}
        />
        <DigestChange
          label="Configuration"
          previous={change.previous?.config_digest}
          current={change.current?.config_digest}
        />
        <div className="flex flex-wrap gap-x-12 gap-y-4">
          {artifact?.repository ? (
            <LabeledValue label="Repository">
              <Text variant="subtext" className="break-all">
                {artifact.repository}
              </Text>
            </LabeledValue>
          ) : null}
          {artifact?.media_type ? (
            <LabeledValue label="Media type">
              <Text variant="subtext" className="break-all">
                {artifact.media_type}
              </Text>
            </LabeledValue>
          ) : null}
          <LabeledValue label="Size">
            <Text variant="subtext">
              {artifact?.size ? formatBytes(artifact.size) : '—'}
            </Text>
          </LabeledValue>
        </div>
        {configSection ? (
          <div className="flex flex-col gap-2">
            <Text variant="label" theme="neutral">
              Configuration diff
            </Text>
            <AppConfigDiffComponent
              sections={[configSection]}
              summary={null}
              defaultSectionsOpen
            />
          </div>
        ) : change.change !== 'unchanged' && configSections.length ? (
          <Text variant="subtext" theme="neutral">
            This artifact changed, but its app configuration did not.
          </Text>
        ) : null}
      </div>
    </Expand>
  )
}

export const BundleComparison = ({
  bundle,
  previousBundle,
  configSections = [],
  orgId,
  appId,
}: {
  bundle: TCustomerManagedBundle
  previousBundle?: TCustomerManagedBundle
  configSections?: DiffSectionData[]
  orgId: string
  appId: string
}) => {
  const [showUnchanged, setShowUnchanged] = useState(false)
  const changes = compareBundleArtifacts(
    previousBundle?.artifacts ?? [],
    bundle.artifacts ?? []
  )
  const counts = changes.reduce(
    (result, change) => ({
      ...result,
      [change.change]: result[change.change] + 1,
    }),
    { added: 0, changed: 0, removed: 0, unchanged: 0 }
  )
  const reusedSize = changes
    .filter(({ change }) => change === 'unchanged')
    .reduce((total, change) => total + (change.current?.size ?? 0), 0)
  const visibleChanges = changes.filter(
    ({ change }) => showUnchanged || change !== 'unchanged'
  )
  const groups = changeGroups
    .map((group) => ({
      ...group,
      changes: visibleChanges.filter(({ kind }) => group.kinds.includes(kind)),
    }))
    .filter(({ changes }) => changes.length > 0)
  const knownKinds = new Set(changeGroups.flatMap(({ kinds }) => kinds))
  const otherChanges = visibleChanges.filter(
    ({ kind }) => !knownKinds.has(kind)
  )
  if (otherChanges.length) {
    groups.push({
      label: 'Other',
      description: 'Additional bundle artifacts',
      kinds: [],
      changes: otherChanges,
    })
  }

  return (
    <Card className="!p-4 !gap-4">
      <HeadingGroup>
        <Text variant="base" weight="strong">
          {previousBundle ? 'Compared with previous bundle' : 'Initial bundle'}
        </Text>
        <Text variant="subtext" theme="neutral">
          {previousBundle
            ? 'Compare this bundle with the previously published bundle.'
            : 'This is the first published bundle. All contents are new.'}
        </Text>
      </HeadingGroup>

      <div className="flex flex-wrap items-center gap-3 py-2">
        <span className="flex flex-col gap-1 min-w-0">
          <Text variant="label" theme="neutral">
            Previous
          </Text>
          {previousBundle?.manifest_digest ? (
            <Hash hash={previousBundle.manifest_digest} length={16} />
          ) : (
            <Text variant="subtext">None</Text>
          )}
        </span>
        <Icon variant="ArrowRightIcon" />
        <span className="flex flex-col gap-1 min-w-0">
          <Text variant="label" theme="neutral">
            This bundle
          </Text>
          <Hash hash={bundle.manifest_digest ?? bundle.id} length={16} />
        </span>
      </div>

      <div className="flex flex-wrap gap-x-10 gap-y-4 border-y py-4">
        {(['changed', 'added', 'removed', 'unchanged'] as const).map(
          (change) => (
            <LabeledValue key={change} label={toSentenceCase(change)}>
              <Text variant="h3" weight="strong">
                {counts[change]}
              </Text>
            </LabeledValue>
          )
        )}
        <LabeledValue label="Unchanged artifact size">
          <Text variant="h3" weight="strong">
            {formatBytes(reusedSize)}
          </Text>
        </LabeledValue>
      </div>

      <div className="flex items-center justify-end gap-4">
        {counts.unchanged ? (
          <Button
            size="sm"
            variant="secondary"
            onClick={() => setShowUnchanged((visible) => !visible)}
          >
            {showUnchanged
              ? 'Hide unchanged'
              : `Show unchanged (${counts.unchanged})`}
          </Button>
        ) : null}
      </div>

      {groups.length ? (
        <div className="flex flex-col gap-6">
          {groups.map((group) => (
            <section className="flex flex-col gap-3" key={group.label}>
              <div className="flex items-center justify-between gap-4">
                <HeadingGroup>
                  <Text weight="strong">{group.label}</Text>
                  <Text variant="subtext" theme="neutral">
                    {group.description}
                  </Text>
                </HeadingGroup>
                <Badge theme="info">{group.changes.length}</Badge>
              </div>
              <div className="flex flex-col gap-2">
                {group.changes.map((change) => (
                  <ArtifactChange
                    change={change}
                    configSections={configSections}
                    orgId={orgId}
                    appId={appId}
                    key={change.key}
                  />
                ))}
              </div>
            </section>
          ))}
        </div>
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No bundle changes"
          emptyMessage="This bundle contains the same recorded artifacts as the previous bundle."
        />
      )}

      <div className="flex flex-col gap-2 rounded-md border bg-cool-grey-50 p-4 dark:bg-dark-grey-900">
        <Text weight="strong">Deployment plans are environment-specific</Text>
        <Text variant="subtext" theme="neutral">
          This page compares immutable bundle contents and app configuration. A
          Terraform, Helm, or Kubernetes plan is generated only after a customer
          stages this bundle for a registered install.
        </Text>
        <Link href={`/${orgId}/apps/${appId}/installs`} className="font-strong">
          View registered installs
          <Icon variant="ArrowRightIcon" size={14} />
        </Link>
      </div>
    </Card>
  )
}
