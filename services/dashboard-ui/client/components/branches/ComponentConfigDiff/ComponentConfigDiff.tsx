import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { File as DiffFile, MultiFileDiff } from '@pierre/diffs/react'
import { FileTree, useFileTree } from '@pierre/trees/react'
import type { GitStatusEntry } from '@pierre/trees'
import { Badge, type TBadgeTheme } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import {
  AppConfigDiffComponent,
  type DiffSectionData,
} from '@/components/branches/AppConfigDiff'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
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
}

const changeTheme: Record<TFileChange, TBadgeTheme> = {
  added: 'success',
  modified: 'warn',
  removed: 'error',
  unchanged: 'neutral',
}

const componentIcons: Record<string, TIconVariant> = {
  helm_chart: 'Helm',
  terraform_module: 'Terraform',
}

const diffOptions = {
  diffStyle: 'unified',
  expandUnchanged: false,
  lineDiffType: 'word',
  overflow: 'scroll',
  themeType: 'system',
} as const

const fileOptions = {
  overflow: 'scroll',
  themeType: 'system',
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
  const filesRef = useRef(files)
  filesRef.current = files
  const treeSelectedPath = useRef<string | undefined>(undefined)
  const resetKeyRef = useRef<string | undefined>(undefined)
  const paths = useMemo(() => files.map(({ path }) => path), [files])
  const gitStatus = useMemo<GitStatusEntry[]>(
    () =>
      files.flatMap(({ change, path }) =>
        change === 'unchanged'
          ? []
          : [
              {
                path,
                status:
                  change === 'removed'
                    ? 'deleted'
                    : change === 'added'
                      ? 'added'
                      : 'modified',
              },
            ]
      ),
    [files]
  )
  const { model } = useFileTree({
    fileTreeSearchMode: 'expand-matches',
    flattenEmptyDirectories: true,
    gitStatus,
    icons: 'standard',
    initialExpansion: 'open',
    initialSelectedPaths: selectedPath ? [selectedPath] : [],
    onSelectionChange: (selectedPaths) => {
      const path = selectedPaths.at(-1)
      treeSelectedPath.current = path
      if (path && filesRef.current.some((file) => file.path === path)) {
        onSelect(path)
      }
    },
    paths,
    search: true,
  })

  // `files` changes identity whenever lazy content fetches resolve, but the
  // path set usually stays the same. resetPaths rebuilds the store with
  // initialExpansion applied, so only reset when paths or statuses change.
  const pathsKey = paths.join('\n')
  const gitStatusKey = gitStatus
    .map(({ path, status }) => `${path}:${status}`)
    .join('\n')
  const pathsRef = useRef(paths)
  pathsRef.current = paths
  const gitStatusRef = useRef(gitStatus)
  gitStatusRef.current = gitStatus

  useEffect(() => {
    if (resetKeyRef.current === `${pathsKey}\n${gitStatusKey}`) return
    resetKeyRef.current = `${pathsKey}\n${gitStatusKey}`
    model.resetPaths(pathsRef.current)
    model.setGitStatus(gitStatusRef.current)
  }, [gitStatusKey, model, pathsKey])

  // Sync a parent-driven selection (initial default or focus target) into the
  // tree. Tree-driven selections already live in the model; scrolling for them
  // would fight the user's own scrolling and expansion.
  useEffect(() => {
    if (!selectedPath || selectedPath === treeSelectedPath.current) return
    model.getItem(selectedPath)?.select()
  }, [model, selectedPath])

  useEffect(() => {
    if (!focus) return
    if (!filesRef.current.some((file) => file.path === focus.path)) return
    // Expand ancestors so scrollToPath can see the target even when its
    // folder is collapsed (canonical directory paths end with '/').
    const segments = focus.path.split('/')
    for (let i = 1; i < segments.length; i++) {
      const dir = model.getItem(`${segments.slice(0, i).join('/')}/`)
      if (dir != null && 'expand' in dir) dir.expand()
    }
    model.getItem(focus.path)?.select()
    model.scrollToPath(focus.path, { offset: 'nearest' })
  }, [focus, model])

  return (
    <FileTree
      model={model}
      // The tree overwrites its host's class attribute, so size it via style.
      style={
        {
          height: files.length > 5 ? '36rem' : '24rem',
          '--trees-bg-override': 'var(--background)',
          '--trees-border-color-override': 'var(--border-color)',
          '--trees-fg-override': 'var(--foreground)',
          '--trees-selected-bg-override': 'var(--background-neutral)',
        } as React.CSSProperties
      }
    />
  )
}

const SourceFileDiff = ({ file }: { file: TComponentSourceFile }) => {
  const lang =
    file.path.endsWith('.yaml') || file.path.endsWith('.yml')
      ? 'yaml'
      : undefined

  if (file.change === 'unchanged') {
    return (
      <div className="flex flex-col gap-2">
        <Badge theme="neutral">{file.kind}</Badge>
        <DiffFile
          key={file.path}
          file={{
            cacheKey: file.path,
            contents: file.after ?? '',
            lang,
            name: file.path,
          }}
          options={fileOptions}
        />
      </div>
    )
  }

  // A failed or still-undelivered content fetch leaves both sides undefined;
  // MultiFileDiff throws when neither side is passed.
  if (file.before === undefined && file.after === undefined) {
    return (
      <div className="flex flex-col gap-2">
        <Badge theme={changeTheme[file.change]}>{file.change}</Badge>
        <Text theme="neutral">Unable to load file contents.</Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <Badge theme={changeTheme[file.change]}>{file.change}</Badge>
        <Badge variant="code">{file.kind}</Badge>
      </div>
      <MultiFileDiff
        key={`${file.path}:${file.before?.length ?? 0}:${file.after?.length ?? 0}`}
        newFile={
          file.after !== undefined
            ? {
                cacheKey: file.path,
                contents: file.after,
                lang,
                name: file.path,
              }
            : undefined
        }
        oldFile={
          file.before !== undefined
            ? {
                cacheKey: file.path,
                contents: file.before,
                lang,
                name: file.path,
              }
            : undefined
        }
        options={diffOptions}
      />
    </div>
  )
}

export const SourceFilesPanel = ({
  files,
  focus,
  onSelectPath,
  isLoadingFile,
}: {
  files: TComponentSourceFile[]
  focus?: TDiffFocus
  onSelectPath?: (path: string) => void
  isLoadingFile?: boolean
}) => {
  const [selectedPath, setSelectedPath] = useState<string | undefined>()
  const [flash, setFlash] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)
  const selected = files.find(({ path }) => path === selectedPath)
  const changed = files.filter(({ change }) => change !== 'unchanged')

  useEffect(() => {
    if (!changed.length) return
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

  return (
    <div className="@container flex flex-col gap-3">
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
      {changed.length ? (
        <div
          ref={panelRef}
          className={cn(
            'flex flex-col overflow-hidden border rounded-md @3xl:flex-row',
            changed.length > 5 ? 'min-h-144' : 'min-h-96'
          )}
        >
          <div className="border-b p-3 @3xl:w-2/5 @3xl:border-b-0 @3xl:border-r">
            <ComponentSourceTree
              files={changed}
              onSelect={setSelectedPath}
              selectedPath={selectedPath}
              focus={focus}
            />
          </div>
          <div
            className={cn(
              'min-w-0 overflow-auto p-4 transition-colors duration-500 @3xl:w-3/5',
              flash && 'bg-primary-50 dark:bg-primary-900/20'
            )}
          >
            {selected ? (
              isLoadingFile ? (
                <Loading />
              ) : (
                <SourceFileDiff file={selected} />
              )
            ) : (
              <Text theme="neutral">Select a file to inspect it.</Text>
            )}
          </div>
        </div>
      ) : (
        <Text theme="neutral">
          {files.length
            ? 'No changed files in this component.'
            : 'This component references no files.'}
        </Text>
      )}
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
}: IAppConfigFilesDiff) => {
  const [focus, setFocus] = useState<TDiffFocus | undefined>()
  const focusFile = (path: string) =>
    setFocus((current) => ({ path, nonce: (current?.nonce ?? 0) + 1 }))
  const showPending = isPending && configSections.length === 0

  return (
    <Card className={cn('flex flex-col gap-4', className)}>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Text weight="strong">{title}</Text>
        {showPending ? (
          <Text variant="subtext" theme="neutral">
            Pending
          </Text>
        ) : (
          previousVersion &&
          currentVersion && (
            <VersionRange previous={previousVersion} current={currentVersion} />
          )
        )}
        {headerAction}
      </div>

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

          {(files.length > 0 || !snapshot) && (
            <SourceFilesPanel
              files={files}
              focus={focus}
              onSelectPath={onSelectPath}
              isLoadingFile={isLoadingFile}
            />
          )}
        </>
      )}
    </Card>
  )
}
