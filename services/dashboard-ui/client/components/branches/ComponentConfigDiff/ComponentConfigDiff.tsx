import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import {
  AppConfigDiffComponent,
  type DiffSectionData,
} from '@/components/branches/AppConfigDiff'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { CodeBlock } from '@/components/diffs/CodeBlock'
import { Diff } from '@/components/diffs/Diff'
import { DiffFileTree } from '@/components/diffs/DiffFileTree'
import { cn } from '@/utils/classnames'

type TFileChange = 'added' | 'modified' | 'removed' | 'unchanged'

export type TDiffFocus = { path: string; nonce: number }

export type TComponentSourceFile = {
  path: string
  kind: string
  change: TFileChange
  before?: string
  after?: string
}

export type IComponentConfigDiff = {
  componentName: string
  componentType: string
  previousVersion: string
  currentVersion: string
  configSections: DiffSectionData[]
  files: TComponentSourceFile[]
}

export type IAppConfigFilesDiff = {
  // Both SHAs render the prev→current version range; omit them for snapshot
  // views that show a single config as-of one run.
  previousVersion?: string
  currentVersion?: string
  configSections: DiffSectionData[]
  files: TComponentSourceFile[]
  onSelectPath?: (path: string) => void
  isLoadingFile?: boolean
  className?: string
  // Snapshot views (single config as-of one run) list entities without
  // added/modified/removed badges and hide the file panel when empty.
  snapshot?: boolean
  title?: string
  headerAction?: ReactNode
  // Matches AppConfigDiffCard: only shown as "Pending" while no sections exist.
  isPending?: boolean
  showConfigSections?: boolean
  bare?: boolean
}

const componentIcons: Record<string, TIconVariant> = {
  helm_chart: 'Helm',
  terraform_module: 'Terraform',
}

const sourceLanguage = (path: string) => {
  const name = path.toLowerCase()
  if (name.endsWith('.yaml') || name.endsWith('.yml')) return 'yaml'
  if (name.endsWith('.json')) return 'json'
  if (name.endsWith('.tf') || name.endsWith('.tfvars')) return 'hcl'
  if (name.endsWith('.toml')) return 'toml'
  if (name.endsWith('.md')) return 'markdown'
  if (name.endsWith('dockerfile') || name.endsWith('.sh')) return 'bash'
  return 'text'
}

const TREE_STATUS = {
  added: 'added',
  modified: 'modified',
  removed: 'deleted',
  unchanged: 'unchanged',
} as const

const ComponentSourceTree = ({
  files,
  onSelect,
  selectedPath,
  focus,
}: {
  files: TComponentSourceFile[]
  onSelect: (path: string) => void
  selectedPath?: string
  focus?: TDiffFocus
}) => {
  const entries = useMemo(
    () =>
      files.map(({ path, change }) => ({ path, status: TREE_STATUS[change] })),
    [files]
  )
  return (
    <DiffFileTree
      files={entries}
      onSelect={onSelect}
      selectedPath={selectedPath}
      focus={focus}
    />
  )
}

const SourceFileDiff = ({ file }: { file: TComponentSourceFile }) => {
  if (file.before === undefined && file.after === undefined) {
    return <Text theme="neutral">Unable to load file contents.</Text>
  }

  const language = sourceLanguage(file.path)
  if (file.change === 'unchanged') {
    return (
      <CodeBlock
        value={file.after ?? file.before ?? ''}
        language={language}
        filename={file.path}
        className="rounded-none border-0"
      />
    )
  }

  return (
    <Diff
      before={file.before ?? ''}
      after={file.after ?? ''}
      language={language}
      filename={file.path}
      search={false}
      className="rounded-none border-0"
    />
  )
}

export const SourceFilesPanel = ({
  files,
  focus,
  onSelectPath,
  isLoadingFile,
  showHeading = true,
  framed = true,
}: {
  files: TComponentSourceFile[]
  focus?: TDiffFocus
  onSelectPath?: (path: string) => void
  isLoadingFile?: boolean
  showHeading?: boolean
  framed?: boolean
}) => {
  const [selectedPath, setSelectedPath] = useState<string | undefined>()
  const [flash, setFlash] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)
  const changed = files.filter(({ change }) => change !== 'unchanged')
  const selected = changed.find(({ path }) => path === selectedPath)

  useEffect(() => {
    if (!changed.length) {
      if (selectedPath) setSelectedPath(undefined)
      return
    }
    if (!changed.some(({ path }) => path === selectedPath)) {
      setSelectedPath(changed[0].path)
    }
  }, [changed, selectedPath])

  useEffect(() => {
    if (selectedPath) onSelectPath?.(selectedPath)
  }, [onSelectPath, selectedPath])

  useEffect(() => {
    if (!focus) return
    setSelectedPath(focus.path)
    setFlash(true)
    panelRef.current?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    const timer = setTimeout(() => setFlash(false), 900)
    return () => clearTimeout(timer)
  }, [focus])

  if (!changed.length) {
    return (
      <Text theme="neutral">
        {files.length
          ? 'No files changed.'
          : 'This component references no files.'}
      </Text>
    )
  }

  return (
    <div className="@container flex flex-col gap-3">
      {showHeading ? (
        <div className="flex flex-wrap items-center justify-between gap-2">
          <Text variant="label" theme="neutral">
            Referenced files
          </Text>
          <div className="flex items-center gap-1.5">
            <Badge theme="success">
              {changed.filter(({ change }) => change === 'added').length} added
            </Badge>
            <Badge theme="warn">
              {changed.filter(({ change }) => change === 'modified').length}{' '}
              modified
            </Badge>
            <Badge theme="error">
              {changed.filter(({ change }) => change === 'removed').length}{' '}
              removed
            </Badge>
          </div>
        </div>
      ) : null}
      <div
        ref={panelRef}
        className={cn(
          'flex max-h-[70vh] flex-col overflow-hidden @3xl:h-144 @3xl:flex-row',
          framed && 'border rounded-md'
        )}
      >
        <div className="h-72 overflow-auto border-b bg-code pt-4 @3xl:h-full @3xl:w-2/5 @3xl:border-b-0 @3xl:border-r">
          <ComponentSourceTree
            files={changed}
            onSelect={setSelectedPath}
            selectedPath={selectedPath}
            focus={focus}
          />
        </div>
        <div
          className={cn(
            'max-h-[36rem] min-w-0 overflow-auto overscroll-contain transition-colors duration-500 @3xl:h-full @3xl:max-h-none @3xl:w-3/5',
            (!selected || isLoadingFile) && 'p-4',
            flash && 'bg-primary-50 dark:bg-primary-900/20'
          )}
        >
          {selected ? (
            isLoadingFile ? (
              <Skeleton lines={8} height="0.75rem" width="90%" />
            ) : (
              <SourceFileDiff file={selected} />
            )
          ) : (
            <Text theme="neutral">Select a file to inspect it.</Text>
          )}
        </div>
      </div>
    </div>
  )
}

const VersionRange = ({
  previous,
  current,
}: {
  previous: string
  current: string
}) => (
  <div className="flex min-w-0 flex-wrap items-center gap-1.5">
    <Badge variant="code">{previous}</Badge>
    <Icon variant="ArrowRightIcon" size={12} />
    <Badge variant="code">{current}</Badge>
  </div>
)

export const ComponentConfigDiff = ({
  componentName,
  componentType,
  previousVersion,
  currentVersion,
  configSections,
  files,
}: IComponentConfigDiff) => {
  return (
    <Card className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 min-w-0">
          <Icon variant={componentIcons[componentType] ?? 'CubeIcon'} />
          <Text weight="strong" className="truncate">
            {componentName}
          </Text>
          <Badge variant="code">{componentType}</Badge>
        </div>
        <VersionRange previous={previousVersion} current={currentVersion} />
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="label" theme="neutral">
          Configuration
        </Text>
        <AppConfigDiffComponent
          sections={configSections}
          summary={null}
          defaultSectionsOpen
        />
      </div>

      <SourceFilesPanel files={files} />
    </Card>
  )
}

const sectionIcons: Record<string, TIconVariant> = {
  components: 'CubeIcon',
  actions: 'LightningIcon',
  runbooks: 'BookOpenIcon',
  inputs: 'ListBulletsIcon',
  secrets: 'KeyIcon',
  sandbox: 'TerminalWindowIcon',
  runner: 'GearIcon',
  permissions: 'ShieldIcon',
  policies: 'ShieldCheckIcon',
  stack: 'StackIcon',
}

const opTheme = { add: 'success', remove: 'error', change: 'info' } as const

const opLabel = { add: 'added', remove: 'removed', change: 'modified' } as const

const EntityRow = ({
  name,
  subtext,
  op,
  targetPath,
  onFocusFile,
}: {
  name: string
  subtext?: string
  op?: 'add' | 'remove' | 'change'
  targetPath?: string
  onFocusFile: (path: string) => void
}) => {
  const hint = targetPath ? subtext : (subtext ?? 'config only')
  const row = (
    <>
      <span className="flex min-w-0 flex-col">
        <Text className="truncate">{name}</Text>
        {hint ? (
          <Text variant="subtext" theme="neutral">
            {hint}
          </Text>
        ) : null}
      </span>
      {op && (
        <Badge theme={opTheme[op]} size="sm">
          {opLabel[op]}
        </Badge>
      )}
    </>
  )
  if (!targetPath) {
    return (
      <div className="flex items-center justify-between gap-2 px-2 py-1.5">
        {row}
      </div>
    )
  }
  return (
    <Button
      variant="ghost"
      onClick={() => onFocusFile(targetPath)}
      className="!flex !h-auto !w-full justify-between !gap-2 !rounded-md !px-2 !py-1.5 text-left !font-normal !whitespace-normal"
    >
      {row}
    </Button>
  )
}

const SectionSummary = ({
  section,
  onFocusFile,
  snapshot = false,
}: {
  section: DiffSectionData
  onFocusFile: (path: string) => void
  // Snapshot views show a config as-of one run: entities exist, they were not
  // added/modified/removed, so op badges and change counts are noise.
  snapshot?: boolean
}) => {
  const summary = (
    <div className="flex items-center justify-between gap-2 px-2 pb-1">
      <span className="flex items-center gap-1.5">
        <Icon
          variant={sectionIcons[section.sectionKey] ?? 'CubeIcon'}
          size={14}
        />
        <Text variant="label" theme="neutral">
          {section.name}
        </Text>
      </span>
      {!snapshot && (
        <ChangeCountSummary
          added={section.additions}
          updated={section.changed}
          removed={section.removals}
        />
      )}
    </div>
  )
  const entityRows = section.grouped ? (
    section.entities.map((entity) => (
      <EntityRow
        key={entity.name}
        name={entity.name}
        subtext={entity.componentType?.replace(/_/g, ' ')}
        op={snapshot ? undefined : entity.op}
        targetPath={entity.files?.[0]?.name}
        onFocusFile={onFocusFile}
      />
    ))
  ) : (
    <EntityRow
      name={section.name}
      op={
        snapshot
          ? undefined
          : section.additions
            ? 'add'
            : section.removals
              ? 'remove'
              : 'change'
      }
      targetPath={section.files?.[0]?.name}
      onFocusFile={onFocusFile}
    />
  )
  return (
    <div className="mb-3 flex flex-col break-inside-avoid rounded-lg border px-3 py-3">
      {summary}
      <div className="flex flex-col">{entityRows}</div>
    </div>
  )
}

export const SourceFilesSkeleton = () => (
  <div className="flex h-144 max-h-[70vh] flex-col @3xl:flex-row">
    <div className="flex flex-col gap-3 border-b bg-code p-4 @3xl:w-2/5 @3xl:border-b-0 @3xl:border-r">
      <Skeleton height="2rem" />
      <Skeleton
        lines={8}
        height="0.75rem"
        width={['72%', '56%', '80%', '44%', '68%', '52%', '76%', '48%']}
      />
    </div>
    <div className="flex min-w-0 flex-1 flex-col gap-3 p-4">
      <Skeleton height="1rem" width="16rem" />
      <Skeleton lines={10} height="0.75rem" width="92%" />
    </div>
  </div>
)

export const AppConfigFilesDiff = ({
  previousVersion,
  currentVersion,
  configSections,
  files,
  onSelectPath,
  isLoadingFile,
  className,
  snapshot = false,
  title = 'App configuration',
  headerAction,
  isPending = false,
  showConfigSections = true,
  bare = false,
}: IAppConfigFilesDiff) => {
  const [focus, setFocus] = useState<TDiffFocus | undefined>()
  const focusFile = (path: string) =>
    setFocus((current) => ({ path, nonce: (current?.nonce ?? 0) + 1 }))
  const showPending = isPending && configSections.length === 0
  const Shell = bare ? 'div' : Card

  return (
    <Shell className={cn('flex flex-col gap-4', className)}>
      {bare ? null : (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Text weight="strong">{title}</Text>
          {showPending ? (
            <Text variant="subtext" theme="neutral">
              Pending
            </Text>
          ) : (
            previousVersion &&
            currentVersion && (
              <VersionRange
                previous={previousVersion}
                current={currentVersion}
              />
            )
          )}
          {headerAction}
        </div>
      )}

      {showPending ? (
        <div className="px-4 py-3 text-center">
          <EmptyState
            emptyTitle="Changes pending"
            emptyMessage="Changes appear after the app config builds."
            variant="diagram"
            size="sm"
          />
        </div>
      ) : (
        <>
          {showConfigSections ? (
            <div className="gap-3 sm:columns-2">
              {configSections.map((section) => (
                <SectionSummary
                  key={section.sectionKey}
                  section={section}
                  onFocusFile={focusFile}
                  snapshot={snapshot}
                />
              ))}
            </div>
          ) : null}

          {(files.length > 0 || !snapshot) && (
            <SourceFilesPanel
              files={files}
              focus={focus}
              onSelectPath={onSelectPath}
              isLoadingFile={isLoadingFile}
              showHeading={!bare}
              framed={!bare}
            />
          )}
        </>
      )}
    </Shell>
  )
}
