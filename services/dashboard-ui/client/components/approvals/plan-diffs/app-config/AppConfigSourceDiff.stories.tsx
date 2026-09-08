import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type ReactNode,
} from 'react'
import type { TAppConfigDiffOperation, TAppConfigDiffSection } from '@/types'
import { cn } from '@/utils/classnames'
import { useDisclosure } from '@/lite/hooks/use-disclosure'
import { changeCounts, emptyDiffSummary } from '@/lite/lib/diffs'
import { ComponentDocs } from '@/lite/components/__stories__/ComponentDocs'
import { Badge } from '@/lite/components/atoms/Badge'
import { Button } from '@/lite/components/atoms/Button'
import { Card } from '@/lite/components/atoms/Card'
import { Icon } from '@/lite/components/atoms/Icon'
import { Text } from '@/lite/components/atoms/Text'
import { Tooltip } from '@/lite/components/atoms/Tooltip'
import { Diff, type TDiffView } from '@/lite/components/molecules/Diff'
import { DiffSummary } from '@/lite/components/molecules/DiffSummary'
import { Disclosure } from '@/lite/components/molecules/Disclosure'
import {
  DisclosureGroup,
  ExpandAllButton,
} from '@/lite/components/molecules/DisclosureGroup'
import '@/lite/styles.css'

export default {
  title: 'Approvals/PlanDiffs/AppConfigSourceDiff (exploration)',
}

// ---------------------------------------------------------------------------
// Mock data — design exploration only. Shapes mirror pkg/config/source_archive.go
// (SourceArchive.Files / SourceArchive.Members) and ComputeAppConfigDiffOutput.
// ---------------------------------------------------------------------------

type TSourceFile = {
  path: string
  op: TAppConfigDiffOperation
  before?: string
  after?: string
  /** Backend signal: text changed but the parsed config is identical. */
  formatOnly?: boolean
}

type TSourceArchiveDiff = {
  baselineVersion: string
  headVersion: string
  /** False when the baseline AppConfig predates source capture (no archive). */
  baselineCaptured: boolean
  /** Head archive Members index: "kind:logicalName" -> relative path. */
  members: Record<string, string>
  files: TSourceFile[]
  unchangedPaths: string[]
}

const whoamiBefore = `name = "whoami"
type = "helm_chart"
namespace = "whoami"

[helm_chart]
chart_name = "whoami"
repo_url = "https://helm.traefik.io/whoami"
version = "3.1.0"

[[helm_chart.values]]
name = "replicaCount"
value = "1"

[[helm_chart.values]]
name = "image.tag"
value = "v1.10.1"

[helm_chart.values_files]
paths = ["values/base.yaml"]
`

const whoamiAfter = `name = "whoami"
type = "helm_chart"
namespace = "whoami"
dependencies = ["redis"]

[helm_chart]
chart_name = "whoami"
repo_url = "https://helm.traefik.io/whoami"
version = "3.2.0"

[[helm_chart.values]]
name = "replicaCount"
value = "3"

[[helm_chart.values]]
name = "image.tag"
value = "v1.10.2"

[[helm_chart.values]]
name = "redis.host"
value = "{{.nuon.components.redis.outputs.host}}"

[helm_chart.values_files]
paths = ["values/base.yaml", "values/prod.yaml"]
`

const inputsBefore = `[[groups]]
name = "cluster"
display_name = "Cluster"

[[inputs]]
name = "region"
group = "cluster"
default = "us-west-2"
required = true

[[inputs]]
name = "cluster_name"
group = "cluster"
required = true
`

const inputsAfter = `[[groups]]
name = "cluster"
display_name = "Cluster"

[[groups]]
name = "cache"
display_name = "Cache"

[[inputs]]
name = "region"
group = "cluster"
default = "us-west-2"
required = true

[[inputs]]
name = "cluster_name"
group = "cluster"
required = true

[[inputs]]
name = "redis_node_type"
group = "cache"
default = "cache.t4g.small"
required = false
sensitive = false
`

const permissionsUnchangedStatements = `
[[statements]]
sid = "DescribeNodegroups"
effect = "Allow"
actions = ["eks:DescribeNodegroup", "eks:ListNodegroups"]
resources = ["*"]

[[statements]]
sid = "ReadClusterLogs"
effect = "Allow"
actions = ["logs:DescribeLogGroups", "logs:GetLogEvents", "logs:FilterLogEvents"]
resources = ["arn:aws:logs:*:*:log-group:/aws/eks/*"]

[[statements]]
sid = "ReadParameters"
effect = "Allow"
actions = ["ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath"]
resources = ["arn:aws:ssm:*:*:parameter/whoami/*"]

[[statements]]
sid = "ReadSecrets"
effect = "Allow"
actions = ["secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"]
resources = ["arn:aws:secretsmanager:*:*:secret:whoami/*"]
`

const permissionsBefore = `name = "eks-access"
description = "Read-only access to the install cluster"

[[statements]]
sid = "DescribeCluster"
effect = "Allow"
actions = ["eks:DescribeCluster", "eks:ListClusters"]
resources = ["*"]
${permissionsUnchangedStatements}`

const permissionsAfter = `name = "eks-access"
description = "Access to the install cluster and the cache subnet group"

[[statements]]
sid = "DescribeCluster"
effect = "Allow"
actions = ["eks:DescribeCluster", "eks:ListClusters"]
resources = ["*"]
${permissionsUnchangedStatements}
[[statements]]
sid = "CacheSubnets"
effect = "Allow"
actions = ["elasticache:DescribeCacheSubnetGroups", "elasticache:CreateCacheSubnetGroup"]
resources = ["*"]
`

const redisAfter = `name = "redis"
type = "helm_chart"
namespace = "cache"

[helm_chart]
chart_name = "redis"
repo_url = "https://charts.bitnami.com/bitnami"
version = "19.6.4"

[[helm_chart.values]]
name = "architecture"
value = "standalone"

[[helm_chart.values]]
name = "master.persistence.size"
value = "8Gi"
`

const legacyWorkerBefore = `name = "legacy-worker"
type = "docker_build"

[docker_build]
dockerfile = "Dockerfile.worker"
build_args = ["TARGET=worker"]

[docker_build.public_repo]
repo = "https://github.com/acme/legacy-worker"
directory = "."
branch = "main"
`

const healthcheckBefore = `name="healthcheck"
type="kubernetes_manifest"
timeout="60s"
[kubernetes_manifest]
manifest="manifests/healthcheck.yaml"
namespace="whoami"
`

const healthcheckAfter = `name = "healthcheck"
type = "kubernetes_manifest"
timeout = "60s"

# Applies the smoke-test job after every deploy.
[kubernetes_manifest]
namespace = "whoami"
manifest = "manifests/healthcheck.yaml"
`

const mockMembers: Record<string, string> = {
  'metadata:metadata': 'nuon.toml',
  'input:inputs': 'inputs.toml',
  'component:whoami': 'components/whoami.toml',
  'component:redis': 'components/redis.toml',
  'component:frontend': 'components/frontend.toml',
  'permission:eks-access': 'permissions/eks-access.toml',
  'action:healthcheck': 'actions/healthcheck.toml',
  'action:backup': 'actions/backup.toml',
}

const mockUnchangedPaths = [
  'nuon.toml',
  'runner.toml',
  'components/frontend.toml',
  'actions/backup.toml',
]

const mockSourceDiff: TSourceArchiveDiff = {
  baselineVersion: 'v12',
  headVersion: 'v13',
  baselineCaptured: true,
  members: mockMembers,
  unchangedPaths: mockUnchangedPaths,
  files: [
    {
      path: 'inputs.toml',
      op: 'change',
      before: inputsBefore,
      after: inputsAfter,
    },
    {
      path: 'components/whoami.toml',
      op: 'change',
      before: whoamiBefore,
      after: whoamiAfter,
    },
    { path: 'components/redis.toml', op: 'add', after: redisAfter },
    {
      path: 'components/legacy-worker.toml',
      op: 'remove',
      before: legacyWorkerBefore,
    },
    {
      path: 'permissions/eks-access.toml',
      op: 'change',
      before: permissionsBefore,
      after: permissionsAfter,
    },
    {
      path: 'actions/healthcheck.toml',
      op: 'change',
      before: healthcheckBefore,
      after: healthcheckAfter,
      formatOnly: true,
    },
  ],
}

const degradedSourceDiff: TSourceArchiveDiff = {
  baselineVersion: 'v12',
  headVersion: 'v13',
  baselineCaptured: false,
  members: mockMembers,
  unchangedPaths: [],
  files: [],
}

const identicalSourceDiff: TSourceArchiveDiff = {
  baselineVersion: 'v13',
  headVersion: 'v14',
  baselineCaptured: true,
  members: mockMembers,
  unchangedPaths: [
    ...mockUnchangedPaths,
    'inputs.toml',
    'components/whoami.toml',
    'components/redis.toml',
    'permissions/eks-access.toml',
    'actions/healthcheck.toml',
  ],
  files: [],
}

const mockComputedSections: TAppConfigDiffSection[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 1,
    removals: 1,
    changed: 1,
    entities: [
      {
        name: 'whoami',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          { key: 'dependencies', op: 'add', diff: "'' -> '[redis]'" },
          {
            key: 'helm_chart.version',
            op: 'change',
            diff: "'3.1.0' -> '3.2.0'",
          },
          { key: 'values.replicaCount', op: 'change', diff: "'1' -> '3'" },
        ],
      },
      {
        name: 'redis',
        op: 'add',
        componentType: 'helm_chart',
        fields: [{ key: 'type', op: 'add', diff: "'' -> 'helm_chart'" }],
      },
      {
        name: 'legacy-worker',
        op: 'remove',
        componentType: 'docker_build',
        fields: [{ key: 'type', op: 'remove', diff: "'docker_build' -> ''" }],
      },
    ],
    fields: [],
  },
  {
    name: 'Install inputs',
    sectionKey: 'inputs',
    grouped: true,
    additions: 1,
    removals: 0,
    changed: 0,
    entities: [
      {
        name: 'redis_node_type',
        op: 'add',
        fields: [
          { key: 'default', op: 'add', diff: "'' -> 'cache.t4g.small'" },
        ],
      },
    ],
    fields: [],
  },
  {
    name: 'Permissions',
    sectionKey: 'permissions',
    grouped: false,
    additions: 0,
    removals: 0,
    changed: 1,
    entities: [],
    fields: [
      {
        key: 'eks-access.description',
        op: 'change',
        diff: "'Read-only access to the install cluster' -> 'Access to the install cluster and the cache subnet group'",
      },
      {
        key: 'eks-access.statements',
        op: 'add',
        diff: "+1 statement 'CacheSubnets'",
      },
    ],
  },
]

const mockComputedSummary = { added: 2, removed: 1, changed: 2 }

// ---------------------------------------------------------------------------
// Local helpers (mock-quality). In production these belong in lib/diffs.
// ---------------------------------------------------------------------------

const SECTION_KIND: Record<string, string> = {
  components: 'component',
  actions: 'action',
  permissions: 'permission',
  runbooks: 'runbook',
  stacks: 'stack',
  secrets: 'secret',
}

const pathForEntity = (
  members: Record<string, string>,
  sectionKey: string,
  name: string
): string | undefined => {
  if (sectionKey === 'inputs') return members['input:inputs']
  const kind = SECTION_KIND[sectionKey] ?? sectionKey.replace(/s$/, '')
  return members[`${kind}:${name}`]
}

const memberForPath = (members: Record<string, string>, path: string) => {
  const entry = Object.entries(members).find(([, p]) => p === path)
  if (!entry) return undefined
  const [kind, name] = entry[0].split(':')
  return { kind, name }
}

const splitPath = (path: string) => {
  const index = path.lastIndexOf('/')
  return index < 0
    ? { dir: '', base: path }
    : { dir: path.slice(0, index + 1), base: path.slice(index + 1) }
}

const fileCounts = (file: TSourceFile) =>
  changeCounts(file.before ?? '', file.after ?? '')

const OP_TEXT: Record<TAppConfigDiffOperation, string> = {
  add: 'text-diff-add',
  change: 'text-diff-change',
  remove: 'text-diff-remove',
}

const OP_RAIL: Record<TAppConfigDiffOperation, string> = {
  add: 'border-l-diff-add',
  change: 'border-l-diff-change',
  remove: 'border-l-diff-remove',
}

const OP_LETTER: Record<TAppConfigDiffOperation, string> = {
  add: 'A',
  change: 'M',
  remove: 'D',
}

const OP_LABEL: Record<TAppConfigDiffOperation, string> = {
  add: 'Added',
  change: 'Modified',
  remove: 'Removed',
}

type TFocus = { path: string; nonce: number }

const StatusLetter = ({
  op,
  muted = false,
}: {
  op: TAppConfigDiffOperation
  muted?: boolean
}) => (
  <Text
    as="span"
    variant="caption"
    family="mono"
    weight="semibold"
    aria-label={OP_LABEL[op]}
    className={cn(
      'w-3 shrink-0 text-center',
      muted ? 'text-tertiary' : OP_TEXT[op]
    )}
  >
    {OP_LETTER[op]}
  </Text>
)

const Counts = ({
  added,
  removed,
  muted = false,
}: {
  added: number
  removed: number
  muted?: boolean
}) => (
  <span
    className="flex items-center gap-1.5 tabular-nums"
    aria-label={`${added} added, ${removed} removed`}
  >
    {added ? (
      <Text
        variant="caption"
        family="mono"
        className={muted ? 'text-tertiary' : 'text-diff-add'}
      >
        +{added}
      </Text>
    ) : null}
    {removed ? (
      <Text
        variant="caption"
        family="mono"
        className={muted ? 'text-tertiary' : 'text-diff-remove'}
      >
        -{removed}
      </Text>
    ) : null}
  </span>
)

const MonoPath = ({
  path,
  className,
}: {
  path: string
  className?: string
}) => {
  const { dir, base } = splitPath(path)
  return (
    <Text
      as="span"
      variant="caption"
      family="mono"
      className={cn('truncate', className)}
    >
      {dir ? <span className="text-tertiary">{dir}</span> : null}
      <span className="text-primary">{base}</span>
    </Text>
  )
}

// ---------------------------------------------------------------------------
// Parsed summary strip — the semantic diff, demoted to orientation + cross-links.
// ---------------------------------------------------------------------------

type TSummaryChip = {
  name: string
  op: TAppConfigDiffOperation
  detail?: string
  path?: string
}

const chipsForSection = (
  section: TAppConfigDiffSection,
  members: Record<string, string>,
  changedPaths: Set<string>
): TSummaryChip[] => {
  const resolve = (name: string) => {
    const path = pathForEntity(members, section.sectionKey, name)
    return path && changedPaths.has(path) ? path : undefined
  }

  if (section.grouped) {
    return section.entities.map((entity) => ({
      name: entity.name,
      op: entity.op,
      detail: entity.componentType?.replace('_', ' '),
      path: resolve(entity.name),
    }))
  }

  const names = [
    ...new Set(section.fields.map((field) => field.key.split('.')[0])),
  ]
  return names.map((name) => ({ name, op: 'change', path: resolve(name) }))
}

const SummaryChip = ({
  chip,
  onFocusFile,
}: {
  chip: TSummaryChip
  onFocusFile: (path: string) => void
}) => {
  const content = (
    <>
      <span
        aria-hidden
        className={cn('size-1.5 rounded-full bg-current', OP_TEXT[chip.op])}
      />
      <Text variant="caption" weight="medium" color="primary">
        {chip.name}
      </Text>
      {chip.detail ? (
        <Text variant="caption" color="tertiary">
          {chip.detail}
        </Text>
      ) : null}
    </>
  )

  if (!chip.path) {
    return (
      <Tooltip content={`No source file for ${chip.name} in the head config`}>
        <span className="inline-flex h-7 items-center gap-1.5 rounded-md border border-dashed border-divider px-2">
          {content}
        </span>
      </Tooltip>
    )
  }

  return (
    <Tooltip content={`Open ${chip.path}`}>
      <button
        type="button"
        onClick={() => onFocusFile(chip.path as string)}
        className={cn(
          'inline-flex h-7 cursor-pointer items-center gap-1.5 rounded-md border border-divider bg-surface-01 px-2 transition-colors',
          'hover:border-divider-accent/60 hover:bg-menu-item-hover focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-focus-ring'
        )}
      >
        {content}
        <Icon
          variant="ArrowLineRightIcon"
          size={12}
          className="text-tertiary"
        />
      </button>
    </Tooltip>
  )
}

const ParsedSummary = ({
  sections,
  summary,
  members,
  changedPaths,
  onFocusFile,
}: {
  sections: TAppConfigDiffSection[]
  summary: { added: number; removed: number; changed: number }
  members: Record<string, string>
  changedPaths: Set<string>
  onFocusFile: (path: string) => void
}) => (
  <Disclosure
    defaultOpen
    title="Summary"
    description="What the parsed config says changed"
    className="rounded-lg bg-surface-02"
    contentClassName="px-3 pb-3"
    status={
      <DiffSummary
        summary={{
          ...emptyDiffSummary(),
          create: summary.added,
          update: summary.changed,
          delete: summary.removed,
        }}
        operations={['create', 'update', 'delete']}
        className="gap-x-3"
      />
    }
  >
    <dl className="flex flex-col gap-2">
      {sections.map((section) => (
        <div
          key={section.sectionKey}
          className="grid grid-cols-[8rem_1fr] items-start gap-x-3 gap-y-1 sm:grid-cols-[9rem_1fr]"
        >
          <dt className="flex items-center gap-1.5 pt-1.5">
            <Text variant="caption" color="secondary" weight="medium">
              {section.name}
            </Text>
          </dt>
          <dd className="flex flex-wrap items-center gap-1.5">
            {chipsForSection(section, members, changedPaths).map((chip) => (
              <SummaryChip
                key={chip.name}
                chip={chip}
                onFocusFile={onFocusFile}
              />
            ))}
          </dd>
        </div>
      ))}
    </dl>
  </Disclosure>
)

// ---------------------------------------------------------------------------
// Changed-files tree — @pierre/trees look (status letter lane, folder dots)
// plus +/− counts as row decoration.
// ---------------------------------------------------------------------------

type TTreeFolder = { dir: string; files: TSourceFile[]; unchanged: string[] }

const buildTree = (
  files: TSourceFile[],
  unchangedPaths: string[]
): TTreeFolder[] => {
  const folders = new Map<string, TTreeFolder>()
  const folder = (dir: string) => {
    const existing = folders.get(dir)
    if (existing) return existing
    const created = { dir, files: [], unchanged: [] }
    folders.set(dir, created)
    return created
  }
  files.forEach((file) => folder(splitPath(file.path).dir).files.push(file))
  unchangedPaths.forEach((path) =>
    folder(splitPath(path).dir).unchanged.push(path)
  )
  return [...folders.values()].sort((a, b) =>
    a.dir === '' ? 1 : b.dir === '' ? -1 : a.dir.localeCompare(b.dir)
  )
}

const TreeRow = ({
  depth,
  selected,
  onClick,
  children,
}: {
  depth: number
  selected?: boolean
  onClick?: () => void
  children: ReactNode
}) => (
  <button
    type="button"
    disabled={!onClick}
    onClick={onClick}
    style={{ paddingLeft: `${8 + depth * 14}px` } as CSSProperties}
    className={cn(
      'flex h-7 w-full min-w-0 items-center gap-1.5 rounded-md pr-2 text-left transition-colors',
      onClick && 'cursor-pointer hover:bg-menu-item-hover',
      selected && 'bg-menu-item-selected',
      'focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-focus-ring'
    )}
  >
    {children}
  </button>
)

const FileTree = ({
  diff,
  focusPath,
  onFocusFile,
}: {
  diff: TSourceArchiveDiff
  focusPath?: string
  onFocusFile: (path: string) => void
}) => {
  const [showUnchanged, setShowUnchanged] = useState(false)
  const folders = useMemo(
    () => buildTree(diff.files, showUnchanged ? diff.unchangedPaths : []),
    [diff.files, diff.unchangedPaths, showUnchanged]
  )
  const totals = useMemo(
    () =>
      diff.files.reduce(
        (sum, file) => {
          const counts = fileCounts(file)
          return {
            added: sum.added + counts.added,
            removed: sum.removed + counts.removed,
          }
        },
        { added: 0, removed: 0 }
      ),
    [diff.files]
  )

  return (
    <nav
      aria-label="Changed files"
      className="flex flex-col gap-1 rounded-lg bg-surface-02 p-2"
    >
      <div className="flex items-center justify-between gap-2 px-2 py-1">
        <Text variant="caption" color="tertiary" weight="medium">
          {diff.files.length} changed{' '}
          {diff.files.length === 1 ? 'file' : 'files'}
        </Text>
        <Counts added={totals.added} removed={totals.removed} />
      </div>

      <div className="flex flex-col gap-px">
        {folders.map((folder) => (
          <div key={folder.dir || '/'} className="flex flex-col gap-px">
            {folder.dir ? (
              <TreeRow depth={0}>
                <Icon
                  variant="CaretDownIcon"
                  size={12}
                  className="text-tertiary"
                />
                <Text variant="caption" color="secondary" className="truncate">
                  {folder.dir.replace(/\/$/, '')}
                </Text>
                {folder.files.length ? (
                  <span
                    aria-hidden
                    className="ml-auto size-1.5 rounded-full bg-status-neutral/50"
                  />
                ) : null}
              </TreeRow>
            ) : null}
            {folder.files.map((file) => {
              const counts = fileCounts(file)
              return (
                <TreeRow
                  key={file.path}
                  depth={folder.dir ? 1 : 0}
                  selected={focusPath === file.path}
                  onClick={() => onFocusFile(file.path)}
                >
                  <Icon
                    variant="FileIcon"
                    size={14}
                    className={cn(
                      'shrink-0',
                      file.formatOnly ? 'text-tertiary' : OP_TEXT[file.op]
                    )}
                  />
                  <Text
                    variant="caption"
                    family="mono"
                    className={cn(
                      'truncate',
                      file.formatOnly ? 'text-secondary' : 'text-primary'
                    )}
                  >
                    {splitPath(file.path).base}
                  </Text>
                  <span className="ml-auto flex shrink-0 items-center gap-2">
                    <Counts
                      added={counts.added}
                      removed={counts.removed}
                      muted={file.formatOnly}
                    />
                    <StatusLetter op={file.op} muted={file.formatOnly} />
                  </span>
                </TreeRow>
              )
            })}
            {folder.unchanged.map((path) => (
              <TreeRow key={path} depth={folder.dir ? 1 : 0}>
                <Icon
                  variant="FileIcon"
                  size={14}
                  className="shrink-0 text-tertiary/60"
                />
                <Text
                  variant="caption"
                  family="mono"
                  color="tertiary"
                  className="truncate"
                >
                  {splitPath(path).base}
                </Text>
              </TreeRow>
            ))}
          </div>
        ))}
      </div>

      {diff.unchangedPaths.length ? (
        <button
          type="button"
          onClick={() => setShowUnchanged((current) => !current)}
          aria-pressed={showUnchanged}
          className="mt-1 flex h-7 cursor-pointer items-center gap-1.5 rounded-md px-2 text-left transition-colors hover:bg-menu-item-hover focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-focus-ring"
        >
          <Icon
            variant={showUnchanged ? 'MinusCircleIcon' : 'PlusIcon'}
            size={12}
            className="text-tertiary"
          />
          <Text variant="caption" color="tertiary">
            {showUnchanged
              ? 'Hide unchanged files'
              : `${diff.unchangedPaths.length} unchanged ${diff.unchangedPaths.length === 1 ? 'file' : 'files'} not shown`}
          </Text>
        </button>
      ) : null}
    </nav>
  )
}

// ---------------------------------------------------------------------------
// Stacked per-file diff with a sticky header. Body is the house Diff molecule
// (@pierre/diffs: Shiki TOML highlighting, word-level intra-line diff, collapsed
// unchanged runs with line-info separators).
// ---------------------------------------------------------------------------

const FileDiffPanel = ({
  file,
  members,
  view,
  focus,
  defaultOpen,
}: {
  file: TSourceFile
  members: Record<string, string>
  view: TDiffView
  focus?: TFocus
  defaultOpen?: boolean
}) => {
  const ref = useRef<HTMLElement>(null)
  const [flash, setFlash] = useState(false)
  const { open, setOpen, triggerProps, contentProps } = useDisclosure({
    id: `file-${file.path}`,
    defaultOpen,
  })
  const counts = useMemo(() => fileCounts(file), [file])
  const member = memberForPath(members, file.path)
  const focused = focus?.path === file.path

  const nonce = focus?.nonce
  useEffect(() => {
    if (!focused) return
    setOpen(true)
    ref.current?.scrollIntoView({ block: 'start', behavior: 'smooth' })
    setFlash(true)
    const timer = window.setTimeout(() => setFlash(false), 1600)
    return () => window.clearTimeout(timer)
  }, [focused, nonce, setOpen])

  return (
    <section
      ref={ref}
      aria-label={file.path}
      className={cn(
        'scroll-mt-4 rounded-lg border border-l-4 border-divider bg-surface-01 transition-shadow duration-300',
        file.formatOnly ? 'border-l-diff-neutral' : OP_RAIL[file.op],
        flash && 'shadow-[0_0_0_2px_var(--divider-accent)]'
      )}
    >
      <header
        className={cn(
          'sticky top-0 z-10 flex min-h-10 items-center gap-2 rounded-tr-lg border-b border-divider bg-surface-01/95 px-2 backdrop-blur-sm',
          !open && 'rounded-br-lg border-b-0'
        )}
      >
        <button
          type="button"
          {...triggerProps}
          className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md py-1.5 pr-1 text-left outline-none focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-focus-ring"
        >
          <Icon
            variant="CaretRightIcon"
            size={14}
            className={cn(
              'shrink-0 text-tertiary transition-transform duration-200',
              open && 'rotate-90'
            )}
          />
          <StatusLetter op={file.op} muted={file.formatOnly} />
          <MonoPath path={file.path} />
        </button>

        {member ? (
          <Badge
            variant="code"
            labelKey={member.kind}
            labelValue={member.name}
          />
        ) : null}
        {file.formatOnly ? (
          <Tooltip content="Whitespace, ordering or comments changed. The parsed config is identical.">
            <Badge>Format only</Badge>
          </Tooltip>
        ) : null}
        <Counts
          added={counts.added}
          removed={counts.removed}
          muted={file.formatOnly}
        />
      </header>

      {open ? (
        <div {...contentProps} className="p-2">
          {file.op === 'remove' ? (
            <Text
              as="p"
              variant="caption"
              color="tertiary"
              className="px-1 pb-2"
            >
              File removed in the head config.
            </Text>
          ) : null}
          <Diff
            before={file.before ?? ''}
            after={file.after ?? ''}
            language="toml"
            view={view}
            search={false}
            maxHeight={720}
          />
        </div>
      ) : null}
    </section>
  )
}

// ---------------------------------------------------------------------------
// Card shell
// ---------------------------------------------------------------------------

const VersionRange = ({ from, to }: { from: string; to: string }) => (
  <span className="flex items-center gap-1">
    <Badge variant="code">{from}</Badge>
    <Icon variant="ArrowLineRightIcon" size={12} className="text-tertiary" />
    <Badge variant="code">{to}</Badge>
  </span>
)

const SourceUnavailable = ({ diff }: { diff: TSourceArchiveDiff }) => {
  const headPaths = [...new Set(Object.values(diff.members))].sort()
  return (
    <div className="flex flex-col gap-3 rounded-lg bg-surface-02 px-4 py-8">
      <div className="flex flex-col items-center gap-2 text-center">
        <Icon variant="InfoIcon" size={20} className="text-tertiary" />
        <Text as="p" variant="body" weight="medium">
          Source diff unavailable for {diff.baselineVersion} →{' '}
          {diff.headVersion}
        </Text>
        <Text as="p" variant="caption" color="tertiary" className="max-w-md">
          {diff.baselineVersion} was published before source capture, so there
          is no baseline to compare against. The summary above still reflects
          the parsed config. File-level diffs start with the next config
          version.
        </Text>
      </div>
      <Disclosure
        title={`${headPaths.length} files captured in ${diff.headVersion}`}
        className="mx-auto w-full max-w-md"
        contentClassName="px-2 pb-1"
      >
        <ul className="flex flex-col gap-px">
          {headPaths.map((path) => (
            <li key={path} className="flex h-6 items-center gap-1.5 px-2">
              <Icon variant="FileIcon" size={12} className="text-tertiary/60" />
              <MonoPath path={path} className="text-secondary" />
            </li>
          ))}
        </ul>
      </Disclosure>
    </div>
  )
}

const NoChanges = ({ diff }: { diff: TSourceArchiveDiff }) => (
  <div className="rounded-lg bg-surface-02 px-4 py-8 text-center">
    <Text as="p" variant="body" weight="medium">
      No config changes
    </Text>
    <Text as="p" variant="caption" color="tertiary">
      {diff.headVersion} matches {diff.baselineVersion}.{' '}
      {diff.unchangedPaths.length} files, all identical.
    </Text>
  </div>
)

const AppConfigSourceDiffCard = ({
  sections,
  summary,
  sourceDiff,
  initialFocusPath,
  defaultFilesOpen = true,
  className,
}: {
  sections: TAppConfigDiffSection[]
  summary: { added: number; removed: number; changed: number }
  sourceDiff: TSourceArchiveDiff
  initialFocusPath?: string
  defaultFilesOpen?: boolean
  className?: string
}) => {
  const [focus, setFocus] = useState<TFocus | undefined>(
    initialFocusPath ? { path: initialFocusPath, nonce: 0 } : undefined
  )
  const [view, setView] = useState<TDiffView>('unified')
  const changedPaths = useMemo(
    () => new Set(sourceDiff.files.map((file) => file.path)),
    [sourceDiff.files]
  )
  const onFocusFile = (path: string) =>
    setFocus((current) => ({ path, nonce: (current?.nonce ?? 0) + 1 }))

  const hasSummary = sections.length > 0
  const hasFiles = sourceDiff.files.length > 0
  const split = view === 'split'

  return (
    <Card
      as="section"
      padding="sm"
      className={cn('flex flex-col gap-3', className)}
    >
      <header className="flex flex-wrap items-center justify-between gap-3 px-1">
        <div className="flex items-center gap-3">
          <Text as="h2" variant="heading">
            Config changes
          </Text>
          <VersionRange
            from={sourceDiff.baselineVersion}
            to={sourceDiff.headVersion}
          />
        </div>
        {hasFiles ? (
          <div className="flex items-center gap-0.5">
            <ExpandAllButton />
            <Button
              size="sm"
              variant="ghost"
              iconOnly
              aria-pressed={split}
              aria-label={split ? 'Unified view' : 'Split view'}
              tooltip={split ? 'Unified view' : 'Split view'}
              onClick={() => setView(split ? 'unified' : 'split')}
            >
              <Icon
                variant={
                  split
                    ? 'SquareSplitVerticalIcon'
                    : 'SquareSplitHorizontalIcon'
                }
                size={14}
              />
            </Button>
          </div>
        ) : null}
      </header>

      {hasSummary ? (
        <DisclosureGroup defaultOpen>
          <ParsedSummary
            sections={sections}
            summary={summary}
            members={sourceDiff.members}
            changedPaths={changedPaths}
            onFocusFile={onFocusFile}
          />
        </DisclosureGroup>
      ) : null}

      {!sourceDiff.baselineCaptured ? (
        <SourceUnavailable diff={sourceDiff} />
      ) : !hasFiles ? (
        <NoChanges diff={sourceDiff} />
      ) : (
        <div className="grid gap-3 lg:grid-cols-[15rem_minmax(0,1fr)]">
          <aside className="lg:sticky lg:top-4 lg:self-start">
            <FileTree
              diff={sourceDiff}
              focusPath={focus?.path}
              onFocusFile={onFocusFile}
            />
          </aside>
          <div className="flex min-w-0 flex-col gap-2">
            {sourceDiff.files.map((file) => (
              <FileDiffPanel
                key={file.path}
                file={file}
                members={sourceDiff.members}
                view={view}
                focus={focus}
                defaultOpen={defaultFilesOpen && !file.formatOnly}
              />
            ))}
          </div>
        </div>
      )}
    </Card>
  )
}

const AppConfigSourceDiff = (
  props: Parameters<typeof AppConfigSourceDiffCard>[0]
) => (
  <DisclosureGroup defaultOpen={props.defaultFilesOpen ?? true}>
    <AppConfigSourceDiffCard {...props} />
  </DisclosureGroup>
)

const Frame = ({
  children,
  wide = true,
}: {
  children: ReactNode
  wide?: boolean
}) => {
  useEffect(() => {
    const root = document.documentElement
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const apply = () =>
      root.setAttribute('data-theme', mq.matches ? 'dark' : 'light')
    apply()
    mq.addEventListener('change', apply)
    return () => mq.removeEventListener('change', apply)
  }, [])

  return (
    <div className={cn('p-8', wide ? 'max-w-6xl' : 'max-w-3xl')}>
      {children}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

export const Overview = () => (
  <ComponentDocs
    name="AppConfigSourceDiff"
    tier="organism"
    summary="Design exploration: the app-branch config diff with authored source files as the primary view and the parsed diff as a summary."
    use={[
      'Review a config version change the way it was authored: file by file, with real TOML.',
      'Orient with the parsed summary, then jump to the file that defines an entity.',
    ]}
    avoid={[
      'Do not render head files as all added when the baseline has no source archive.',
      'Do not treat a format-only file as a semantic change. It needs a backend signal.',
      'Do not ship this file. Every helper here is mock-quality and lives in the story.',
    ]}
    rules={[
      'Source is primary. The parsed summary is one collapsible strip above it.',
      'Summary chips resolve to files through the head archive Members index. Entities without a head file render as unlinked chips.',
      'File diffs reuse the lite Diff molecule (@pierre/diffs), so highlighting, word-level diffs and collapsed hunks match every other lite diff.',
      'The file tree follows @pierre/trees conventions: status letter lane (A, M, D), folder dots, muted unchanged rows behind a toggle.',
    ]}
    sections={[
      {
        heading: 'Adopt, not emulate',
        body: 'dashboard-ui already depends on @pierre/diffs (lite/molecules/Diff registers a CSS-variable Shiki theme). @pierre/trees is not a dependency; the tree here is a small Tailwind list that borrows its status-lane look because a static changed-files list does not need virtualization, drag and drop or search.',
      },
      {
        heading: 'Data',
        body: 'Files come from diffing SourceArchive.Files between the baseline and head AppConfig. Members comes from the head archive. formatOnly is a backend signal that the parsed AppConfig.Diff() is empty for that file while the text differs.',
      },
    ]}
  />
)

export const Default = () => (
  <Frame>
    <AppConfigSourceDiff
      sections={mockComputedSections}
      summary={mockComputedSummary}
      sourceDiff={mockSourceDiff}
    />
  </Frame>
)

export const CollapsedFiles = () => (
  <Frame>
    <AppConfigSourceDiff
      sections={mockComputedSections}
      summary={mockComputedSummary}
      sourceDiff={mockSourceDiff}
      defaultFilesOpen={false}
    />
  </Frame>
)

export const FileDrillDown = () => (
  <Frame wide={false}>
    <DisclosureGroup defaultOpen>
      <div className="flex flex-col gap-2">
        <Text as="p" variant="caption" color="tertiary">
          One changed file. Word-level highlights on version, replicaCount and
          paths; the member badge links the file back to the parsed entity.
        </Text>
        <FileDiffPanel
          file={mockSourceDiff.files[1]}
          members={mockMembers}
          view="unified"
          defaultOpen
        />
      </div>
    </DisclosureGroup>
  </Frame>
)

export const FileDrillDownCollapsedHunks = () => (
  <Frame wide={false}>
    <DisclosureGroup defaultOpen>
      <div className="flex flex-col gap-2">
        <Text as="p" variant="caption" color="tertiary">
          A long file with a change at the top and bottom. The unchanged
          statements in the middle collapse behind a line-info separator.
        </Text>
        <FileDiffPanel
          file={mockSourceDiff.files[4]}
          members={mockMembers}
          view="unified"
          defaultOpen
        />
      </div>
    </DisclosureGroup>
  </Frame>
)

export const FileDrillDownSplit = () => (
  <Frame>
    <DisclosureGroup defaultOpen>
      <FileDiffPanel
        file={mockSourceDiff.files[1]}
        members={mockMembers}
        view="split"
        defaultOpen
      />
    </DisclosureGroup>
  </Frame>
)

export const FileDrillDownVariants = () => (
  <Frame wide={false}>
    <DisclosureGroup defaultOpen>
      <div className="flex flex-col gap-2">
        <FileDiffPanel
          file={mockSourceDiff.files[2]}
          members={mockMembers}
          view="unified"
          defaultOpen
        />
        <FileDiffPanel
          file={mockSourceDiff.files[3]}
          members={mockMembers}
          view="unified"
          defaultOpen
        />
        <FileDiffPanel
          file={mockSourceDiff.files[5]}
          members={mockMembers}
          view="unified"
          defaultOpen
        />
      </div>
    </DisclosureGroup>
  </Frame>
)

export const CrossLink = () => (
  <Frame>
    <div className="flex flex-col gap-3">
      <Text as="p" variant="caption" color="tertiary">
        Opened with the whoami component focused, as if the user clicked its
        summary chip: the tree row is selected, the file is expanded and briefly
        ringed. legacy-worker has no head file, so its chip is unlinked.
      </Text>
      <AppConfigSourceDiff
        sections={mockComputedSections}
        summary={mockComputedSummary}
        sourceDiff={mockSourceDiff}
        initialFocusPath="components/whoami.toml"
        defaultFilesOpen={false}
      />
    </div>
  </Frame>
)

export const DegradedNoBaselineArchive = () => (
  <Frame>
    <AppConfigSourceDiff
      sections={mockComputedSections}
      summary={mockComputedSummary}
      sourceDiff={degradedSourceDiff}
    />
  </Frame>
)

export const NoFileChanges = () => (
  <Frame>
    <AppConfigSourceDiff
      sections={[]}
      summary={{ added: 0, removed: 0, changed: 0 }}
      sourceDiff={identicalSourceDiff}
    />
  </Frame>
)
