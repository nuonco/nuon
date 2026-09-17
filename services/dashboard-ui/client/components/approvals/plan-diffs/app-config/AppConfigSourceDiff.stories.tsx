import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import type { TAppConfigDiffOperation, TAppConfigDiffSection } from '@/types'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { diffLines } from '@/utils/code-utils'
import { cn } from '@/utils/classnames'
import { DiffCodeBlock, WrapLinesProvider } from '../wrap-lines-context'

export default {
  title: 'Approvals/PlanDiffs/AppConfigSourceDiff (exploration)',
}

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
    ? { dir: '', base: path.slice(index + 1) }
    : { dir: path.slice(0, index + 1), base: path.slice(index + 1) }
}

const MUTED_TEXT = 'text-cool-grey-500 dark:text-dark-grey-400'
const SURFACE_02 = 'bg-cool-grey-100 dark:bg-dark-grey-800'
const BORDER = 'border-cool-grey-300 dark:border-dark-grey-500'

const fileCounts = (file: TSourceFile) => {
  const text = diffLines(file.before ?? '', file.after ?? '')
  const counts = { added: 0, removed: 0 }
  for (const line of text.split('\n')) {
    if (line.startsWith('+')) counts.added++
    else if (line.startsWith('-')) counts.removed++
  }
  return counts
}

const fileDiffText = (file: TSourceFile) =>
  diffLines(file.before ?? '', file.after ?? '')

const OP_TEXT: Record<TAppConfigDiffOperation, string> = {
  add: 'text-green-800 dark:text-green-500',
  change: 'text-orange-800 dark:text-orange-400',
  remove: 'text-red-800 dark:text-red-500',
}

const OP_RAIL: Record<TAppConfigDiffOperation, string> = {
  add: 'bg-green-400 dark:bg-green-500/40',
  change: 'bg-orange-300 dark:bg-orange-500/40',
  remove: 'bg-red-300 dark:bg-red-500/40',
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

const ExpandAllContext = createContext<{
  value: boolean | null
  toggle: () => void
} | null>(null)

const useExpandAll = () => useContext(ExpandAllContext)

const ExpandAllToggle = () => {
  const ctx = useExpandAll()
  if (!ctx) return null
  const isAllExpanded = ctx.value === true
  return (
    <Button
      className="!p-1 flex items-center gap-1.5"
      variant="ghost"
      size="sm"
      aria-pressed={isAllExpanded}
      onClick={ctx.toggle}
    >
      {isAllExpanded ? 'Collapse all' : 'Expand all'}
      <Icon
        variant={isAllExpanded ? 'CaretUpIcon' : 'CaretDownIcon'}
        size={14}
      />
    </Button>
  )
}

const StatusLetter = ({
  op,
  muted = false,
}: {
  op: TAppConfigDiffOperation
  muted?: boolean
}) => (
  <span
    aria-label={OP_LABEL[op]}
    className={cn(
      'w-3 shrink-0 text-center font-mono text-[11px] leading-[14px] font-semibold',
      muted ? MUTED_TEXT : OP_TEXT[op]
    )}
  >
    {OP_LETTER[op]}
  </span>
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
      <span
        className={cn(
          'font-mono text-[11px] leading-[14px]',
          muted ? MUTED_TEXT : 'text-green-800 dark:text-green-500'
        )}
      >
        +{added}
      </span>
    ) : null}
    {removed ? (
      <span
        className={cn(
          'font-mono text-[11px] leading-[14px]',
          muted ? MUTED_TEXT : 'text-red-800 dark:text-red-500'
        )}
      >
        -{removed}
      </span>
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
    <span
      className={cn('truncate font-mono text-[11px] leading-[14px]', className)}
    >
      {dir ? <span className={MUTED_TEXT}>{dir}</span> : null}
      <span>{base}</span>
    </span>
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
      <span className="text-[11px] leading-[14px] font-medium">
        {chip.name}
      </span>
      {chip.detail ? (
        <span className={cn('text-[11px] leading-[14px]', MUTED_TEXT)}>
          {chip.detail}
        </span>
      ) : null}
    </>
  )

  if (!chip.path) {
    return (
      <Tooltip
        tipContent={`No source file for ${chip.name} in the head config`}
      >
        <span className="inline-flex h-7 items-center gap-1.5 rounded-md border border-dashed px-2">
          {content}
        </span>
      </Tooltip>
    )
  }

  return (
    <Tooltip tipContent={`Open ${chip.path}`}>
      <button
        type="button"
        onClick={() => onFocusFile(chip.path as string)}
        className={cn(
          'inline-flex h-7 cursor-pointer items-center gap-1.5 rounded-md border bg-cool-grey-50 px-2 transition-colors dark:bg-dark-grey-700',
          'hover:bg-cool-grey-200 dark:hover:bg-dark-grey-600 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-500'
        )}
      >
        {content}
        <Icon variant="ArrowLineRightIcon" size={12} className={MUTED_TEXT} />
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
  <Expand
    id="parsed-summary"
    heading="Summary"
    toggleContent={
      <span className="flex items-center gap-2">
        <Text variant="subtext" className="font-medium">
          What the parsed config says changed
        </Text>
        <span className="flex items-center gap-x-3 text-[11px] leading-[14px]">
          <span className="text-green-800 dark:text-green-500">
            {summary.added} to create
          </span>
          <span className="text-orange-800 dark:text-orange-400">
            {summary.changed} to update
          </span>
          <span className="text-red-800 dark:text-red-500">
            {summary.removed} to delete
          </span>
        </span>
      </span>
    }
    headerClassName="px-3 py-2"
    isOpen
    className={cn('rounded-lg', SURFACE_02)}
  >
    <dl className="flex flex-col gap-2 px-3 pb-3">
      {sections.map((section) => (
        <div
          key={section.sectionKey}
          className="grid grid-cols-[8rem_1fr] items-start gap-x-3 gap-y-1 sm:grid-cols-[9rem_1fr]"
        >
          <dt className="flex items-center gap-1.5 pt-1.5">
            <Text variant="subtext" className="font-medium">
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
  </Expand>
)

// ---------------------------------------------------------------------------
// Changed-files tree — status letter lane, folder dots, muted unchanged rows
// behind a toggle.
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
    style={{ paddingLeft: `${8 + depth * 14}px` }}
    className={cn(
      'flex h-7 w-full min-w-0 items-center gap-1.5 rounded-md pr-2 text-left transition-colors',
      onClick &&
        'cursor-pointer hover:bg-cool-grey-200 dark:hover:bg-dark-grey-600',
      selected && 'bg-cool-grey-200 dark:bg-dark-grey-500',
      'focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-500'
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
      className={cn('flex flex-col gap-1 rounded-lg p-2', SURFACE_02)}
    >
      <div className="flex items-center justify-between gap-2 px-2 py-1">
        <Text variant="subtext" className={cn('font-medium', MUTED_TEXT)}>
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
                  className={MUTED_TEXT}
                />
                <Text variant="subtext" className="truncate">
                  {folder.dir.replace(/\/$/, '')}
                </Text>
                {folder.files.length ? (
                  <span
                    aria-hidden
                    className="ml-auto size-1.5 rounded-full bg-cool-grey-400 dark:bg-dark-grey-400"
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
                    variant="FileCodeIcon"
                    size={14}
                    className={cn(
                      'shrink-0',
                      file.formatOnly ? MUTED_TEXT : OP_TEXT[file.op]
                    )}
                  />
                  <MonoPath
                    path={file.path}
                    className={file.formatOnly ? MUTED_TEXT : undefined}
                  />
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
                  variant="FileCodeIcon"
                  size={14}
                  className="shrink-0 opacity-60"
                />
                <MonoPath path={path} className={MUTED_TEXT} />
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
          className={cn(
            'mt-1 flex h-7 cursor-pointer items-center gap-1.5 rounded-md px-2 text-left transition-colors',
            'hover:bg-cool-grey-200 dark:hover:bg-dark-grey-600 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-500'
          )}
        >
          <Icon
            variant={showUnchanged ? 'MinusCircleIcon' : 'PlusIcon'}
            size={12}
            className={MUTED_TEXT}
          />
          <Text variant="subtext" className={MUTED_TEXT}>
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
// Stacked per-file diff with a sticky header. Body is the house CodeBlock
// diff rendering (diffLines + CodeBlock isDiff), the same one AppConfigDiff
// uses for embedded files.
// ---------------------------------------------------------------------------

const FileDiffPanel = ({
  file,
  members,
  splitView,
  focus,
  defaultOpen,
}: {
  file: TSourceFile
  members: Record<string, string>
  splitView: boolean
  focus?: TFocus
  defaultOpen?: boolean
}) => {
  const ref = useRef<HTMLElement>(null)
  const [flash, setFlash] = useState(false)
  const [open, setOpen] = useState(defaultOpen ?? true)
  const expandAll = useExpandAll()
  const counts = useMemo(() => fileCounts(file), [file])
  const member = memberForPath(members, file.path)
  const focused = focus?.path === file.path

  useEffect(() => {
    if (expandAll?.value != null) setOpen(expandAll.value)
  }, [expandAll?.value, expandAll?.toggle])

  const nonce = focus?.nonce
  useEffect(() => {
    if (!focused) return
    setOpen(true)
    ref.current?.scrollIntoView({ block: 'start', behavior: 'smooth' })
    setFlash(true)
    const timer = window.setTimeout(() => setFlash(false), 1600)
    return () => window.clearTimeout(timer)
  }, [focused, nonce])

  return (
    <section
      ref={ref}
      aria-label={file.path}
      className={cn(
        'relative scroll-mt-4 rounded-lg border bg-cool-grey-50 transition-shadow duration-300 dark:bg-dark-grey-700',
        BORDER,
        flash && 'ring-2 ring-primary-400'
      )}
    >
      <span
        aria-hidden
        className={cn(
          'absolute inset-y-0 left-0 z-20 w-1 rounded-l-lg',
          file.formatOnly
            ? 'bg-cool-grey-300 dark:bg-dark-grey-500'
            : OP_RAIL[file.op]
        )}
      />
      <header
        className={cn(
          'sticky top-0 z-10 flex min-h-10 items-center gap-2 rounded-tr-lg border-b pl-3 pr-2 backdrop-blur-sm bg-cool-grey-50/95 dark:bg-dark-grey-700/95',
          BORDER,
          !open && 'rounded-br-lg border-b-0'
        )}
      >
        <button
          type="button"
          aria-expanded={open}
          onClick={() => setOpen((current) => !current)}
          className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md py-1.5 pr-1 text-left outline-none focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-500"
        >
          <Icon
            variant="CaretRightIcon"
            size={14}
            className={cn(
              'shrink-0 transition-transform duration-200',
              MUTED_TEXT,
              open && 'rotate-90'
            )}
          />
          <StatusLetter op={file.op} muted={file.formatOnly} />
          <MonoPath path={file.path} />
        </button>

        {member ? (
          <Badge variant="code" theme="neutral" size="sm">
            {member.kind}: {member.name}
          </Badge>
        ) : null}
        {file.formatOnly ? (
          <Tooltip tipContent="Whitespace, ordering or comments changed. The parsed config is identical.">
            <Badge size="sm" theme="neutral">
              Format only
            </Badge>
          </Tooltip>
        ) : null}
        <Counts
          added={counts.added}
          removed={counts.removed}
          muted={file.formatOnly}
        />
      </header>

      {open ? (
        <div className="p-2">
          {file.op === 'remove' ? (
            <Text
              as="p"
              variant="subtext"
              className={cn('px-1 pb-2', MUTED_TEXT)}
            >
              File removed in the head config.
            </Text>
          ) : null}
          {splitView ? (
            <div className="grid grid-cols-2 gap-2">
              <div className="overflow-hidden rounded-md border">
                <div
                  className={cn(
                    'border-b px-3 py-1 text-[11px] leading-[14px]',
                    SURFACE_02
                  )}
                >
                  Before
                </div>
                <CodeBlock language="toml" isDiff={false}>
                  {file.before ?? ''}
                </CodeBlock>
              </div>
              <div className="overflow-hidden rounded-md border">
                <div
                  className={cn(
                    'border-b px-3 py-1 text-[11px] leading-[14px]',
                    SURFACE_02
                  )}
                >
                  After
                </div>
                <CodeBlock language="toml" isDiff={false}>
                  {file.after ?? ''}
                </CodeBlock>
              </div>
            </div>
          ) : (
            <DiffCodeBlock language="toml" isDiff>
              {fileDiffText(file)}
            </DiffCodeBlock>
          )}
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
    <Badge variant="code" size="sm">
      {from}
    </Badge>
    <Icon variant="ArrowLineRightIcon" size={12} className={MUTED_TEXT} />
    <Badge variant="code" size="sm">
      {to}
    </Badge>
  </span>
)

const SourceUnavailable = ({ diff }: { diff: TSourceArchiveDiff }) => {
  const headPaths = [...new Set(Object.values(diff.members))].sort()
  return (
    <div className={cn('flex flex-col gap-3 rounded-lg px-4 py-8', SURFACE_02)}>
      <div className="flex flex-col items-center gap-2 text-center">
        <Icon variant="InfoIcon" size={20} className={MUTED_TEXT} />
        <Text as="p" variant="body" className="font-medium">
          Source diff unavailable for {diff.baselineVersion} →{' '}
          {diff.headVersion}
        </Text>
        <Text as="p" variant="subtext" className={cn('max-w-md', MUTED_TEXT)}>
          {diff.baselineVersion} was published before source capture, so there
          is no baseline to compare against. The summary above still reflects
          the parsed config. File-level diffs start with the next config
          version.
        </Text>
      </div>
      <Expand
        id="head-captured-files"
        heading={`${headPaths.length} files captured in ${diff.headVersion}`}
        headerClassName="px-2 py-1 mx-auto w-full max-w-md"
        className="mx-auto w-full max-w-md"
      >
        <ul className="flex flex-col gap-px px-2 pb-1">
          {headPaths.map((path) => (
            <li key={path} className="flex h-6 items-center gap-1.5 px-2">
              <Icon
                variant="FileCodeIcon"
                size={12}
                className="shrink-0 opacity-60"
              />
              <MonoPath path={path} className={MUTED_TEXT} />
            </li>
          ))}
        </ul>
      </Expand>
    </div>
  )
}

const NoChanges = ({ diff }: { diff: TSourceArchiveDiff }) => (
  <div className={cn('rounded-lg px-4 py-8 text-center', SURFACE_02)}>
    <Text as="p" variant="body" className="font-medium">
      No config changes
    </Text>
    <Text as="p" variant="subtext" className={MUTED_TEXT}>
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
  const [splitView, setSplitView] = useState(false)
  const [expandAll, setExpandAll] = useState<boolean | null>(null)
  const changedPaths = useMemo(
    () => new Set(sourceDiff.files.map((file) => file.path)),
    [sourceDiff.files]
  )
  const onFocusFile = (path: string) =>
    setFocus((current) => ({ path, nonce: (current?.nonce ?? 0) + 1 }))

  const hasSummary = sections.length > 0
  const hasFiles = sourceDiff.files.length > 0

  return (
    <ExpandAllContext.Provider
      value={{
        value: expandAll,
        toggle: () =>
          setExpandAll((current) => (current === true ? false : true)),
      }}
    >
      <Card
        className={cn(
          '!gap-3 !p-4 flex flex-col bg-cool-grey-50 dark:bg-dark-grey-800',
          className
        )}
      >
        <header className="flex flex-wrap items-center justify-between gap-3 px-1">
          <div className="flex items-center gap-3">
            <Text as="h2" variant="h3">
              Config changes
            </Text>
            <VersionRange
              from={sourceDiff.baselineVersion}
              to={sourceDiff.headVersion}
            />
          </div>
          {hasFiles ? (
            <div className="flex items-center gap-0.5">
              <ExpandAllToggle />
              <Button
                size="sm"
                variant="ghost"
                aria-pressed={splitView}
                aria-label={splitView ? 'Unified view' : 'Split view'}
                onClick={() => setSplitView((current) => !current)}
              >
                <Icon
                  variant={splitView ? 'ListIcon' : 'SplitHorizontalIcon'}
                  size={14}
                />
                {splitView ? 'Unified' : 'Split'}
              </Button>
            </div>
          ) : null}
        </header>

        {hasSummary ? (
          <ParsedSummary
            sections={sections}
            summary={summary}
            members={sourceDiff.members}
            changedPaths={changedPaths}
            onFocusFile={onFocusFile}
          />
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
                  splitView={splitView}
                  focus={focus}
                  defaultOpen={defaultFilesOpen && !file.formatOnly}
                />
              ))}
            </div>
          </div>
        )}
      </Card>
    </ExpandAllContext.Provider>
  )
}

const AppConfigSourceDiff = (
  props: Parameters<typeof AppConfigSourceDiffCard>[0]
) => (
  <WrapLinesProvider>
    <AppConfigSourceDiffCard {...props} />
  </WrapLinesProvider>
)

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="mx-auto max-w-6xl p-8">{children}</div>
)

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

export const Overview = () => (
  <Frame>
    <Card className="!gap-3">
      <Text as="h2" variant="h3">
        AppConfigSourceDiff — design exploration
      </Text>
      <Text as="p" variant="body">
        The app-branch config diff with authored source files as the primary
        view and the parsed diff as a collapsible summary strip.
      </Text>
      <ul className="flex list-disc flex-col gap-1 pl-5">
        <Text as="li" variant="body">
          Review a config version change the way it was authored: file by file,
          with real TOML.
        </Text>
        <Text as="li" variant="body">
          Orient with the parsed summary, then jump to the file that defines an
          entity.
        </Text>
      </ul>
      <Text as="p" variant="subtext" className={MUTED_TEXT}>
        File diffs use the house CodeBlock diff rendering (diffLines + isDiff),
        the same one the parsed AppConfigDiff card uses for embedded files.
        Every helper in this story is mock-quality and local to the file.
      </Text>
    </Card>
  </Frame>
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
  <Frame>
    <WrapLinesProvider>
      <div className="flex flex-col gap-2">
        <Text as="p" variant="subtext" className={MUTED_TEXT}>
          One changed file. The member badge links the file back to the parsed
          entity; +/- counts come from the house diff renderer.
        </Text>
        <FileDiffPanel
          file={mockSourceDiff.files[1]}
          members={mockMembers}
          splitView={false}
          defaultOpen
        />
      </div>
    </WrapLinesProvider>
  </Frame>
)

export const FileDrillDownSplit = () => (
  <Frame>
    <WrapLinesProvider>
      <FileDiffPanel
        file={mockSourceDiff.files[1]}
        members={mockMembers}
        splitView
        defaultOpen
      />
    </WrapLinesProvider>
  </Frame>
)

export const FileDrillDownVariants = () => (
  <Frame>
    <WrapLinesProvider>
      <div className="flex flex-col gap-2">
        <FileDiffPanel
          file={mockSourceDiff.files[2]}
          members={mockMembers}
          splitView={false}
          defaultOpen
        />
        <FileDiffPanel
          file={mockSourceDiff.files[3]}
          members={mockMembers}
          splitView={false}
          defaultOpen
        />
        <FileDiffPanel
          file={mockSourceDiff.files[5]}
          members={mockMembers}
          splitView={false}
          defaultOpen
        />
      </div>
    </WrapLinesProvider>
  </Frame>
)

export const CrossLink = () => (
  <Frame>
    <WrapLinesProvider>
      <AppConfigSourceDiff
        sections={mockComputedSections}
        summary={mockComputedSummary}
        sourceDiff={mockSourceDiff}
        initialFocusPath="components/whoami.toml"
        defaultFilesOpen={false}
      />
    </WrapLinesProvider>
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
      sections={mockComputedSections}
      summary={{ added: 0, removed: 0, changed: 0 }}
      sourceDiff={identicalSourceDiff}
    />
  </Frame>
)
